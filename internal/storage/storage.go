// Package storage defines the persistence seam for the gateway and its errors.
// Implementations live in subpackages (sqlite today, postgres later).
package storage

import (
	"context"
	"errors"
	"time"

	"github.com/edaywalid/whook/internal/core"
)

// ErrNotFound is returned when a requested row does not exist.
var ErrNotFound = errors.New("storage: not found")

// EventFilter narrows and paginates an event listing. A zero Source or Status
// means "any".
type EventFilter struct {
	Source string
	Status core.EventStatus
	Limit  int
	Offset int
}

// Store is the single persistence interface backing the gateway. SQLite and,
// later, Postgres implement it so the backend is swappable.
type Store interface {
	CreateSource(ctx context.Context, s core.Source, secret []byte) error
	UpdateSource(ctx context.Context, s core.Source, secret []byte) error
	GetSource(ctx context.Context, name string) (core.Source, error)
	ListSources(ctx context.Context) ([]core.Source, error)
	SourceSecret(ctx context.Context, name string) ([]byte, error)

	// SaveEvent persists an event. When the event carries a dedup key already
	// seen for its source, no row is inserted and the existing id is returned
	// with created=false.
	SaveEvent(ctx context.Context, e core.Event) (id string, created bool, err error)
	GetEvent(ctx context.Context, id string) (core.Event, error)
	ListEvents(ctx context.Context, f EventFilter) ([]core.Event, error)

	CreateDestination(ctx context.Context, d core.Destination) error
	UpdateDestination(ctx context.Context, d core.Destination) error
	GetDestination(ctx context.Context, id string) (core.Destination, error)
	ListDestinations(ctx context.Context, source string) ([]core.Destination, error)

	RecordAttempt(ctx context.Context, a core.DeliveryAttempt) error
	ListAttempts(ctx context.Context, eventID string) ([]core.DeliveryAttempt, error)

	Enqueue(ctx context.Context, p core.PendingDelivery) error
	// ClaimDue atomically leases up to limit due deliveries so concurrent
	// workers never claim the same row.
	ClaimDue(ctx context.Context, now time.Time, lease time.Duration, limit int) ([]core.PendingDelivery, error)
	MarkDelivered(ctx context.Context, id string) error
	Reschedule(ctx context.Context, id string, attempt int, next time.Time) error
	MarkDeadLetter(ctx context.Context, id string) error

	Close() error
}
