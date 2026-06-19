// Package app wires the gateway's components and runs the HTTP server and
// delivery worker until the context is cancelled.
package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/edaywalid/whook/internal/api"
	"github.com/edaywalid/whook/internal/config"
	"github.com/edaywalid/whook/internal/crypto"
	"github.com/edaywalid/whook/internal/delivery"
	"github.com/edaywalid/whook/internal/ingest"
	"github.com/edaywalid/whook/internal/metrics"
	"github.com/edaywalid/whook/internal/pruner"
	"github.com/edaywalid/whook/internal/retry"
	"github.com/edaywalid/whook/internal/signature"
	"github.com/edaywalid/whook/internal/source"
	"github.com/edaywalid/whook/internal/storage/sqlite"
	"github.com/edaywalid/whook/internal/worker"
)

// Run builds the gateway from cfg and blocks until ctx is cancelled, then shuts
// down gracefully.
func Run(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	store, err := sqlite.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer func() { _ = store.Close() }()

	sealer, err := crypto.NewSealer(cfg.SecretKey)
	if err != nil {
		return err
	}

	metricsCollector := metrics.New()
	verifiers := signature.NewRegistry()
	sources := source.NewService(store, sealer, verifiers)
	capturer := ingest.NewCapturer(store, sources, verifiers, logger)
	engine := delivery.New(&http.Client{Timeout: 30 * time.Second})

	deliveryWorker := worker.New(store, engine, logger, worker.Options{
		Concurrency:  cfg.Workers,
		PollInterval: cfg.PollInterval,
		Backoff:      retry.DefaultBackoff,
		Metrics:      metricsCollector,
	})

	retentionPruner := pruner.New(store, logger, pruner.Options{
		RetentionAge:       cfg.RetentionAge,
		FailedRetentionAge: cfg.FailedRetentionAge,
		Interval:           cfg.PruneInterval,
	})

	srv := api.NewServer(api.Config{
		Store:      store,
		Capturer:   capturer,
		Sources:    sources,
		Metrics:    metricsCollector,
		AdminToken: cfg.AdminToken,
		MaxBody:    cfg.MaxBodyBytes,
		Logger:     logger,
	})
	httpServer := &http.Server{
		Addr:              cfg.Addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 2)
	go func() {
		if err := deliveryWorker.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			errCh <- err
		}
	}()
	go func() {
		if err := retentionPruner.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			errCh <- err
		}
	}()
	if retentionPruner.Enabled() {
		logger.Info("retention pruning enabled", "retention", cfg.RetentionAge, "interval", cfg.PruneInterval)
	}
	go func() {
		logger.Info("http server listening", "addr", cfg.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-errCh:
		logger.Error("component failed", "error", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("http shutdown", "error", err)
	}
	return nil
}
