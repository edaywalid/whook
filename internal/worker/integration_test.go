package worker_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/edaywalid/whook/internal/core"
	"github.com/edaywalid/whook/internal/delivery"
	"github.com/edaywalid/whook/internal/dispatch"
	"github.com/edaywalid/whook/internal/storage/sqlite"
	"github.com/edaywalid/whook/internal/worker"
)

func openStore(t *testing.T) *sqlite.Store {
	t.Helper()
	dsn := "file:" + filepath.Join(t.TempDir(), "test.db")
	st, err := sqlite.Open(context.Background(), dsn)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func runWorker(t *testing.T, st *sqlite.Store, eng *delivery.Engine, backoff core.BackoffConfig) {
	t.Helper()
	w := worker.New(st, eng, slog.New(slog.NewTextHandler(io.Discard, nil)), worker.Options{
		Concurrency:  2,
		PollInterval: 10 * time.Millisecond,
		Lease:        time.Second,
		Backoff:      backoff,
	})
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go w.Run(ctx)
}

func seed(t *testing.T, st *sqlite.Store, url string) core.Event {
	t.Helper()
	ctx := context.Background()
	dest := core.Destination{ID: "d1", Source: "stripe", URL: url, Enabled: true, CreatedAt: time.Now().UTC()}
	if err := st.CreateDestination(ctx, dest); err != nil {
		t.Fatalf("create destination: %v", err)
	}
	event := core.Event{
		ID:         "e1",
		Source:     "stripe",
		ReceivedAt: time.Now().UTC(),
		Method:     "POST",
		Path:       "/ingest/stripe",
		Headers:    map[string][]string{},
		Body:       []byte(`{"type":"payment.succeeded"}`),
		Status:     core.StatusReceived,
	}
	if _, _, err := st.SaveEvent(ctx, event); err != nil {
		t.Fatalf("save event: %v", err)
	}
	if _, err := dispatch.Enqueue(ctx, st, event, false, time.Now().UTC()); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	return event
}

func waitForAttempts(t *testing.T, st *sqlite.Store, eventID string, want int) []core.DeliveryAttempt {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		attempts, err := st.ListAttempts(context.Background(), eventID)
		if err != nil {
			t.Fatalf("list attempts: %v", err)
		}
		if len(attempts) >= want {
			return attempts
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %d attempts", want)
	return nil
}

func TestDeliversSuccessfully(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	st := openStore(t)
	runWorker(t, st, delivery.New(srv.Client()), backoffFast())
	event := seed(t, st, srv.URL)

	attempts := waitForAttempts(t, st, event.ID, 1)
	if attempts[0].Outcome != core.OutcomeSuccess {
		t.Fatalf("expected success, got %v", attempts[0].Outcome)
	}
}

func TestRetriesThenSucceeds(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	st := openStore(t)
	runWorker(t, st, delivery.New(srv.Client()), backoffFast())
	event := seed(t, st, srv.URL)

	attempts := waitForAttempts(t, st, event.ID, 2)
	last := attempts[len(attempts)-1]
	if last.Outcome != core.OutcomeSuccess {
		t.Fatalf("expected final success, got %v", last.Outcome)
	}
	if attempts[0].Outcome != core.OutcomeRetryable {
		t.Fatalf("expected first attempt retryable, got %v", attempts[0].Outcome)
	}
}

func TestPermanentFailureDeadLetters(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	st := openStore(t)
	runWorker(t, st, delivery.New(srv.Client()), backoffFast())
	event := seed(t, st, srv.URL)

	attempts := waitForAttempts(t, st, event.ID, 1)
	if attempts[0].Outcome != core.OutcomePermanent {
		t.Fatalf("expected permanent, got %v", attempts[0].Outcome)
	}

	// A permanent failure must not be retried.
	time.Sleep(100 * time.Millisecond)
	again, _ := st.ListAttempts(context.Background(), event.ID)
	if len(again) != 1 {
		t.Fatalf("permanent failure should not retry, got %d attempts", len(again))
	}
}

func backoffFast() core.BackoffConfig {
	return core.BackoffConfig{
		BaseDelay:   10 * time.Millisecond,
		Multiplier:  2,
		MaxDelay:    50 * time.Millisecond,
		MaxAttempts: 5,
	}
}
