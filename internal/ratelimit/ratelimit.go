// Package ratelimit provides a per-key token-bucket limiter, used to cap the
// ingest rate per source so one provider cannot overwhelm the gateway.
package ratelimit

import (
	"math"
	"sync"

	"golang.org/x/time/rate"
)

// Limiter caps requests per key with an independent token bucket for each key.
// A non-positive rate disables limiting, and a nil *Limiter allows everything,
// so callers can construct one unconditionally.
type Limiter struct {
	mu    sync.Mutex
	byKey map[string]*rate.Limiter
	limit rate.Limit
	burst int
}

// New builds a limiter allowing perSecond requests per key with the given burst.
// A burst below 1 defaults to one second of capacity. A non-positive rate makes
// every Allow return true.
func New(perSecond float64, burst int) *Limiter {
	if burst < 1 {
		burst = max(1, int(math.Ceil(perSecond)))
	}
	return &Limiter{
		byKey: map[string]*rate.Limiter{},
		limit: rate.Limit(perSecond),
		burst: burst,
	}
}

// Allow reports whether a request for key may proceed now.
func (l *Limiter) Allow(key string) bool {
	if l == nil || l.limit <= 0 {
		return true
	}
	l.mu.Lock()
	lim, ok := l.byKey[key]
	if !ok {
		lim = rate.NewLimiter(l.limit, l.burst)
		l.byKey[key] = lim
	}
	l.mu.Unlock()
	return lim.Allow()
}
