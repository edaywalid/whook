package api

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/edaywalid/whook/internal/core"
	"github.com/edaywalid/whook/internal/dispatch"
	"github.com/edaywalid/whook/internal/id"
	"github.com/edaywalid/whook/internal/ingest"
	"github.com/edaywalid/whook/internal/source"
	"github.com/edaywalid/whook/internal/storage"
)

func (s *Server) handleIngest(w http.ResponseWriter, r *http.Request) {
	if !s.ingestLimit.Allow(r.PathValue("source")) {
		w.Header().Set("Retry-After", "1")
		writeError(w, http.StatusTooManyRequests, "rate limit exceeded for this source")
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, s.maxBody))
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "reading body")
		return
	}

	req := ingest.Request{
		Method:   r.Method,
		Path:     r.URL.Path,
		Query:    r.URL.RawQuery,
		Headers:  r.Header,
		Body:     body,
		SourceIP: clientIP(r),
	}

	source := r.PathValue("source")
	start := s.now()
	res, err := s.capturer.Capture(r.Context(), source, req)
	if errors.Is(err, ingest.ErrSourceNotFound) {
		writeError(w, http.StatusNotFound, "unknown source")
		return
	}
	if err != nil {
		s.logger.Error("capturing webhook", "error", err)
		writeError(w, http.StatusInternalServerError, "capture failed")
		return
	}

	if s.metrics != nil {
		s.metrics.RecordCapture(source, captureStatus(res.Accepted), s.now().Sub(start).Seconds())
	}

	if !res.Accepted {
		writeJSON(w, http.StatusUnauthorized, map[string]any{
			"event_id": res.EventID,
			"error":    "signature verification failed",
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"event_id": res.EventID})
}

func (s *Server) handleListEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := storage.EventFilter{
		Source: q.Get("source"),
		Status: core.EventStatus(q.Get("status")),
		Limit:  atoiDefault(q.Get("limit"), 50),
		Offset: atoiDefault(q.Get("offset"), 0),
	}
	events, err := s.store.ListEvents(r.Context(), filter)
	if err != nil {
		s.logger.Error("listing events", "error", err)
		writeError(w, http.StatusInternalServerError, "listing events")
		return
	}

	out := make([]eventSummary, 0, len(events))
	for _, e := range events {
		out = append(out, toSummary(e))
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": out})
}

func (s *Server) handleGetEvent(w http.ResponseWriter, r *http.Request) {
	eventID := r.PathValue("id")
	event, err := s.store.GetEvent(r.Context(), eventID)
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, "event not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "loading event")
		return
	}

	attempts, err := s.store.ListAttempts(r.Context(), eventID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "loading attempts")
		return
	}
	deliveries, err := s.store.DeliveriesForEvent(r.Context(), eventID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "loading deliveries")
		return
	}

	urlOf := s.destinationURLs(r.Context(), event.Source)
	writeJSON(w, http.StatusOK, toDetail(event, attempts, deliveries, urlOf))
}

// destinationURLs returns a lookup from destination id to URL for a source.
// Failures yield an empty map so the caller still renders without URLs.
func (s *Server) destinationURLs(ctx context.Context, source string) func(string) string {
	urls := map[string]string{}
	if dests, err := s.store.ListDestinations(ctx, source); err == nil {
		for _, d := range dests {
			urls[d.ID] = d.URL
		}
	}
	return func(id string) string { return urls[id] }
}

func (s *Server) handleListDeadLetters(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	dead, err := s.store.ListDeadLetters(r.Context(), atoiDefault(q.Get("limit"), 50), atoiDefault(q.Get("offset"), 0))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "listing dead letters")
		return
	}

	out := make([]deadLetterDTO, 0, len(dead))
	for _, p := range dead {
		url := ""
		if dest, err := s.store.GetDestination(r.Context(), p.DestinationID); err == nil {
			url = dest.URL
		}
		out = append(out, toDeadLetterDTO(p, url))
	}
	writeJSON(w, http.StatusOK, map[string]any{"dead_letters": out})
}

