// Package ratelimit provides a per-key token-bucket limiter, used to cap the
// ingest rate per source so one provider cannot overwhelm the gateway.
package ratelimit

import (
	"context"
	"math"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// bucket pairs a token-bucket with the last time it was used, so stale entries
// can be evicted to keep the key map from growing without bound.
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
	now   func() time.Time
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
		now:   func() time.Time { return time.Now() },
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
	b.lastSeen = l.now()
	l.mu.Unlock()
	return b.lim.Allow()
}

// EvictStale removes per-key buckets not used since before olderThan and returns
// the number of entries removed. It is a no-op on a nil or disabled limiter.
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

// len reports the current number of tracked keys. Used in tests.
func (l *Limiter) len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.byKey)
}

// RunEviction prunes buckets idle for longer than idleTTL on every interval tick
// until ctx is cancelled. It returns immediately on a nil or disabled limiter, so
// callers can start it unconditionally. Run it in its own goroutine.
func (l *Limiter) RunEviction(ctx context.Context, interval, idleTTL time.Duration) {
	if l == nil || l.limit <= 0 || interval <= 0 {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			l.EvictStale(l.now().Add(-idleTTL))
		}
	}
}
