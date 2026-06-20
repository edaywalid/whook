package worker_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/edaywalid/whook/internal/core"
	"github.com/edaywalid/whook/internal/delivery"
	"github.com/edaywalid/whook/internal/worker"
)

func TestGracefulDrainCompletesInFlight(t *testing.T) {
	var once sync.Once
	started := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		once.Do(func() { close(started) })
		time.Sleep(150 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	st := openStore(t)
	w := worker.New(st, delivery.New(srv.Client()), slog.New(slog.NewTextHandler(io.Discard, nil)), worker.Options{
		Concurrency:  2,
		PollInterval: 10 * time.Millisecond,
		Lease:        time.Second,
		Backoff:      backoffFast(),
		DrainTimeout: 3 * time.Second,
	})
	event := seed(t, st, srv.URL)

	ctx, cancel := context.WithCancel(context.Background())
	runDone := make(chan struct{})
	go func() {
		_ = w.Run(ctx)
		close(runDone)
	}()

	// Cancel only once the delivery is mid-flight.
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("delivery never started")
	}
	cancel()

	select {
	case <-runDone:
	case <-time.After(3 * time.Second):
		t.Fatal("Run did not return after draining")
	}

	// The in-flight delivery should have finished rather than being cancelled.
	attempts := waitForAttempts(t, st, event.ID, 1)
	if attempts[0].Outcome != core.OutcomeSuccess {
		t.Fatalf("expected the in-flight delivery to complete with success, got %v", attempts[0].Outcome)
	}
}
