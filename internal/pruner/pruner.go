// Package pruner periodically deletes captured events past their retention
// window so the store does not grow without bound.
package pruner

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/edaywalid/whook/internal/storage"
)

// Options configures retention.
type Options struct {
	// RetentionAge is how long received events are kept. Zero disables pruning.
	RetentionAge time.Duration
	// FailedRetentionAge keeps rejected and dead-lettered events longer.
	FailedRetentionAge time.Duration
	// Interval is how often the pruner runs.
	Interval time.Duration
}

// Pruner deletes aged events on a schedule.
type Pruner struct {
	store  storage.Store
	logger *slog.Logger
	opts   Options
	now    func() time.Time
}

// New builds a pruner.
func New(store storage.Store, logger *slog.Logger, opts Options) *Pruner {
	if opts.Interval <= 0 {
		opts.Interval = time.Hour
	}
	if opts.FailedRetentionAge < opts.RetentionAge {
		opts.FailedRetentionAge = opts.RetentionAge
	}
	return &Pruner{
		store:  store,
		logger: logger,
		opts:   opts,
		now:    func() time.Time { return time.Now().UTC() },
	}
}

// Enabled reports whether retention is configured.
func (p *Pruner) Enabled() bool {
	return p.opts.RetentionAge > 0
}

// Run prunes on Interval until the context is cancelled. It returns immediately
// when retention is disabled.
func (p *Pruner) Run(ctx context.Context) error {
	if !p.Enabled() {
		return nil
	}

	ticker := time.NewTicker(p.opts.Interval)
	defer ticker.Stop()

	p.pruneOnce(ctx)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			p.pruneOnce(ctx)
		}
	}
}

func (p *Pruner) pruneOnce(ctx context.Context) {
	now := p.now()
	normalCutoff := now.Add(-p.opts.RetentionAge)
	failedCutoff := now.Add(-p.opts.FailedRetentionAge)

	deleted, err := p.store.Prune(ctx, normalCutoff, failedCutoff)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			p.logger.Error("pruning events", "error", err)
		}
		return
	}
	if deleted > 0 {
		p.logger.Info("pruned events", "deleted", deleted)
	}
}
