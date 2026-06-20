package ratelimit_test

import (
	"testing"

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
