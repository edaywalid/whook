package ingest_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"github.com/edaywalid/whook/internal/core"
	"github.com/edaywalid/whook/internal/crypto"
	"github.com/edaywalid/whook/internal/ingest"
	"github.com/edaywalid/whook/internal/signature"
	"github.com/edaywalid/whook/internal/source"
	"github.com/edaywalid/whook/internal/storage/sqlite"
)

type fixture struct {
	store    *sqlite.Store
	sources  *source.Service
	capturer *ingest.Capturer
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	dsn := "file:" + filepath.Join(t.TempDir(), "test.db")
	st, err := sqlite.Open(context.Background(), dsn)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	sealer, _ := crypto.NewSealer("test-key")
	reg := signature.NewRegistry()
	svc := source.NewService(st, sealer, reg)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return fixture{store: st, sources: svc, capturer: ingest.NewCapturer(st, svc, reg, logger)}
}

func (f fixture) addDestination(t *testing.T, src string) {
	t.Helper()
	err := f.store.CreateDestination(context.Background(), core.Destination{
		ID: "d1", Source: src, URL: "http://localhost:9/sink", Enabled: true, CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create destination: %v", err)
	}
}

func TestCaptureAcceptedEnqueuesDelivery(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	if _, err := f.sources.Register(ctx, source.Input{Name: "stripe"}); err != nil {
		t.Fatalf("register: %v", err)
	}
	f.addDestination(t, "stripe")

	res, err := f.capturer.Capture(ctx, "stripe", ingest.Request{
		Method:  "POST",
		Headers: map[string][]string{"Content-Type": {"application/json"}},
		Body:    []byte(`{"type":"payment.succeeded"}`),
	})
	if err != nil {
		t.Fatalf("capture: %v", err)
	}
	if !res.Accepted {
		t.Fatal("expected accepted")
	}

	deliveries, err := f.store.DeliveriesForEvent(ctx, res.EventID)
	if err != nil {
		t.Fatalf("deliveries: %v", err)
	}
	if len(deliveries) != 1 {
		t.Fatalf("expected 1 queued delivery, got %d", len(deliveries))
	}
}

func TestCaptureRejectedNotEnqueued(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	if _, err := f.sources.Register(ctx, source.Input{Name: "stripe", Verifier: "stripe", Secret: "whsec"}); err != nil {
		t.Fatalf("register: %v", err)
	}
	f.addDestination(t, "stripe")

	res, err := f.capturer.Capture(ctx, "stripe", ingest.Request{
		Method:  "POST",
		Headers: map[string][]string{"Stripe-Signature": {"t=1,v1=deadbeef"}},
		Body:    []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("capture: %v", err)
	}
	if res.Accepted {
		t.Fatal("expected rejected for bad signature")
	}

	event, err := f.store.GetEvent(ctx, res.EventID)
	if err != nil {
		t.Fatalf("get event: %v", err)
	}
	if event.Status != core.StatusRejected {
		t.Fatalf("expected rejected status, got %q", event.Status)
	}

	deliveries, _ := f.store.DeliveriesForEvent(ctx, res.EventID)
	if len(deliveries) != 0 {
		t.Fatalf("rejected event must not enqueue delivery, got %d", len(deliveries))
	}
}

func TestCaptureUnknownSource(t *testing.T) {
	f := newFixture(t)
	_, err := f.capturer.Capture(context.Background(), "nope", ingest.Request{Method: "POST", Headers: map[string][]string{}, Body: []byte(`{}`)})
	if !errors.Is(err, ingest.ErrSourceNotFound) {
		t.Fatalf("expected ErrSourceNotFound, got %v", err)
	}
}

func TestCaptureDedup(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	if _, err := f.sources.Register(ctx, source.Input{Name: "stripe", DedupHeader: "Idempotency-Key"}); err != nil {
		t.Fatalf("register: %v", err)
	}

	req := ingest.Request{
		Method:  "POST",
		Headers: map[string][]string{"Idempotency-Key": {"abc-123"}},
		Body:    []byte(`{}`),
	}
	first, err := f.capturer.Capture(ctx, "stripe", req)
	if err != nil {
		t.Fatalf("first capture: %v", err)
	}
	second, err := f.capturer.Capture(ctx, "stripe", req)
	if err != nil {
		t.Fatalf("second capture: %v", err)
	}
	if first.EventID != second.EventID {
		t.Fatalf("dedup key should resolve to one event, got %q and %q", first.EventID, second.EventID)
	}
}
