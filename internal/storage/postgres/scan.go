package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/edaywalid/whook/internal/core"
	"github.com/edaywalid/whook/internal/storage"
)

// scanner is satisfied by both *sql.Row and *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func scanEvent(row scanner) (core.Event, error) {
	var (
		e           core.Event
		sourceIP    sql.NullString
		query       sql.NullString
		dedupKey    sql.NullString
		headersJSON string
		status      string
	)
	err := row.Scan(&e.ID, &e.Source, &e.ReceivedAt, &sourceIP, &e.Method, &e.Path,
		&query, &headersJSON, &e.Body, &dedupKey, &status)
	if err != nil {
		return core.Event{}, err
	}
	if err := json.Unmarshal([]byte(headersJSON), &e.Headers); err != nil {
		return core.Event{}, fmt.Errorf("unmarshaling headers: %w", err)
	}
	e.SourceIP = sourceIP.String
	e.Query = query.String
	e.DedupKey = dedupKey.String
	e.Status = core.EventStatus(status)
	return e, nil
}

func scanDestination(row scanner) (core.Destination, error) {
	var (
		d       core.Destination
		backoff sql.NullString
		filter  sql.NullString
	)
	err := row.Scan(&d.ID, &d.Source, &d.URL, &d.Enabled, &backoff, &filter, &d.CreatedAt)
	if err != nil {
		return core.Destination{}, err
	}
	if filter.Valid && filter.String != "" {
		var spec core.FilterSpec
		if err := json.Unmarshal([]byte(filter.String), &spec); err != nil {
			return core.Destination{}, fmt.Errorf("unmarshaling filter spec: %w", err)
		}
		d.FilterSpec = &spec
	}
	if backoff.Valid && backoff.String != "" {
		var cfg core.BackoffConfig
		if err := json.Unmarshal([]byte(backoff.String), &cfg); err != nil {
			return core.Destination{}, fmt.Errorf("unmarshaling backoff config: %w", err)
		}
		d.Backoff = &cfg
	}
	return d, nil
}

func scanPending(row scanner) (core.PendingDelivery, error) {
	var p core.PendingDelivery
	err := row.Scan(&p.ID, &p.EventID, &p.DestinationID, &p.AttemptNumber, &p.NextAttemptAt, &p.State, &p.IsReplay)
	return p, err
}

func marshalDestination(d core.Destination) (filter, backoff any, err error) {
	if d.FilterSpec != nil {
		b, err := json.Marshal(d.FilterSpec)
		if err != nil {
			return nil, nil, fmt.Errorf("marshaling filter spec: %w", err)
		}
		filter = string(b)
	}
	if d.Backoff != nil {
		b, err := json.Marshal(d.Backoff)
		if err != nil {
			return nil, nil, fmt.Errorf("marshaling backoff config: %w", err)
		}
		backoff = string(b)
	}
	return filter, backoff, nil
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// requireRow turns a no-op UPDATE into storage.ErrNotFound.
func requireRow(res sql.Result, entity string) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking %s rows affected: %w", entity, err)
	}
	if n == 0 {
		return storage.ErrNotFound
	}
	return nil
}
