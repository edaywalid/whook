// Package retry computes the deterministic backoff schedule for redeliveries
// and decides when a delivery has exhausted its budget. It is pure logic.
package retry

import (
	"math"
	"time"

	"github.com/edaywalid/whook/internal/core"
)

// DefaultBackoff is used when a destination sets no override.
var DefaultBackoff = core.BackoffConfig{
	BaseDelay:   2 * time.Second,
	Multiplier:  3,
	MaxDelay:    time.Hour,
	MaxAttempts: 8,
}

// Plan is the scheduler's decision after an attempt completes.
type Plan struct {
	// DeadLetter is true when the retry budget is exhausted; Delay is then
	// meaningless.
	DeadLetter bool
	// Delay is how long to wait before the next attempt.
	Delay time.Duration
}

// Next decides what to do after the given attempt number has failed. attempt is
// 1-based: 1 means the first attempt just failed. When attempt reaches
// MaxAttempts the delivery is dead-lettered.
//
// When retryAfter is non-zero (a 429 Retry-After), it overrides the computed
// backoff for this step.
func Next(cfg core.BackoffConfig, attempt int, retryAfter time.Duration) Plan {
	if attempt >= cfg.MaxAttempts {
		return Plan{DeadLetter: true}
	}
	if retryAfter > 0 {
		return Plan{Delay: retryAfter}
	}
	return Plan{Delay: delayFor(cfg, attempt)}
}

// delayFor returns base * multiplier^(attempt-1), capped at MaxDelay.
func delayFor(cfg core.BackoffConfig, attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	factor := math.Pow(cfg.Multiplier, float64(attempt-1))
	delay := time.Duration(float64(cfg.BaseDelay) * factor)
	if delay > cfg.MaxDelay || delay < 0 {
		return cfg.MaxDelay
	}
	return delay
}
