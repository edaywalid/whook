package api_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/edaywalid/whook/internal/api"
	"github.com/edaywalid/whook/internal/crypto"
	"github.com/edaywalid/whook/internal/ingest"
	"github.com/edaywalid/whook/internal/ratelimit"
	"github.com/edaywalid/whook/internal/signature"
	"github.com/edaywalid/whook/internal/source"
	"github.com/edaywalid/whook/internal/storage/sqlite"
)

func TestIngestRateLimited(t *testing.T) {
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
	if _, err := svc.Register(context.Background(), source.Input{Name: "stripe"}); err != nil {
		t.Fatalf("register: %v", err)
	}

	srv := api.NewServer(api.Config{
		Store:         st,
		Capturer:      ingest.NewCapturer(st, svc, reg, logger),
		Sources:       svc,
		IngestLimiter: ratelimit.New(0.01, 1), // burst of 1, slow refill
		MaxBody:       1 << 20,
		Logger:        logger,
	})
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)

	first := ingestPost(t, ts.URL)
	if first != http.StatusOK {
		t.Fatalf("first request should be accepted, got %d", first)
	}
	second := ingestPost(t, ts.URL)
	if second != http.StatusTooManyRequests {
		t.Fatalf("second request should be rate limited, got %d", second)
	}
}

func ingestPost(t *testing.T, base string) int {
	t.Helper()
	resp, err := http.Post(base+"/ingest/stripe", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}
