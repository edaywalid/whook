package sqlite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/edaywalid/whook/internal/core"
)

func (s *Store) DeliveriesForEvent(ctx context.Context, eventID string) ([]core.PendingDelivery, error) {
	const q = `SELECT id, event_id, destination_id, attempt_number, next_attempt_at, state, is_replay
		FROM pending_deliveries WHERE event_id = ? ORDER BY created_at`
	rows, err := s.db.QueryContext(ctx, q, eventID)
	if err != nil {
		return nil, fmt.Errorf("listing deliveries for event %q: %w", eventID, err)
	}
	defer rows.Close()

	out := []core.PendingDelivery{}
	for rows.Next() {
		p, err := scanPending(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning delivery: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) ListDeadLetters(ctx context.Context, limit, offset int) ([]core.PendingDelivery, error) {
	if limit <= 0 {
		limit = 50
	}
	const q = `SELECT id, event_id, destination_id, attempt_number, next_attempt_at, state, is_replay
		FROM pending_deliveries WHERE state = 'dead_letter'
		ORDER BY updated_at DESC LIMIT ? OFFSET ?`
	rows, err := s.db.QueryContext(ctx, q, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("listing dead letters: %w", err)
	}
	defer rows.Close()

	out := []core.PendingDelivery{}
	for rows.Next() {
		p, err := scanPending(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning dead letter: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) Prune(ctx context.Context, normalCutoff, failedCutoff time.Time) (int, error) {
	ids, err := s.deletableEventIDs(ctx, normalCutoff, failedCutoff)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("beginning prune transaction: %w", err)
	}
	defer tx.Rollback()

	tables := []string{"delivery_attempts", "pending_deliveries", "events"}
	for _, chunk := range chunkStrings(ids, 400) {
		placeholders, args := inClause(chunk)
		for _, table := range tables {
			column := "event_id"
			if table == "events" {
				column = "id"
			}
			q := "DELETE FROM " + table + " WHERE " + column + " IN (" + placeholders + ")"
			if _, err := tx.ExecContext(ctx, q, args...); err != nil {
				return 0, fmt.Errorf("pruning %s: %w", table, err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("committing prune: %w", err)
	}
	return len(ids), nil
}

// deletableEventIDs finds normal events past normalCutoff and failed (rejected
// or dead-lettered) events past the longer failedCutoff.
func (s *Store) deletableEventIDs(ctx context.Context, normalCutoff, failedCutoff time.Time) ([]string, error) {
	const q = `SELECT id FROM events WHERE
		(received_at < ? AND status = 'received'
			AND id NOT IN (SELECT event_id FROM pending_deliveries WHERE state = 'dead_letter'))
		OR
		(received_at < ? AND (status = 'rejected'
			OR id IN (SELECT event_id FROM pending_deliveries WHERE state = 'dead_letter')))`
	rows, err := s.db.QueryContext(ctx, q, normalCutoff, failedCutoff)
	if err != nil {
		return nil, fmt.Errorf("selecting prunable events: %w", err)
	}
	defer rows.Close()

	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scanning event id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func inClause(ids []string) (string, []any) {
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return strings.Repeat("?,", len(ids)-1) + "?", args
}

func chunkStrings(s []string, size int) [][]string {
	var chunks [][]string
	for i := 0; i < len(s); i += size {
		end := min(i+size, len(s))
		chunks = append(chunks, s[i:end])
	}
	return chunks
}
