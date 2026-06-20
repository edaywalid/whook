package postgres_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/edaywalid/whook/internal/core"
	"github.com/edaywalid/whook/internal/storage"
	"github.com/edaywalid/whook/internal/storage/postgres"
)

// newStore opens the Postgres store against WHOOK_TEST_POSTGRES_URL and starts
// each test from clean tables. The whole package skips when the env var is
// unset, so the suite is a no-op without a database.
func newStore(t *testing.T) *postgres.Store {
	t.Helper()
	dsn := os.Getenv("WHOOK_TEST_POSTGRES_URL")
	if dsn == "" {
		t.Skip("set WHOOK_TEST_POSTGRES_URL to run the postgres storage tests")
	}

	st, err := postgres.Open(context.Background(), dsn)
	if err != nil {
		t.Fatalf("opening store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	raw, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("opening cleanup connection: %v", err)
	}
	defer raw.Close()
	_, err = raw.Exec("TRUNCATE sources, events, destinations, delivery_attempts, pending_deliveries")
	if err != nil {
		t.Fatalf("truncating: %v", err)
	}
	return st
}

func days(n int) time.Duration { return time.Duration(n) * 24 * time.Hour }

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
	if err != nil || !created1 {
		t.Fatalf("first save: id=%q created=%v err=%v", id1, created1, err)
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
	if err != storage.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDestinationRoundTrip(t *testing.T) {
	ctx := context.Background()
	st := newStore(t)

	dest := core.Destination{
		ID: "d1", Source: "stripe", URL: "https://app.example/hook",
		Enabled: true, CreatedAt: time.Now().UTC(),
		FilterSpec: &core.FilterSpec{BodyEquals: map[string]string{"type": "x"}},
	}
	if err := st.CreateDestination(ctx, dest); err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := st.GetDestination(ctx, "d1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !got.Enabled || got.FilterSpec == nil || got.FilterSpec.BodyEquals["type"] != "x" {
		t.Fatalf("unexpected destination: %+v", got)
	}
}

func TestPruneRetention(t *testing.T) {
	ctx := context.Background()
	st := newStore(t)

	save := func(id string, age time.Duration, status core.EventStatus) {
		err := func() error {
			_, _, err := st.SaveEvent(ctx, core.Event{
				ID: id, Source: "stripe", ReceivedAt: time.Now().UTC().Add(-age),
				Method: "POST", Path: "/x", Headers: map[string][]string{}, Body: []byte(`{}`),
				Status: status,
			})
			return err
		}()
		if err != nil {
			t.Fatalf("save %s: %v", id, err)
		}
	}
	save("old-received", days(40), core.StatusReceived)
	save("recent-received", days(5), core.StatusReceived)
	save("ancient-rejected", days(100), core.StatusRejected)

	normalCutoff := time.Now().UTC().Add(-days(30))
	failedCutoff := time.Now().UTC().Add(-days(90))
	deleted, err := st.Prune(ctx, normalCutoff, failedCutoff)
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if deleted != 2 {
		t.Fatalf("expected 2 deleted (old-received, ancient-rejected), got %d", deleted)
	}
	if _, err := st.GetEvent(ctx, "recent-received"); err != nil {
		t.Fatalf("recent-received should survive: %v", err)
	}
}
