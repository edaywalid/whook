// Package core holds the domain types shared across the gateway modules.
package core

import "time"

// EventStatus is the capture status of an inbound request.
type EventStatus string

const (
	StatusReceived EventStatus = "received"
	StatusRejected EventStatus = "rejected"
)

// Outcome classifies a single delivery attempt.
type Outcome string

const (
	OutcomeSuccess   Outcome = "success"
	OutcomeRetryable Outcome = "retryable"
	OutcomePermanent Outcome = "permanent"
)

// DeliveryState is the pending-or-terminal state of an (event, destination)
// delivery in the queue.
type DeliveryState string

const (
	StatePending    DeliveryState = "pending"
	StateDelivered  DeliveryState = "delivered"
	StateDeadLetter DeliveryState = "dead_letter"
)

// Source is a registered inbound integration. Webhooks POST to /ingest/{name}.
type Source struct {
	ID          string
	Name        string
	Verifier    string // empty means no signature verification
	HasSecret   bool
	DedupHeader string
	CreatedAt   time.Time
}

// Event is a captured inbound request.
type Event struct {
	ID         string
	Source     string
	ReceivedAt time.Time
	SourceIP   string
	Method     string
	Path       string
	Query      string
	Headers    map[string][]string
	Body       []byte
	DedupKey   string // empty when the provider supplied none
	Status     EventStatus
}

// Destination is a target the gateway forwards matching events to.
type Destination struct {
	ID         string
	Source     string
	URL        string
	Enabled    bool
	FilterSpec *FilterSpec    // nil matches everything
	Backoff    *BackoffConfig // nil uses the gateway default
	CreatedAt  time.Time
}

// DeliveryAttempt records the result of one delivery try.
type DeliveryAttempt struct {
	ID              string
	EventID         string
	DestinationID   string
	AttemptNumber   int
	StatusCode      int // 0 when no response was received
	ResponseSnippet string
	DurationMS      int64
	Outcome         Outcome
	IsReplay        bool
	AttemptedAt     time.Time
}

// PendingDelivery is a queued delivery for an (event, destination) pair.
type PendingDelivery struct {
	ID            string
	EventID       string
	DestinationID string
	AttemptNumber int
	NextAttemptAt time.Time
	State         DeliveryState
	IsReplay      bool
}

// BackoffConfig controls retry timing.
type BackoffConfig struct {
	BaseDelay   time.Duration `json:"base_delay"`
	Multiplier  float64       `json:"multiplier"`
	MaxDelay    time.Duration `json:"max_delay"`
	MaxAttempts int           `json:"max_attempts"`
}

// FilterSpec is a destination's routing filter. All clauses must match (AND);
// an empty spec matches every event.
type FilterSpec struct {
	BodyEquals   map[string]string `json:"body_equals,omitempty"`
	HeaderEquals map[string]string `json:"header_equals,omitempty"`
}
