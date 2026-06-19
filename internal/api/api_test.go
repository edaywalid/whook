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
	"github.com/edaywalid/whook/internal/signature"
	"github.com/edaywalid/whook/internal/source"
	"github.com/edaywalid/whook/internal/storage/sqlite"
)

const adminToken = "secret-token"

func newTestServer(t *testing.T) (*httptest.Server, *source.Service) {
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
	capturer := ingest.NewCapturer(st, svc, reg, logger)

	srv := api.NewServer(api.Config{
		Store:      st,
		Capturer:   capturer,
		Sources:    svc,
		AdminToken: adminToken,
		MaxBody:    1 << 20,
		Logger:     logger,
	})
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts, svc
}

func TestIngestUnknownSourceReturns404(t *testing.T) {
	ts, _ := newTestServer(t)
	resp, err := http.Post(ts.URL+"/ingest/nope", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestIngestAcceptedReturns200(t *testing.T) {
	ts, svc := newTestServer(t)
	if _, err := svc.Register(context.Background(), source.Input{Name: "stripe"}); err != nil {
		t.Fatalf("register: %v", err)
	}

	resp, err := http.Post(ts.URL+"/ingest/stripe", "application/json", strings.NewReader(`{"x":1}`))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "event_id") {
		t.Fatalf("expected event_id in response, got %s", body)
	}
}

func TestIngestIsPublicButManagementRequiresToken(t *testing.T) {
	ts, _ := newTestServer(t)

	noToken, err := http.Get(ts.URL + "/events")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer noToken.Body.Close()
	if noToken.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", noToken.StatusCode)
	}

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/events", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	withToken, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("get with token: %v", err)
	}
	defer withToken.Body.Close()
	if withToken.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 with token, got %d", withToken.StatusCode)
	}
}

func TestHealthIsPublic(t *testing.T) {
	ts, _ := newTestServer(t)
	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
