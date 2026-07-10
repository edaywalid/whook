package ratelimit

import (
	"context"
	"testing"
	"time"
)

// TestRunEvictionPrunesIdleKeys proves the background loop actually removes idle
// buckets, not just that EvictStale would. idleTTL is a nanosecond so every key
// touched before a tick is stale by the next one.
func TestRunEvictionPrunesIdleKeys(t *testing.T) {
	l := New(1000, 1)
	l.Allow("a")
	l.Allow("b")
	if got := l.len(); got != 2 {
		t.Fatalf("expected 2 keys before eviction, got %d", got)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go l.RunEviction(ctx, time.Millisecond, time.Nanosecond)

	deadline := time.Now().Add(2 * time.Second)
	for l.len() != 0 {
		if time.Now().After(deadline) {
			t.Fatalf("keys were not evicted, still %d after 2s", l.len())
		}
		time.Sleep(2 * time.Millisecond)
	}
}

func TestRunEvictionReturnsOnDisabledLimiter(t *testing.T) {
	l := New(0, 0) // disabled
	// Must return immediately rather than spin, even with a tiny interval.
	done := make(chan struct{})
	go func() {
		l.RunEviction(context.Background(), time.Millisecond, time.Millisecond)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("RunEviction should return immediately on a disabled limiter")
	}
}
