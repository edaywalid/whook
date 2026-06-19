// Package ingest captures inbound webhooks: it verifies signatures, persists
// the raw request durably, and queues delivery. It does the minimum synchronous
// work so the provider gets a fast response.
package ingest

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/edaywalid/whook/internal/core"
	"github.com/edaywalid/whook/internal/dispatch"
	"github.com/edaywalid/whook/internal/id"
	"github.com/edaywalid/whook/internal/signature"
	"github.com/edaywalid/whook/internal/source"
	"github.com/edaywalid/whook/internal/storage"
)

// ErrSourceNotFound means the request targeted an unregistered source.
var ErrSourceNotFound = errors.New("ingest: source not found")

// Request is a captured inbound HTTP request.
type Request struct {
	Method   string
	Path     string
	Query    string
	Headers  map[string][]string
	Body     []byte
	SourceIP string
}

// Result reports what happened to a captured request.
type Result struct {
	EventID  string
	Accepted bool // false when signature verification failed
}

// Capturer owns the capture path.
type Capturer struct {
	store     storage.Store
	sources   *source.Service
	verifiers *signature.Registry
	logger    *slog.Logger
	now       func() time.Time
}

// NewCapturer wires the capture path.
func NewCapturer(store storage.Store, sources *source.Service, verifiers *signature.Registry, logger *slog.Logger) *Capturer {
	return &Capturer{
		store:     store,
		sources:   sources,
		verifiers: verifiers,
		logger:    logger,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

// Capture persists the request and, when accepted, queues delivery. A request
// that fails signature verification is still captured (status rejected) but not
// queued, so forged requests remain inspectable.
func (c *Capturer) Capture(ctx context.Context, sourceName string, req Request) (Result, error) {
	src, err := c.sources.Get(ctx, sourceName)
	if errors.Is(err, storage.ErrNotFound) {
		return Result{}, ErrSourceNotFound
	}
	if err != nil {
		return Result{}, err
	}

	status := core.StatusReceived
	if src.Verifier != "" {
		if err := c.verify(ctx, src, req); err != nil {
			status = core.StatusRejected
		}
	}

	event := core.Event{
		ID:         id.New("evt"),
		Source:     src.Name,
		ReceivedAt: c.now(),
		SourceIP:   req.SourceIP,
		Method:     req.Method,
		Path:       req.Path,
		Query:      req.Query,
		Headers:    req.Headers,
		Body:       req.Body,
		DedupKey:   dedupKey(src, req.Headers),
		Status:     status,
	}

	eventID, created, err := c.store.SaveEvent(ctx, event)
	if err != nil {
		return Result{}, err
	}
	event.ID = eventID

	accepted := status == core.StatusReceived
	if accepted && created {
		if _, err := dispatch.Enqueue(ctx, c.store, event, false, c.now()); err != nil {
			return Result{}, err
		}
	}
	return Result{EventID: eventID, Accepted: accepted}, nil
}

func (c *Capturer) verify(ctx context.Context, src core.Source, req Request) error {
	secret, err := c.sources.Secret(ctx, src.Name)
	if err != nil {
		return err
	}
	return c.verifiers.Verify(src.Verifier, secret, signature.Request{
		Headers: req.Headers,
		Body:    req.Body,
	})
}

func dedupKey(src core.Source, headers map[string][]string) string {
	if src.DedupHeader == "" {
		return ""
	}
	for k, v := range headers {
		if len(v) > 0 && strings.EqualFold(k, src.DedupHeader) {
			return v[0]
		}
	}
	return ""
}
