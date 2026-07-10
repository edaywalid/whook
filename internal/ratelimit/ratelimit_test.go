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

// Eviction is observable through Allow: once a bucket is removed, the next call
// creates a fresh one with a full burst, so a previously exhausted key passes
// again. The cutoff is passed explicitly, so these tests need no sleeps.

func TestEvictStaleRemovesStaleBuckets(t *testing.T) {
	l := ratelimit.New(0.01, 1) // burst of 1, very slow refill

	if !l.Allow("stripe") {
		t.Fatal("first request should pass")
	}
	if l.Allow("stripe") {
		t.Fatal("burst exhausted: second request should be blocked")
	}

	if removed := l.EvictStale(time.Now().Add(time.Hour)); removed != 1 {
		t.Fatalf("expected 1 eviction, got %d", removed)
	}

	if !l.Allow("stripe") {
		t.Fatal("after eviction a fresh bucket should allow the request")
	}
}

func TestEvictStaleKeepsRecentBuckets(t *testing.T) {
	l := ratelimit.New(0.01, 1)

	l.Allow("stripe") // stamps lastSeen to roughly now

	if removed := l.EvictStale(time.Now().Add(-time.Hour)); removed != 0 {
		t.Fatalf("expected 0 evictions, got %d", removed)
	}
	if l.Allow("stripe") {
		t.Fatal("non-evicted bucket should still be rate-limited")
	}
}

func TestEvictStaleOnlyRemovesMatchingKeys(t *testing.T) {
	l := ratelimit.New(0.01, 1)

	l.Allow("old")
	cutoff := time.Now()
	l.Allow("new")

	if removed := l.EvictStale(cutoff); removed != 1 {
		t.Fatalf("expected 1 eviction (old), got %d", removed)
	}
	if !l.Allow("old") {
		t.Fatal("evicted key should get a fresh bucket")
	}
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
	l := ratelimit.New(0, 0) // rate <= 0 disables limiting
	l.Allow("x")
	if removed := l.EvictStale(time.Now().Add(time.Hour)); removed != 0 {
		t.Fatalf("disabled limiter should remove nothing, got %d", removed)
	}
}
