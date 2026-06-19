package retry_test

import (
	"testing"
	"time"

	"github.com/edaywalid/whook/internal/core"
	"github.com/edaywalid/whook/internal/retry"
)

func TestNext(t *testing.T) {
	cfg := core.BackoffConfig{
		BaseDelay:   time.Second,
		Multiplier:  2,
		MaxDelay:    10 * time.Second,
		MaxAttempts: 4,
	}

	tests := []struct {
		name          string
		attempt       int
		retryAfter    time.Duration
		expectedDelay time.Duration
		expectedDead  bool
	}{
		{name: "first attempt", attempt: 1, expectedDelay: time.Second},
		{name: "second attempt doubles", attempt: 2, expectedDelay: 2 * time.Second},
		{name: "third attempt quadruples", attempt: 3, expectedDelay: 4 * time.Second},
		{name: "exhaustion boundary dead-letters", attempt: 4, expectedDead: true},
		{name: "past budget dead-letters", attempt: 5, expectedDead: true},
		{name: "retry-after overrides backoff", attempt: 1, retryAfter: 30 * time.Second, expectedDelay: 30 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := retry.Next(cfg, tt.attempt, tt.retryAfter)
			if got.DeadLetter != tt.expectedDead {
				t.Fatalf("DeadLetter = %v, expected %v", got.DeadLetter, tt.expectedDead)
			}
			if !tt.expectedDead && got.Delay != tt.expectedDelay {
				t.Fatalf("Delay = %v, expected %v", got.Delay, tt.expectedDelay)
			}
		})
	}
}

func TestNextCapsAtMaxDelay(t *testing.T) {
	cfg := core.BackoffConfig{
		BaseDelay:   time.Second,
		Multiplier:  2,
		MaxDelay:    3 * time.Second,
		MaxAttempts: 10,
	}
	got := retry.Next(cfg, 3, 0) // would be 4s uncapped
	if got.Delay != 3*time.Second {
		t.Fatalf("expected delay capped at 3s, got %v", got.Delay)
	}
}
