// Package postgres implements storage.Store on top of PostgreSQL via the pgx
// stdlib driver. It mirrors the SQLite backend but uses $N placeholders and a
// SELECT ... FOR UPDATE SKIP LOCKED claim so the queue scales to many workers
// and instances.
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/edaywalid/whook/internal/core"
	"github.com/edaywalid/whook/internal/migrate"
	"github.com/edaywalid/whook/internal/storage"
)

// Store is the Postgres-backed storage.Store implementation.
type Store struct {
	db *sql.DB
}

// Open connects to the Postgres database at dsn, applies migrations, and returns
// a ready Store.
func Open(ctx context.Context, dsn string) (*Store, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening postgres: %w", err)
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}
	if err := migrate.Run(db, migrate.Postgres); err != nil {
		return nil, fmt.Errorf("running migrations: %w", err)
	}
	return &Store{db: db}, nil
}

// Close closes the underlying database.
func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) CreateSource(ctx context.Context, src core.Source, secret []byte) error {
	const q = `INSERT INTO sources (id, name, verifier, signing_secret_encrypted, dedup_header, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := s.db.ExecContext(ctx, q,
		src.ID, src.Name, nullStr(src.Verifier), secret, nullStr(src.DedupHeader), src.CreatedAt)
	if err != nil {
		return fmt.Errorf("creating source %q: %w", src.Name, err)
	}
	return nil
}

func (s *Store) UpdateSource(ctx context.Context, src core.Source, secret []byte) error {
	const q = `UPDATE sources
		SET verifier = $1, signing_secret_encrypted = $2, dedup_header = $3
		WHERE name = $4`
	res, err := s.db.ExecContext(ctx, q, nullStr(src.Verifier), secret, nullStr(src.DedupHeader), src.Name)
	if err != nil {
		return fmt.Errorf("updating source %q: %w", src.Name, err)
	}
	return requireRow(res, "source")
}

func (s *Store) GetSource(ctx context.Context, name string) (core.Source, error) {
	const q = `SELECT id, name, verifier, signing_secret_encrypted IS NOT NULL, dedup_header, created_at
		FROM sources WHERE name = $1`
	var (
		src      core.Source
		verifier sql.NullString
		dedup    sql.NullString
	)
	err := s.db.QueryRowContext(ctx, q, name).
		Scan(&src.ID, &src.Name, &verifier, &src.HasSecret, &dedup, &src.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return core.Source{}, storage.ErrNotFound
	}
	if err != nil {
		return core.Source{}, fmt.Errorf("getting source %q: %w", name, err)
	}
	src.Verifier = verifier.String
	src.DedupHeader = dedup.String
	return src, nil
}

func (s *Store) ListSources(ctx context.Context) ([]core.Source, error) {
	const q = `SELECT id, name, verifier, signing_secret_encrypted IS NOT NULL, dedup_header, created_at
		FROM sources ORDER BY name`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("listing sources: %w", err)
	}
	defer rows.Close()

	out := []core.Source{}
	for rows.Next() {
		var (
			src      core.Source
			verifier sql.NullString
			dedup    sql.NullString
		)
		if err := rows.Scan(&src.ID, &src.Name, &verifier, &src.HasSecret, &dedup, &src.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning source: %w", err)
		}
		src.Verifier = verifier.String
		src.DedupHeader = dedup.String
		out = append(out, src)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating sources: %w", err)
	}
	return out, nil
}

func (s *Store) SourceSecret(ctx context.Context, name string) ([]byte, error) {
	const q = `SELECT signing_secret_encrypted FROM sources WHERE name = $1`
	var secret []byte
	err := s.db.QueryRowContext(ctx, q, name).Scan(&secret)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, storage.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("getting secret for source %q: %w", name, err)
	}
	return secret, nil
}

func (s *Store) SaveEvent(ctx context.Context, e core.Event) (string, bool, error) {
	headers, err := json.Marshal(e.Headers)
	if err != nil {
		return "", false, fmt.Errorf("marshaling headers: %w", err)
	}

	if e.DedupKey != "" {
		if id, found, err := s.existingEvent(ctx, e.Source, e.DedupKey); err != nil {
			return "", false, err
		} else if found {
			return id, false, nil
		}
	}

	const q = `INSERT INTO events
		(id, source, received_at, source_ip, method, path, query, headers, body, dedup_key, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err = s.db.ExecContext(ctx, q,
		e.ID, e.Source, e.ReceivedAt, nullStr(e.SourceIP), e.Method, e.Path,
		nullStr(e.Query), string(headers), e.Body, nullStr(e.DedupKey), string(e.Status))
	if err != nil {
		// Lost the race on the unique (source, dedup_key) index. Return the
		// event that won.
		if e.DedupKey != "" {
			if id, found, lookupErr := s.existingEvent(ctx, e.Source, e.DedupKey); lookupErr == nil && found {
				return id, false, nil
			}
		}
		return "", false, fmt.Errorf("saving event: %w", err)
	}
	return e.ID, true, nil
}

func (s *Store) existingEvent(ctx context.Context, source, dedupKey string) (string, bool, error) {
	const q = `SELECT id FROM events WHERE source = $1 AND dedup_key = $2`
	var id string
	err := s.db.QueryRowContext(ctx, q, source, dedupKey).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("looking up dedup key: %w", err)
	}
	return id, true, nil
}

