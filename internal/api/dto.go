package api

import (
	"time"

	"github.com/edaywalid/whook/internal/core"
	"github.com/edaywalid/whook/internal/source"
)

type eventSummary struct {
	ID         string    `json:"id"`
	Source     string    `json:"source"`
	ReceivedAt time.Time `json:"received_at"`
	Status     string    `json:"status"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
}

func toSummary(e core.Event) eventSummary {
	return eventSummary{
		ID:         e.ID,
		Source:     e.Source,
		ReceivedAt: e.ReceivedAt,
		Status:     string(e.Status),
		Method:     e.Method,
		Path:       e.Path,
	}
}

type attemptDTO struct {
	ID            string    `json:"id"`
	DestinationID string    `json:"destination_id"`
	AttemptNumber int       `json:"attempt_number"`
	StatusCode    int       `json:"status_code,omitempty"`
	Snippet       string    `json:"response_snippet,omitempty"`
	DurationMS    int64     `json:"duration_ms"`
	Outcome       string    `json:"outcome"`
	IsReplay      bool      `json:"is_replay"`
	AttemptedAt   time.Time `json:"attempted_at"`
}

type eventDetail struct {
	eventSummary
	SourceIP string              `json:"source_ip,omitempty"`
	Query    string              `json:"query,omitempty"`
	DedupKey string              `json:"dedup_key,omitempty"`
	Headers  map[string][]string `json:"headers"`
	Body     string              `json:"body"`
	Attempts []attemptDTO        `json:"attempts"`
}

func toDetail(e core.Event, attempts []core.DeliveryAttempt) eventDetail {
	out := eventDetail{
		eventSummary: toSummary(e),
		SourceIP:     e.SourceIP,
		Query:        e.Query,
		DedupKey:     e.DedupKey,
		Headers:      e.Headers,
		Body:         string(e.Body),
		Attempts:     make([]attemptDTO, 0, len(attempts)),
	}
	for _, a := range attempts {
		out.Attempts = append(out.Attempts, attemptDTO{
			ID:            a.ID,
			DestinationID: a.DestinationID,
			AttemptNumber: a.AttemptNumber,
			StatusCode:    a.StatusCode,
			Snippet:       a.ResponseSnippet,
			DurationMS:    a.DurationMS,
			Outcome:       string(a.Outcome),
			IsReplay:      a.IsReplay,
			AttemptedAt:   a.AttemptedAt,
		})
	}
	return out
}

type sourceDTO struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Verifier    string    `json:"verifier,omitempty"`
	HasSecret   bool      `json:"has_secret"`
	DedupHeader string    `json:"dedup_header,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func toSourceDTO(s core.Source) sourceDTO {
	return sourceDTO{
		ID:          s.ID,
		Name:        s.Name,
		Verifier:    s.Verifier,
		HasSecret:   s.HasSecret,
		DedupHeader: s.DedupHeader,
		CreatedAt:   s.CreatedAt,
	}
}

type sourceInput struct {
	Name        string `json:"name"`
	Verifier    string `json:"verifier"`
	Secret      string `json:"secret"`
	DedupHeader string `json:"dedup_header"`
}

func (in sourceInput) toServiceInput() source.Input {
	return source.Input{
		Name:        in.Name,
		Verifier:    in.Verifier,
		Secret:      in.Secret,
		DedupHeader: in.DedupHeader,
	}
}

type destinationDTO struct {
	ID         string              `json:"id"`
	Source     string              `json:"source"`
	URL        string              `json:"url"`
	Enabled    bool                `json:"enabled"`
	FilterSpec *core.FilterSpec    `json:"filter_spec,omitempty"`
	Backoff    *core.BackoffConfig `json:"backoff,omitempty"`
	CreatedAt  time.Time           `json:"created_at"`
}

func toDestinationDTO(d core.Destination) destinationDTO {
	return destinationDTO{
		ID:         d.ID,
		Source:     d.Source,
		URL:        d.URL,
		Enabled:    d.Enabled,
		FilterSpec: d.FilterSpec,
		Backoff:    d.Backoff,
		CreatedAt:  d.CreatedAt,
	}
}

type destinationInput struct {
	Source     string              `json:"source"`
	URL        string              `json:"url"`
	Enabled    *bool               `json:"enabled"`
	FilterSpec *core.FilterSpec    `json:"filter_spec"`
	Backoff    *core.BackoffConfig `json:"backoff"`
}

// enabled defaults to true when the field is omitted on create.
func (in destinationInput) enabled() bool {
	return in.Enabled == nil || *in.Enabled
}
