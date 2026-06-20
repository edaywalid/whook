package ratelimit_test

import (
	"testing"
	"time"

	"github.com/edaywalid/whook/internal/ratelimit"
)

func TestAllowsBurstThenBlocks(t *testing.T) {
	l := ratelimit.New(0.01, 2) // slow refill, burst of 2

	first := l.Allow("stripe")
	second := l.Allow("stripe")
	if !first || !second {
		t.Fatal("the first two requests should pass within the burst")
	}
	if l.Allow("stripe") {
		t.Fatal("the third request should be blocked")
	}
}

func TestKeysAreIndependent(t *testing.T) {
	l := ratelimit.New(0.01, 1)

	if !l.Allow("stripe") {
		t.Fatal("first stripe request should pass")
	}
	if l.Allow("stripe") {
		t.Fatal("second stripe request should be blocked")
	}
	if !l.Allow("github") {
		t.Fatal("a different source should have its own bucket")
	}
}

func TestZeroRateDisablesLimiting(t *testing.T) {
	l := ratelimit.New(0, 0)
	for range 100 {
		if !l.Allow("x") {
			t.Fatal("a zero rate should never block")
		}
	}
}

func TestNilLimiterAllows(t *testing.T) {
	var l *ratelimit.Limiter
	if !l.Allow("x") {
		t.Fatal("a nil limiter should allow everything")
	}
}

// TestEvictStaleRemovesStaleBuckets verifies that a bucket not used since
// before the cutoff is removed. Eviction is observable through Allow: after
// the bucket is gone a fresh one is created with a full burst, so the next
// call returns true even though the old bucket was exhausted.
func TestEvictStaleRemovesStaleBuckets(t *testing.T) {
	l := ratelimit.New(0.01, 1) // burst of 1, very slow refill

	// Exhaust the burst for "stripe".
	if !l.Allow("stripe") {
		t.Fatal("first request should pass")
	}
	if l.Allow("stripe") {
		t.Fatal("burst exhausted: second request should be blocked")
	}

	// Evict everything last seen before an hour from now (all buckets qualify).
	removed := l.EvictStale(time.Now().Add(time.Hour))
	if removed != 1 {
		t.Fatalf("expected 1 eviction, got %d", removed)
	}

	// A fresh bucket is created on the next Allow, so the burst is full again.
	if !l.Allow("stripe") {
		t.Fatal("after eviction a fresh bucket should allow the request")
	}
}

// TestEvictStaleKeepsRecentBuckets verifies that a bucket used after the
// cutoff is left untouched.
func TestEvictStaleKeepsRecentBuckets(t *testing.T) {
	l := ratelimit.New(0.01, 1)

	l.Allow("stripe") // stamps lastSeen to roughly now

	// Cutoff is one hour in the past: "stripe" was seen after that, so it stays.
	removed := l.EvictStale(time.Now().Add(-time.Hour))
	if removed != 0 {
		t.Fatalf("expected 0 evictions, got %d", removed)
	}

	// The bucket is still present and the burst is still exhausted.
	if l.Allow("stripe") {
		t.Fatal("non-evicted bucket should still be rate-limited")
	}
}

// TestEvictStaleOnlyRemovesMatchingKeys verifies that eviction is selective:
// keys used after the cutoff are retained while old ones are removed.
func TestEvictStaleOnlyRemovesMatchingKeys(t *testing.T) {
	l := ratelimit.New(0.01, 1)

	past := time.Now().Add(-2 * time.Hour)
	future := time.Now().Add(time.Hour)

	// Touch "old" first so its lastSeen is before the cutoff we will use.
	l.Allow("old")
	// Advance the cutoff past "old" but before "new".
	cutoff := time.Now()
	// Touch "new" after the cutoff.
	l.Allow("new")

	_ = past
	_ = future
	removed := l.EvictStale(cutoff)
	if removed != 1 {
		t.Fatalf("expected 1 eviction (old), got %d", removed)
	}

	// "old" is gone: a fresh bucket lets the next request through.
	if !l.Allow("old") {
		t.Fatal("evicted key should get a fresh bucket")
	}
	// "new" is still present and its burst is exhausted.
	if l.Allow("new") {
		t.Fatal("retained key should still be rate-limited")
	}
}

func TestEvictStaleNilLimiterIsNoop(t *testing.T) {
	var l *ratelimit.Limiter
	if removed := l.EvictStale(time.Now()); removed != 0 {
		t.Fatalf("nil limiter should remove nothing, got %d", removed)
	}
}

func TestEvictStaleDisabledLimiterIsNoop(t *testing.T) {
	l := ratelimit.New(0, 0) // rate ≤ 0 disables limiting
	l.Allow("x")
	if removed := l.EvictStale(time.Now().Add(time.Hour)); removed != 0 {
		t.Fatalf("disabled limiter should remove nothing, got %d", removed)
	}
}