func (s *Server) handleReplay(w http.ResponseWriter, r *http.Request) {
	eventID := r.PathValue("id")
	event, err := s.store.GetEvent(r.Context(), eventID)
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, "event not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "loading event")
		return
	}

	var queued int
	if destID := r.URL.Query().Get("destination_id"); destID != "" {
		dest, err := s.store.GetDestination(r.Context(), destID)
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "destination not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "loading destination")
			return
		}
		if dest.Source != event.Source {
			writeError(w, http.StatusBadRequest, "destination belongs to a different source")
			return
		}
		queued, err = dispatch.EnqueueOne(r.Context(), s.store, event, dest, true, s.now())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "queuing replay")
			return
		}
	} else {
		queued, err = dispatch.Enqueue(r.Context(), s.store, event, true, s.now())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "queuing replay")
			return
		}
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"replayed": queued})
}

func (s *Server) handleListSources(w http.ResponseWriter, r *http.Request) {
	sources, err := s.sources.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "listing sources")
		return
	}
	out := make([]sourceDTO, 0, len(sources))
	for _, src := range sources {
		out = append(out, toSourceDTO(src))
	}
	writeJSON(w, http.StatusOK, map[string]any{"sources": out})
}

func (s *Server) handleCreateSource(w http.ResponseWriter, r *http.Request) {
	var in sourceInput
	if err := readJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	src, err := s.sources.Register(r.Context(), in.toServiceInput())
	if err != nil {
		s.writeSourceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toSourceDTO(src))
}

func (s *Server) handleUpdateSource(w http.ResponseWriter, r *http.Request) {
	var in sourceInput
	if err := readJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	in.Name = r.PathValue("name")
	src, err := s.sources.Update(r.Context(), in.toServiceInput())
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, "source not found")
		return
	}
	if err != nil {
		s.writeSourceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toSourceDTO(src))
}

func (s *Server) writeSourceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, source.ErrUnknownVerifier):
		writeError(w, http.StatusBadRequest, err.Error())
	case strings.Contains(err.Error(), "UNIQUE"):
		writeError(w, http.StatusConflict, "source name already exists")
	default:
		s.logger.Error("source operation", "error", err)
		writeError(w, http.StatusInternalServerError, "source operation failed")
	}
}

func (s *Server) handleListDestinations(w http.ResponseWriter, r *http.Request) {
	src := r.URL.Query().Get("source")
	if src == "" {
		writeError(w, http.StatusBadRequest, "source query parameter is required")
		return
	}
	dests, err := s.store.ListDestinations(r.Context(), src)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "listing destinations")
		return
	}
	out := make([]destinationDTO, 0, len(dests))
	for _, d := range dests {
		out = append(out, toDestinationDTO(d))
	}
	writeJSON(w, http.StatusOK, map[string]any{"destinations": out})
}

func (s *Server) handleCreateDestination(w http.ResponseWriter, r *http.Request) {
	var in destinationInput
	if err := readJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if in.Source == "" || in.URL == "" {
		writeError(w, http.StatusBadRequest, "source and url are required")
		return
	}

	dest := core.Destination{
		ID:         id.New("dst"),
		Source:     in.Source,
		URL:        in.URL,
		Enabled:    in.enabled(),
		FilterSpec: in.FilterSpec,
		Backoff:    in.Backoff,
		CreatedAt:  s.now(),
	}
	if err := s.store.CreateDestination(r.Context(), dest); err != nil {
		writeError(w, http.StatusInternalServerError, "creating destination")
		return
	}
	writeJSON(w, http.StatusCreated, toDestinationDTO(dest))
}

func (s *Server) handleUpdateDestination(w http.ResponseWriter, r *http.Request) {
	destID := r.PathValue("id")
	dest, err := s.store.GetDestination(r.Context(), destID)
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, "destination not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "loading destination")
		return
	}

	var in destinationInput
	if err := readJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if in.URL != "" {
		dest.URL = in.URL
	}
	if in.Enabled != nil {
		dest.Enabled = *in.Enabled
	}
	if in.FilterSpec != nil {
		dest.FilterSpec = in.FilterSpec
	}
	if in.Backoff != nil {
		dest.Backoff = in.Backoff
	}
	if err := s.store.UpdateDestination(r.Context(), dest); err != nil {
		writeError(w, http.StatusInternalServerError, "updating destination")
		return
	}
	writeJSON(w, http.StatusOK, toDestinationDTO(dest))
}

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.TrimSpace(strings.Split(fwd, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func captureStatus(accepted bool) string {
	if accepted {
		return "received"
	}
	return "rejected"
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
