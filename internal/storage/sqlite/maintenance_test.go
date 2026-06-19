package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/edaywalid/whook/internal/core"
	"github.com/edaywalid/whook/internal/storage/sqlite"
)

func days(n int) time.Duration { return time.Duration(n) * 24 * time.Hour }

func saveAged(t *testing.T, st *sqlite.Store, id string, age time.Duration, status core.EventStatus, deadLetter bool) {
	t.Helper()
	ctx := context.Background()
	event := core.Event{
		ID:         id,
		Source:     "stripe",
		ReceivedAt: time.Now().UTC().Add(-age),
		Method:     "POST",
		Path:       "/ingest/stripe",
		Headers:    map[string][]string{},
		Body:       []byte(`{}`),
		Status:     status,
	}
	if _, _, err := st.SaveEvent(ctx, event); err != nil {
		t.Fatalf("save %s: %v", id, err)
	}
	if deadLetter {
		pendingID := id + "-pd"
		if err := st.Enqueue(ctx, core.PendingDelivery{
			ID:            pendingID,
			EventID:       id,
			DestinationID: "d1",
			NextAttemptAt: time.Now().UTC(),
		}); err != nil {
			t.Fatalf("enqueue %s: %v", id, err)
		}
		if err := st.MarkDeadLetter(ctx, pendingID); err != nil {
			t.Fatalf("dead-letter %s: %v", id, err)
		}
	}
}

func TestPruneRetention(t *testing.T) {
	ctx := context.Background()
	st := newStore(t)

	saveAged(t, st, "old-received", days(40), core.StatusReceived, false)
	saveAged(t, st, "recent-received", days(10), core.StatusReceived, false)
	saveAged(t, st, "old-rejected", days(40), core.StatusRejected, false)
	saveAged(t, st, "ancient-rejected", days(100), core.StatusRejected, false)
	saveAged(t, st, "old-deadlettered", days(40), core.StatusReceived, true)
	saveAged(t, st, "ancient-deadlettered", days(100), core.StatusReceived, true)

	normalCutoff := time.Now().UTC().Add(-days(30))
	failedCutoff := time.Now().UTC().Add(-days(90))

	deleted, err := st.Prune(ctx, normalCutoff, failedCutoff)
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if deleted != 3 {
		t.Fatalf("expected 3 events deleted, got %d", deleted)
	}

	survivors := map[string]bool{
		"recent-received":  true,
		"old-rejected":     true,
		"old-deadlettered": true,
	}
	for id := range survivors {
		if _, err := st.GetEvent(ctx, id); err != nil {
			t.Errorf("expected %q to survive prune, got %v", id, err)
		}
	}
	for _, id := range []string{"old-received", "ancient-rejected", "ancient-deadlettered"} {
		if _, err := st.GetEvent(ctx, id); err == nil {
			t.Errorf("expected %q to be pruned", id)
		}
	}
}

func TestListDeadLetters(t *testing.T) {
	ctx := context.Background()
	st := newStore(t)

	saveAged(t, st, "dl-event", days(1), core.StatusReceived, true)
	saveAged(t, st, "live-event", days(1), core.StatusReceived, false)

	dead, err := st.ListDeadLetters(ctx, 50, 0)
	if err != nil {
		t.Fatalf("list dead letters: %v", err)
	}
	if len(dead) != 1 {
		t.Fatalf("expected 1 dead letter, got %d", len(dead))
	}
	if dead[0].EventID != "dl-event" {
		t.Fatalf("expected dl-event, got %q", dead[0].EventID)
	}
}
