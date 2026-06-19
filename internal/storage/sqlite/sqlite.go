// Package sqlite implements storage.Store on top of SQLite using the pure-Go
// modernc driver, so the gateway needs no cgo and ships as one binary.
package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/edaywalid/whook/internal/core"
	"github.com/edaywalid/whook/internal/migrate"
	"github.com/edaywalid/whook/internal/storage"
)

// Store is the SQLite-backed storage.Store implementation.
type Store struct {
	db *sql.DB
}

// Open connects to the SQLite database at dsn, applies migrations, and returns
// a ready Store. The pool is capped at one connection because SQLite allows a
// single writer; this trades throughput for freedom from "database is locked".
func Open(ctx context.Context, dsn string) (*Store, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("pinging sqlite: %w", err)
	}
	if err := migrate.Run(db); err != nil {
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
		VALUES (?, ?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, q,
		src.ID, src.Name, nullStr(src.Verifier), secret, nullStr(src.DedupHeader), src.CreatedAt)
	if err != nil {
		return fmt.Errorf("creating source %q: %w", src.Name, err)
	}
	return nil
}

func (s *Store) UpdateSource(ctx context.Context, src core.Source, secret []byte) error {
	const q = `UPDATE sources
		SET verifier = ?, signing_secret_encrypted = ?, dedup_header = ?
		WHERE name = ?`
	res, err := s.db.ExecContext(ctx, q, nullStr(src.Verifier), secret, nullStr(src.DedupHeader), src.Name)
	if err != nil {
		return fmt.Errorf("updating source %q: %w", src.Name, err)
	}
	return requireRow(res, "source")
}

func (s *Store) GetSource(ctx context.Context, name string) (core.Source, error) {
	const q = `SELECT id, name, verifier, signing_secret_encrypted IS NOT NULL, dedup_header, created_at
		FROM sources WHERE name = ?`
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
	const q = `SELECT signing_secret_encrypted FROM sources WHERE name = ?`
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
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = s.db.ExecContext(ctx, q,
		e.ID, e.Source, e.ReceivedAt, nullStr(e.SourceIP), e.Method, e.Path,
		nullStr(e.Query), string(headers), e.Body, nullStr(e.DedupKey), string(e.Status))
	if err != nil {
		// Lost the race on the unique (source, dedup_key) index: another
		// request inserted the same event first. Return that one.
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
	const q = `SELECT id FROM events WHERE source = ? AND dedup_key = ?`
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
		FROM events WHERE id = ?`
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
		WHERE (? = '' OR source = ?) AND (? = '' OR status = ?)
		ORDER BY received_at DESC
		LIMIT ? OFFSET ?`
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, q, f.Source, f.Source, string(f.Status), string(f.Status), limit, f.Offset)
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
		VALUES (?, ?, ?, ?, ?, ?, ?)`
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
	const q = `UPDATE destinations SET url = ?, enabled = ?, backoff_config = ?, filter_spec = ? WHERE id = ?`
	res, err := s.db.ExecContext(ctx, q, d.URL, d.Enabled, backoff, filter, d.ID)
	if err != nil {
		return fmt.Errorf("updating destination %q: %w", d.ID, err)
	}
	return requireRow(res, "destination")
}

func (s *Store) GetDestination(ctx context.Context, id string) (core.Destination, error) {
	const q = `SELECT id, source, url, enabled, backoff_config, filter_spec, created_at
		FROM destinations WHERE id = ?`
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
		FROM destinations WHERE source = ? ORDER BY created_at`
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
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
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
		FROM delivery_attempts WHERE event_id = ? ORDER BY attempted_at`
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
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := p.NextAttemptAt
	_, err := s.db.ExecContext(ctx, q,
		p.ID, p.EventID, p.DestinationID, p.AttemptNumber, p.NextAttemptAt, string(core.StatePending), p.IsReplay, now, now)
	if err != nil {
		return fmt.Errorf("enqueueing delivery: %w", err)
	}
	return nil
}

func (s *Store) ClaimDue(ctx context.Context, now time.Time, lease time.Duration, limit int) ([]core.PendingDelivery, error) {
	// SQLite has no SKIP LOCKED, but the single-writer pool plus an atomic
	// UPDATE ... RETURNING means two workers can never claim the same row.
	const q = `UPDATE pending_deliveries
		SET locked_until = ?, updated_at = ?
		WHERE id IN (
			SELECT id FROM pending_deliveries
			WHERE state = 'pending'
			  AND next_attempt_at <= ?
			  AND (locked_until IS NULL OR locked_until < ?)
			ORDER BY next_attempt_at
			LIMIT ?
		)
		RETURNING id, event_id, destination_id, attempt_number, next_attempt_at, state, is_replay`
	leaseUntil := now.Add(lease)
	rows, err := s.db.QueryContext(ctx, q, leaseUntil, now, now, now, limit)
	if err != nil {
		return nil, fmt.Errorf("claiming due deliveries: %w", err)
	}
	defer rows.Close()

	out := []core.PendingDelivery{}
	for rows.Next() {
		var p core.PendingDelivery
		if err := rows.Scan(&p.ID, &p.EventID, &p.DestinationID, &p.AttemptNumber, &p.NextAttemptAt, &p.State, &p.IsReplay); err != nil {
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
	const q = `UPDATE pending_deliveries SET state = ?, locked_until = NULL, updated_at = ? WHERE id = ?`
	res, err := s.db.ExecContext(ctx, q, string(state), s.now(), id)
	if err != nil {
		return fmt.Errorf("setting delivery state: %w", err)
	}
	return requireRow(res, "pending delivery")
}

func (s *Store) Reschedule(ctx context.Context, id string, attempt int, next time.Time) error {
	const q = `UPDATE pending_deliveries
		SET attempt_number = ?, next_attempt_at = ?, locked_until = NULL, updated_at = ?
		WHERE id = ?`
	res, err := s.db.ExecContext(ctx, q, attempt, next, s.now(), id)
	if err != nil {
		return fmt.Errorf("rescheduling delivery %q: %w", id, err)
	}
	return requireRow(res, "pending delivery")
}

func (s *Store) now() time.Time {
	return time.Now().UTC()
}
