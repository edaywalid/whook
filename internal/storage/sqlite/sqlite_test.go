package sqlite_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/edaywalid/whook/internal/core"
	"github.com/edaywalid/whook/internal/storage"
	"github.com/edaywalid/whook/internal/storage/sqlite"
)

func newStore(t *testing.T) *sqlite.Store {
	t.Helper()
	dsn := "file:" + filepath.Join(t.TempDir(), "test.db")
	st, err := sqlite.Open(context.Background(), dsn)
	if err != nil {
		t.Fatalf("opening store: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func TestSaveEventIdempotentOnDedupKey(t *testing.T) {
	ctx := context.Background()
	st := newStore(t)

	base := core.Event{
		Source:     "stripe",
		ReceivedAt: time.Now().UTC(),
		Method:     "POST",
		Path:       "/ingest/stripe",
		Headers:    map[string][]string{"Content-Type": {"application/json"}},
		Body:       []byte(`{"type":"payment.succeeded"}`),
		DedupKey:   "evt_123",
		Status:     core.StatusReceived,
	}

	first := base
	first.ID = "id-1"
	id1, created1, err := st.SaveEvent(ctx, first)
	if err != nil {
		t.Fatalf("first save: %v", err)
	}
	if !created1 {
		t.Fatal("first save should create the event")
	}

	second := base
	second.ID = "id-2"
	id2, created2, err := st.SaveEvent(ctx, second)
	if err != nil {
		t.Fatalf("second save: %v", err)
	}
	if created2 {
		t.Fatal("second save with same dedup key should not create a new event")
	}
	if id1 != id2 {
		t.Fatalf("expected same event id, got %q and %q", id1, id2)
	}
}

func TestSaveEventWithoutDedupKeyAlwaysDistinct(t *testing.T) {
	ctx := context.Background()
	st := newStore(t)

	mk := func(id string) core.Event {
		return core.Event{
			ID:         id,
			Source:     "github",
			ReceivedAt: time.Now().UTC(),
			Method:     "POST",
			Path:       "/ingest/github",
			Headers:    map[string][]string{},
			Body:       []byte(`{}`),
			Status:     core.StatusReceived,
		}
	}

	for _, id := range []string{"a", "b"} {
		_, created, err := st.SaveEvent(ctx, mk(id))
		if err != nil {
			t.Fatalf("save %s: %v", id, err)
		}
		if !created {
			t.Fatalf("event %s should be created", id)
		}
	}
}

func TestClaimDueLeasesEachRowOnce(t *testing.T) {
	ctx := context.Background()
	st := newStore(t)

	now := time.Now().UTC()
	for _, id := range []string{"p1", "p2"} {
		err := st.Enqueue(ctx, core.PendingDelivery{
			ID:            id,
			EventID:       "e1",
			DestinationID: "d1",
			NextAttemptAt: now.Add(-time.Minute),
		})
		if err != nil {
			t.Fatalf("enqueue %s: %v", id, err)
		}
	}

	claimed, err := st.ClaimDue(ctx, now, time.Minute, 10)
	if err != nil {
		t.Fatalf("first claim: %v", err)
	}
	if len(claimed) != 2 {
		t.Fatalf("expected 2 due deliveries, got %d", len(claimed))
	}

	again, err := st.ClaimDue(ctx, now, time.Minute, 10)
	if err != nil {
		t.Fatalf("second claim: %v", err)
	}
	if len(again) != 0 {
		t.Fatalf("leased rows should not be reclaimed, got %d", len(again))
	}
}

func TestGetSourceNotFound(t *testing.T) {
	st := newStore(t)
	_, err := st.GetSource(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for missing source")
	}
	if want := storage.ErrNotFound; err != want {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
