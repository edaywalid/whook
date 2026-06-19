// Package delivery performs a single HTTP delivery to a destination and
// classifies the result. The HTTP transport is injected so the classification
// core can be tested without real network calls.
package delivery

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/edaywalid/whook/internal/core"
)

// maxSnippet caps how much of the destination response body we retain.
const maxSnippet = 512

// Doer is the subset of *http.Client the engine needs.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Result is the outcome of one delivery attempt.
type Result struct {
	Outcome    core.Outcome
	StatusCode int           // 0 when no response was received
	Snippet    string        // truncated response body
	Duration   time.Duration // wall-clock time of the attempt
	RetryAfter time.Duration // from a 429 Retry-After header, if present
}

// Engine delivers events over an injected HTTP client.
type Engine struct {
	client Doer
	now    func() time.Time
}

// New returns an Engine using the given HTTP client.
func New(client Doer) *Engine {
	return &Engine{client: client, now: time.Now}
}

// Deliver POSTs the event body to the destination, forwarding the captured
// headers, and classifies the response. A transport error is retryable.
func (e *Engine) Deliver(ctx context.Context, event core.Event, dest core.Destination) Result {
	start := e.now()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, dest.URL, bytes.NewReader(event.Body))
	if err != nil {
		return Result{Outcome: core.OutcomePermanent, Duration: e.now().Sub(start)}
	}
	forwardHeaders(req, event)

	resp, err := e.client.Do(req)
	if err != nil {
		return Result{Outcome: core.OutcomeRetryable, Duration: e.now().Sub(start)}
	}
	defer resp.Body.Close()

	snippet, _ := io.ReadAll(io.LimitReader(resp.Body, maxSnippet))

	return Result{
		Outcome:    Classify(resp.StatusCode),
		StatusCode: resp.StatusCode,
		Snippet:    string(snippet),
		Duration:   e.now().Sub(start),
		RetryAfter: retryAfter(resp),
	}
}

// Classify maps an HTTP status code to a delivery outcome. 2xx succeeds; 429
// and 5xx are retryable; every other non-2xx is permanent.
func Classify(status int) core.Outcome {
	switch {
	case status >= 200 && status < 300:
		return core.OutcomeSuccess
	case status == http.StatusTooManyRequests || status >= 500:
		return core.OutcomeRetryable
	default:
		return core.OutcomePermanent
	}
}

func forwardHeaders(req *http.Request, event core.Event) {
	for name, values := range event.Headers {
		switch http.CanonicalHeaderKey(name) {
		case "Host", "Content-Length", "Connection", "Transfer-Encoding":
			continue
		}
		for _, v := range values {
			req.Header.Add(name, v)
		}
	}
	req.Header.Set("X-Whook-Event-Id", event.ID)
	req.Header.Set("X-Whook-Source", event.Source)
}

// retryAfter reads a Retry-After header expressed in seconds. Header-date form
// is ignored for the MVP; the scheduler falls back to its computed backoff.
func retryAfter(resp *http.Response) time.Duration {
	if resp.StatusCode != http.StatusTooManyRequests {
		return 0
	}
	v := resp.Header.Get("Retry-After")
	if v == "" {
		return 0
	}
	secs, err := strconv.Atoi(v)
	if err != nil || secs < 0 {
		return 0
	}
	return time.Duration(secs) * time.Second
}
