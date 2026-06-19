// Package worker drains the pending-delivery queue: it claims due deliveries,
// performs them through the delivery engine, records each attempt, and either
// reschedules with backoff or dead-letters on exhaustion.
package worker

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/edaywalid/whook/internal/core"
	"github.com/edaywalid/whook/internal/delivery"
	"github.com/edaywalid/whook/internal/id"
	"github.com/edaywalid/whook/internal/retry"
	"github.com/edaywalid/whook/internal/storage"
)

// Options configures the worker pool.
type Options struct {
	Concurrency  int
	PollInterval time.Duration
	Lease        time.Duration
	Backoff      core.BackoffConfig
}

// Worker drains the queue using an injected delivery engine.
type Worker struct {
	store  storage.Store
	engine *delivery.Engine
	logger *slog.Logger
	opts   Options
	now    func() time.Time
}

// New builds a worker. Concurrency, PollInterval, and Lease fall back to sane
// defaults when zero.
func New(store storage.Store, engine *delivery.Engine, logger *slog.Logger, opts Options) *Worker {
	if opts.Concurrency < 1 {
		opts.Concurrency = 4
	}
	if opts.PollInterval <= 0 {
		opts.PollInterval = time.Second
	}
	if opts.Lease <= 0 {
		opts.Lease = 30 * time.Second
	}
	if opts.Backoff.MaxAttempts == 0 {
		opts.Backoff = retry.DefaultBackoff
	}
	return &Worker{
		store:  store,
		engine: engine,
		logger: logger,
		opts:   opts,
		now:    func() time.Time { return time.Now().UTC() },
	}
}

// Run polls the queue until the context is cancelled.
func (w *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.opts.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *Worker) tick(ctx context.Context) {
	due, err := w.store.ClaimDue(ctx, w.now(), w.opts.Lease, w.opts.Concurrency)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			w.logger.Error("claiming due deliveries", "error", err)
		}
		return
	}

	var wg sync.WaitGroup
	for _, pending := range due {
		wg.Add(1)
		go func(p core.PendingDelivery) {
			defer wg.Done()
			w.process(ctx, p)
		}(pending)
	}
	wg.Wait()
}

func (w *Worker) process(ctx context.Context, pending core.PendingDelivery) {
	event, err := w.store.GetEvent(ctx, pending.EventID)
	if err != nil {
		w.logger.Error("loading event for delivery", "event_id", pending.EventID, "error", err)
		return
	}
	dest, err := w.store.GetDestination(ctx, pending.DestinationID)
	if err != nil {
		w.logger.Error("loading destination for delivery", "destination_id", pending.DestinationID, "error", err)
		return
	}

	attempt := pending.AttemptNumber + 1
	res := w.engine.Deliver(ctx, event, dest)

	record := core.DeliveryAttempt{
		ID:              id.New("att"),
		EventID:         event.ID,
		DestinationID:   dest.ID,
		AttemptNumber:   attempt,
		StatusCode:      res.StatusCode,
		ResponseSnippet: res.Snippet,
		DurationMS:      res.Duration.Milliseconds(),
		Outcome:         res.Outcome,
		IsReplay:        pending.IsReplay,
		AttemptedAt:     w.now(),
	}
	if err := w.store.RecordAttempt(ctx, record); err != nil {
		w.logger.Error("recording attempt", "event_id", event.ID, "error", err)
	}

	w.settle(ctx, pending, dest, attempt, res)
}

func (w *Worker) settle(ctx context.Context, pending core.PendingDelivery, dest core.Destination, attempt int, res delivery.Result) {
	switch res.Outcome {
	case core.OutcomeSuccess:
		w.mark(ctx, w.store.MarkDelivered, pending.ID, "delivered")
	case core.OutcomePermanent:
		w.mark(ctx, w.store.MarkDeadLetter, pending.ID, "dead-letter")
	case core.OutcomeRetryable:
		w.reschedule(ctx, pending, dest, attempt, res.RetryAfter)
	}
}

func (w *Worker) reschedule(ctx context.Context, pending core.PendingDelivery, dest core.Destination, attempt int, retryAfter time.Duration) {
	cfg := w.opts.Backoff
	if dest.Backoff != nil {
		cfg = *dest.Backoff
	}

	plan := retry.Next(cfg, attempt, retryAfter)
	if plan.DeadLetter {
		w.mark(ctx, w.store.MarkDeadLetter, pending.ID, "dead-letter")
		return
	}
	next := w.now().Add(plan.Delay)
	if err := w.store.Reschedule(ctx, pending.ID, attempt, next); err != nil {
		w.logger.Error("rescheduling delivery", "pending_id", pending.ID, "error", err)
	}
}

func (w *Worker) mark(ctx context.Context, fn func(context.Context, string) error, pendingID, label string) {
	if err := fn(ctx, pendingID); err != nil {
		w.logger.Error("marking delivery "+label, "pending_id", pendingID, "error", err)
	}
}
