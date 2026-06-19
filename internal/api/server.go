// Package api exposes the gateway over HTTP: the public ingest endpoint, the
// admin JSON API, and a thin server-rendered dashboard.
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/edaywalid/whook/internal/ingest"
	"github.com/edaywalid/whook/internal/source"
	"github.com/edaywalid/whook/internal/storage"
)

// Server holds the API dependencies.
type Server struct {
	store      storage.Store
	capturer   *ingest.Capturer
	sources    *source.Service
	adminToken string
	maxBody    int64
	logger     *slog.Logger
	now        func() time.Time
}

// Config wires a Server.
type Config struct {
	Store      storage.Store
	Capturer   *ingest.Capturer
	Sources    *source.Service
	AdminToken string
	MaxBody    int64
	Logger     *slog.Logger
}

// NewServer builds a Server.
func NewServer(cfg Config) *Server {
	return &Server{
		store:      cfg.Store,
		capturer:   cfg.Capturer,
		sources:    cfg.Sources,
		adminToken: cfg.AdminToken,
		maxBody:    cfg.MaxBody,
		logger:     cfg.Logger,
		now:        func() time.Time { return time.Now().UTC() },
	}
}

// Handler builds the HTTP router. The ingest endpoint and health check are
// public; everything else requires the admin token.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("POST /ingest/{source}", s.handleIngest)

	mux.Handle("GET /events", s.admin(s.handleListEvents))
	mux.Handle("GET /events/{id}", s.admin(s.handleGetEvent))
	mux.Handle("POST /events/{id}/replay", s.admin(s.handleReplay))

	mux.Handle("GET /sources", s.admin(s.handleListSources))
	mux.Handle("POST /sources", s.admin(s.handleCreateSource))
	mux.Handle("PATCH /sources/{name}", s.admin(s.handleUpdateSource))

	mux.Handle("GET /destinations", s.admin(s.handleListDestinations))
	mux.Handle("POST /destinations", s.admin(s.handleCreateDestination))
	mux.Handle("PATCH /destinations/{id}", s.admin(s.handleUpdateDestination))

	mux.Handle("GET /ui", s.admin(s.handleDashboard))
	mux.HandleFunc("GET /{$}", s.handleRoot)

	return mux
}

// admin wraps a handler with bearer-token authentication. The token may arrive
// as "Authorization: Bearer <token>" or, for browser convenience on the
// dashboard, as a "token" query parameter. An empty configured token disables
// the check (local development only).
func (s *Server) admin(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.adminToken == "" {
			next(w, r)
			return
		}
		if subtleEqual(s.presentedToken(r), s.adminToken) {
			next(w, r)
			return
		}
		writeError(w, http.StatusUnauthorized, "missing or invalid admin token")
	})
}

func (s *Server) presentedToken(r *http.Request) string {
	if h := r.Header.Get("Authorization"); h != "" {
		if token, ok := strings.CutPrefix(h, "Bearer "); ok {
			return token
		}
	}
	return r.URL.Query().Get("token")
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/ui", http.StatusFound)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func readJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

// subtleEqual compares two short secrets in constant time.
func subtleEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := range len(a) {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}
