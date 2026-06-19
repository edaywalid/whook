// Package dispatch turns matched destinations into queued deliveries. It is
// shared by the ingest path and by replay.
package dispatch

import (
	"context"
	"time"

	"github.com/edaywalid/whook/internal/core"
	"github.com/edaywalid/whook/internal/id"
	"github.com/edaywalid/whook/internal/routing"
)

// Enqueuer is the slice of storage needed to queue deliveries.
type Enqueuer interface {
	ListDestinations(ctx context.Context, source string) ([]core.Destination, error)
	Enqueue(ctx context.Context, p core.PendingDelivery) error
}

// Enqueue resolves the event to its source's destinations and queues an
// immediately-due delivery for each. It returns the number of deliveries
// queued.
func Enqueue(ctx context.Context, store Enqueuer, event core.Event, isReplay bool, now time.Time) (int, error) {
	dests, err := store.ListDestinations(ctx, event.Source)
	if err != nil {
		return 0, err
	}
	return enqueueTo(ctx, store, event, routing.Resolve(event, dests), isReplay, now)
}

// EnqueueOne queues a delivery for a single destination, applying its filter.
// It is used by replay targeting a specific destination.
func EnqueueOne(ctx context.Context, store Enqueuer, event core.Event, dest core.Destination, isReplay bool, now time.Time) (int, error) {
	return enqueueTo(ctx, store, event, routing.Resolve(event, []core.Destination{dest}), isReplay, now)
}

func enqueueTo(ctx context.Context, store Enqueuer, event core.Event, dests []core.Destination, isReplay bool, now time.Time) (int, error) {
	for _, d := range dests {
		p := core.PendingDelivery{
			ID:            id.New("pd"),
			EventID:       event.ID,
			DestinationID: d.ID,
			AttemptNumber: 0,
			NextAttemptAt: now,
			IsReplay:      isReplay,
		}
		if err := store.Enqueue(ctx, p); err != nil {
			return 0, err
		}
	}
	return len(dests), nil
}
