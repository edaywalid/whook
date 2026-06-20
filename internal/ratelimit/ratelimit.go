// Package ratelimit provides a per-key token-bucket limiter, used to cap the
// ingest rate per source so one provider cannot overwhelm the gateway.
package ratelimit

import (
	"math"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// bucket pairs a token-bucket with the last time it was used, so stale entries
// can be evicted to prevent the key map from growing without bound.
type bucket struct {
	lim      *rate.Limiter
	lastSeen time.Time
}

// Limiter caps requests per key with an independent token bucket for each key.
// A non-positive rate disables limiting, and a nil *Limiter allows everything,
// so callers can construct one unconditionally.
type Limiter struct {
	mu    sync.Mutex
	byKey map[string]*bucket
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
		byKey: map[string]*bucket{},
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
	b, ok := l.byKey[key]
	if !ok {
		b = &bucket{lim: rate.NewLimiter(l.limit, l.burst)}
		l.byKey[key] = b
	}
	b.lastSeen = time.Now()
	l.mu.Unlock()
	return b.lim.Allow()
}

// EvictStale removes per-key buckets that have not been used since before
// olderThan, and returns the number of entries removed. Call it periodically
// (e.g. via a time.Ticker in the same goroutine that created the limiter) to
// prevent an attacker who sends requests under many different source names from
// growing the key map without bound.
//
// EvictStale is a no-op on a nil or disabled Limiter.
func (l *Limiter) EvictStale(olderThan time.Time) int {
	if l == nil || l.limit <= 0 {
		return 0
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	removed := 0
	for key, b := range l.byKey {
		if b.lastSeen.Before(olderThan) {
			delete(l.byKey, key)
			removed++
		}
	}
	return removed
}
