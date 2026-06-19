package delivery_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/edaywalid/whook/internal/core"
	"github.com/edaywalid/whook/internal/delivery"
)

type fakeDoer struct {
	resp *http.Response
	err  error
}

func (f fakeDoer) Do(*http.Request) (*http.Response, error) {
	return f.resp, f.err
}

func response(status int, headers map[string]string) *http.Response {
	h := http.Header{}
	for k, v := range headers {
		h.Set(k, v)
	}
	return &http.Response{
		StatusCode: status,
		Header:     h,
		Body:       io.NopCloser(strings.NewReader("ok")),
	}
}

func TestClassify(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		expected core.Outcome
	}{
		{"200 success", 200, core.OutcomeSuccess},
		{"201 success", 201, core.OutcomeSuccess},
		{"429 retryable", 429, core.OutcomeRetryable},
		{"500 retryable", 500, core.OutcomeRetryable},
		{"503 retryable", 503, core.OutcomeRetryable},
		{"400 permanent", 400, core.OutcomePermanent},
		{"404 permanent", 404, core.OutcomePermanent},
		{"422 permanent", 422, core.OutcomePermanent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := delivery.Classify(tt.status); got != tt.expected {
				t.Errorf("Classify(%d) = %v, expected %v", tt.status, got, tt.expected)
			}
		})
	}
}

func TestDeliver(t *testing.T) {
	event := core.Event{ID: "e1", Source: "stripe", Body: []byte(`{}`)}
	dest := core.Destination{URL: "https://example.test/hook"}

	t.Run("2xx is success", func(t *testing.T) {
		eng := delivery.New(fakeDoer{resp: response(200, nil)})
		res := eng.Deliver(context.Background(), event, dest)
		if res.Outcome != core.OutcomeSuccess {
			t.Fatalf("expected success, got %v", res.Outcome)
		}
		if res.StatusCode != 200 {
			t.Fatalf("expected status 200, got %d", res.StatusCode)
		}
	})

	t.Run("transport error is retryable", func(t *testing.T) {
		eng := delivery.New(fakeDoer{err: errors.New("connection refused")})
		res := eng.Deliver(context.Background(), event, dest)
		if res.Outcome != core.OutcomeRetryable {
			t.Fatalf("expected retryable, got %v", res.Outcome)
		}
		if res.StatusCode != 0 {
			t.Fatalf("expected no status code, got %d", res.StatusCode)
		}
	})

	t.Run("429 with retry-after is honored", func(t *testing.T) {
		eng := delivery.New(fakeDoer{resp: response(429, map[string]string{"Retry-After": "30"})})
		res := eng.Deliver(context.Background(), event, dest)
		if res.Outcome != core.OutcomeRetryable {
			t.Fatalf("expected retryable, got %v", res.Outcome)
		}
		if res.RetryAfter != 30*time.Second {
			t.Fatalf("expected RetryAfter 30s, got %v", res.RetryAfter)
		}
	})

	t.Run("429 without retry-after has no override", func(t *testing.T) {
		eng := delivery.New(fakeDoer{resp: response(429, nil)})
		res := eng.Deliver(context.Background(), event, dest)
		if res.RetryAfter != 0 {
			t.Fatalf("expected no RetryAfter, got %v", res.RetryAfter)
		}
	})

	t.Run("4xx is permanent", func(t *testing.T) {
		eng := delivery.New(fakeDoer{resp: response(400, nil)})
		res := eng.Deliver(context.Background(), event, dest)
		if res.Outcome != core.OutcomePermanent {
			t.Fatalf("expected permanent, got %v", res.Outcome)
		}
	})
}