func (s *Store) GetEvent(ctx context.Context, id string) (core.Event, error) {
	const q = `SELECT id, source, received_at, source_ip, method, path, query, headers, body, dedup_key, status
		FROM events WHERE id = $1`
	e, err := scanEvent(s.db.QueryRowContext(ctx, q, id))
	if errors.Is(err, sql.ErrNoRows) {
		return core.Event{}, storage.ErrNotFound
	}
	if err != nil {
		return core.Event{}, fmt.Errorf("getting event %q: %w", id, err)
	}
	return e, nil
}

func (s *Store) ListEvents(ctx context.Context, f storage.EventFilter) ([]core.Event, error) {
	const q = `SELECT id, source, received_at, source_ip, method, path, query, headers, body, dedup_key, status
		FROM events
		WHERE ($1 = '' OR source = $1) AND ($2 = '' OR status = $2)
		ORDER BY received_at DESC
		LIMIT $3 OFFSET $4`
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, q, f.Source, string(f.Status), limit, f.Offset)
	if err != nil {
		return nil, fmt.Errorf("listing events: %w", err)
	}
	defer rows.Close()

	out := []core.Event{}
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning event: %w", err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating events: %w", err)
	}
	return out, nil
}

func (s *Store) CreateDestination(ctx context.Context, d core.Destination) error {
	filter, backoff, err := marshalDestination(d)
	if err != nil {
		return err
	}
	const q = `INSERT INTO destinations (id, source, url, enabled, backoff_config, filter_spec, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	if _, err := s.db.ExecContext(ctx, q, d.ID, d.Source, d.URL, d.Enabled, backoff, filter, d.CreatedAt); err != nil {
		return fmt.Errorf("creating destination: %w", err)
	}
	return nil
}

func (s *Store) UpdateDestination(ctx context.Context, d core.Destination) error {
	filter, backoff, err := marshalDestination(d)
	if err != nil {
		return err
	}
	const q = `UPDATE destinations SET url = $1, enabled = $2, backoff_config = $3, filter_spec = $4 WHERE id = $5`
	res, err := s.db.ExecContext(ctx, q, d.URL, d.Enabled, backoff, filter, d.ID)
	if err != nil {
		return fmt.Errorf("updating destination %q: %w", d.ID, err)
	}
	return requireRow(res, "destination")
}

func (s *Store) GetDestination(ctx context.Context, id string) (core.Destination, error) {
	const q = `SELECT id, source, url, enabled, backoff_config, filter_spec, created_at
		FROM destinations WHERE id = $1`
	d, err := scanDestination(s.db.QueryRowContext(ctx, q, id))
	if errors.Is(err, sql.ErrNoRows) {
		return core.Destination{}, storage.ErrNotFound
	}
	if err != nil {
		return core.Destination{}, fmt.Errorf("getting destination %q: %w", id, err)
	}
	return d, nil
}

func (s *Store) ListDestinations(ctx context.Context, source string) ([]core.Destination, error) {
	const q = `SELECT id, source, url, enabled, backoff_config, filter_spec, created_at
		FROM destinations WHERE source = $1 ORDER BY created_at`
	rows, err := s.db.QueryContext(ctx, q, source)
	if err != nil {
		return nil, fmt.Errorf("listing destinations: %w", err)
	}
	defer rows.Close()

	out := []core.Destination{}
	for rows.Next() {
		d, err := scanDestination(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning destination: %w", err)
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating destinations: %w", err)
	}
	return out, nil
}

func (s *Store) RecordAttempt(ctx context.Context, a core.DeliveryAttempt) error {
	const q = `INSERT INTO delivery_attempts
		(id, event_id, destination_id, attempt_number, status_code, response_snippet, duration_ms, outcome, is_replay, attempted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	var status any
	if a.StatusCode != 0 {
		status = a.StatusCode
	}
	_, err := s.db.ExecContext(ctx, q,
		a.ID, a.EventID, a.DestinationID, a.AttemptNumber, status,
		a.ResponseSnippet, a.DurationMS, string(a.Outcome), a.IsReplay, a.AttemptedAt)
	if err != nil {
		return fmt.Errorf("recording attempt: %w", err)
	}
	return nil
}

func (s *Store) ListAttempts(ctx context.Context, eventID string) ([]core.DeliveryAttempt, error) {
	const q = `SELECT id, event_id, destination_id, attempt_number, status_code, response_snippet, duration_ms, outcome, is_replay, attempted_at
		FROM delivery_attempts WHERE event_id = $1 ORDER BY attempted_at`
	rows, err := s.db.QueryContext(ctx, q, eventID)
	if err != nil {
		return nil, fmt.Errorf("listing attempts: %w", err)
	}
	defer rows.Close()

	out := []core.DeliveryAttempt{}
	for rows.Next() {
		var (
			a      core.DeliveryAttempt
			status sql.NullInt64
		)
		if err := rows.Scan(&a.ID, &a.EventID, &a.DestinationID, &a.AttemptNumber, &status,
			&a.ResponseSnippet, &a.DurationMS, &a.Outcome, &a.IsReplay, &a.AttemptedAt); err != nil {
			return nil, fmt.Errorf("scanning attempt: %w", err)
		}
		a.StatusCode = int(status.Int64)
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating attempts: %w", err)
	}
	return out, nil
}

func (s *Store) Enqueue(ctx context.Context, p core.PendingDelivery) error {
	const q = `INSERT INTO pending_deliveries
		(id, event_id, destination_id, attempt_number, next_attempt_at, state, is_replay, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	now := p.NextAttemptAt
	_, err := s.db.ExecContext(ctx, q,
		p.ID, p.EventID, p.DestinationID, p.AttemptNumber, p.NextAttemptAt, string(core.StatePending), p.IsReplay, now, now)
	if err != nil {
		return fmt.Errorf("enqueueing delivery: %w", err)
	}
	return nil
}

func (s *Store) ClaimDue(ctx context.Context, now time.Time, lease time.Duration, limit int) ([]core.PendingDelivery, error) {
	const q = `WITH due AS (
			SELECT id FROM pending_deliveries
			WHERE state = 'pending'
			  AND next_attempt_at <= $1
			  AND (locked_until IS NULL OR locked_until < $1)
			ORDER BY next_attempt_at
			LIMIT $2
			FOR UPDATE SKIP LOCKED
		)
		UPDATE pending_deliveries p
		SET locked_until = $3, updated_at = $4
		FROM due
		WHERE p.id = due.id
		RETURNING p.id, p.event_id, p.destination_id, p.attempt_number, p.next_attempt_at, p.state, p.is_replay`
	rows, err := s.db.QueryContext(ctx, q, now, limit, now.Add(lease), now)
	if err != nil {
		return nil, fmt.Errorf("claiming due deliveries: %w", err)
	}
	defer rows.Close()

	out := []core.PendingDelivery{}
	for rows.Next() {
		p, err := scanPending(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning pending delivery: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating pending deliveries: %w", err)
	}
	return out, nil
}

func (s *Store) MarkDelivered(ctx context.Context, id string) error {
	return s.setState(ctx, id, core.StateDelivered)
}

func (s *Store) MarkDeadLetter(ctx context.Context, id string) error {
	return s.setState(ctx, id, core.StateDeadLetter)
}

func (s *Store) setState(ctx context.Context, id string, state core.DeliveryState) error {
	const q = `UPDATE pending_deliveries SET state = $1, locked_until = NULL, updated_at = $2 WHERE id = $3`
	res, err := s.db.ExecContext(ctx, q, string(state), now(), id)
	if err != nil {
		return fmt.Errorf("setting delivery state: %w", err)
	}
	return requireRow(res, "pending delivery")
}

func (s *Store) Reschedule(ctx context.Context, id string, attempt int, next time.Time) error {
	const q = `UPDATE pending_deliveries
		SET attempt_number = $1, next_attempt_at = $2, locked_until = NULL, updated_at = $3
		WHERE id = $4`
	res, err := s.db.ExecContext(ctx, q, attempt, next, now(), id)
	if err != nil {
		return fmt.Errorf("rescheduling delivery %q: %w", id, err)
	}
	return requireRow(res, "pending delivery")
}

func now() time.Time {
	return time.Now().UTC()
}
