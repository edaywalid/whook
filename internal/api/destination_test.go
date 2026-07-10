package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/edaywalid/whook/internal/source"
)

func authedRequest(t *testing.T, method, url string, body []byte) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func doJSON(t *testing.T, req *http.Request) *http.Response {
	t.Helper()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", req.Method, req.URL, err)
	}
	return resp
}

func TestCreateDestinationURLValidation(t *testing.T) {
	ts, svc := newTestServer(t)
	if _, err := svc.Register(context.Background(), source.Input{Name: "stripe"}); err != nil {
		t.Fatalf("register source: %v", err)
	}

	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{"valid http url", "http://example.com/hook", http.StatusCreated},
		{"valid https url", "https://example.com/hook", http.StatusCreated},
		{"relative path rejected", "/just/a/path", http.StatusBadRequest},
		{"missing scheme rejected", "example.com/hook", http.StatusBadRequest},
		{"ftp scheme rejected", "ftp://example.com/hook", http.StatusBadRequest},
		{"empty host rejected", "http:///hook", http.StatusBadRequest},
		{"scheme only rejected", "https://", http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(map[string]string{"source": "stripe", "url": tt.url})
			resp := doJSON(t, authedRequest(t, http.MethodPost, ts.URL+"/destinations", body))
			defer resp.Body.Close()
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("url %q: expected %d, got %d", tt.url, tt.wantStatus, resp.StatusCode)
			}
		})
	}
}

func TestCreateDestinationBackoffValidation(t *testing.T) {
	ts, svc := newTestServer(t)
	if _, err := svc.Register(context.Background(), source.Input{Name: "stripe"}); err != nil {
		t.Fatalf("register source: %v", err)
	}

	// MaxAttempts of 0 would dead-letter on the first failure; reject it early.
	body, _ := json.Marshal(map[string]any{
		"source":  "stripe",
		"url":     "https://example.com/hook",
		"backoff": map[string]any{"base_delay": 1e9, "multiplier": 2, "max_delay": 6e10, "max_attempts": 0},
	})
	resp := doJSON(t, authedRequest(t, http.MethodPost, ts.URL+"/destinations", body))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for zero max_attempts, got %d", resp.StatusCode)
	}
}

func TestUpdateDestinationURLValidation(t *testing.T) {
	ts, svc := newTestServer(t)
	if _, err := svc.Register(context.Background(), source.Input{Name: "stripe"}); err != nil {
		t.Fatalf("register source: %v", err)
	}

	body, _ := json.Marshal(map[string]string{"source": "stripe", "url": "https://example.com/hook"})
	createResp := doJSON(t, authedRequest(t, http.MethodPost, ts.URL+"/destinations", body))
	defer createResp.Body.Close()
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", createResp.StatusCode)
	}
	var created map[string]any
	if err := json.NewDecoder(createResp.Body).Decode(&created); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	destID := created["id"].(string)

	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{"valid https url accepted", "https://other.example.com/hook", http.StatusOK},
		{"ftp scheme rejected", "ftp://example.com/hook", http.StatusBadRequest},
		{"relative path rejected", "/bad", http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateBody, _ := json.Marshal(map[string]string{"url": tt.url})
			resp := doJSON(t, authedRequest(t, http.MethodPatch, fmt.Sprintf("%s/destinations/%s", ts.URL, destID), updateBody))
			defer resp.Body.Close()
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("url %q: expected %d, got %d", tt.url, tt.wantStatus, resp.StatusCode)
			}
		})
	}
}

func TestListEventsRejectsUnknownStatus(t *testing.T) {
	ts, _ := newTestServer(t)

	resp := doJSON(t, authedRequest(t, http.MethodGet, ts.URL+"/events?status=bogus", nil))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown status, got %d", resp.StatusCode)
	}

	ok := doJSON(t, authedRequest(t, http.MethodGet, ts.URL+"/events?status=received", nil))
	defer ok.Body.Close()
	if ok.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for known status, got %d", ok.StatusCode)
	}
}

func TestListEventsClampsHugeLimit(t *testing.T) {
	ts, _ := newTestServer(t)
	// A huge limit must not error; it is clamped server-side.
	resp := doJSON(t, authedRequest(t, http.MethodGet, ts.URL+"/events?limit=100000000&offset=-5", nil))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 with clamped paging, got %d", resp.StatusCode)
	}
}
