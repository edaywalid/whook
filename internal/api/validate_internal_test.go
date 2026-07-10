package api

import (
	"testing"
	"time"

	"github.com/edaywalid/whook/internal/core"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		url string
		ok  bool
	}{
		{"http://example.com/hook", true},
		{"https://example.com/hook", true},
		{"https://user:pass@example.com:8443/hook?x=1", true},
		{"/just/a/path", false},
		{"example.com/hook", false},
		{"ftp://example.com/hook", false},
		{"http:///hook", false},
		{"https://", false},
		{"", false},
		{"javascript:alert(1)", false},
	}
	for _, tt := range tests {
		err := validateURL(tt.url)
		if tt.ok && err != nil {
			t.Errorf("validateURL(%q) = %v, want nil", tt.url, err)
		}
		if !tt.ok && err == nil {
			t.Errorf("validateURL(%q) = nil, want error", tt.url)
		}
	}
}

func TestValidateBackoff(t *testing.T) {
	good := &core.BackoffConfig{BaseDelay: time.Second, Multiplier: 2, MaxDelay: time.Minute, MaxAttempts: 5}
	if err := validateBackoff(good); err != nil {
		t.Errorf("valid backoff rejected: %v", err)
	}
	if err := validateBackoff(nil); err != nil {
		t.Errorf("nil backoff should be allowed (use default): %v", err)
	}

	bad := map[string]*core.BackoffConfig{
		"zero attempts":        {BaseDelay: time.Second, Multiplier: 2, MaxDelay: time.Minute, MaxAttempts: 0},
		"negative attempts":    {BaseDelay: time.Second, Multiplier: 2, MaxDelay: time.Minute, MaxAttempts: -1},
		"zero base delay":      {BaseDelay: 0, Multiplier: 2, MaxDelay: time.Minute, MaxAttempts: 5},
		"zero max delay":       {BaseDelay: time.Second, Multiplier: 2, MaxDelay: 0, MaxAttempts: 5},
		"max below base":       {BaseDelay: time.Minute, Multiplier: 2, MaxDelay: time.Second, MaxAttempts: 5},
		"shrinking multiplier": {BaseDelay: time.Second, Multiplier: 0.5, MaxDelay: time.Minute, MaxAttempts: 5},
	}
	for name, cfg := range bad {
		if err := validateBackoff(cfg); err == nil {
			t.Errorf("%s: expected error, got nil", name)
		}
	}
}

func TestValidateFilterSpec(t *testing.T) {
	if err := validateFilterSpec(nil); err != nil {
		t.Errorf("nil spec should be allowed: %v", err)
	}
	good := &core.FilterSpec{
		BodyEquals:   map[string]string{"data.status": "paid"},
		HeaderEquals: map[string]string{"X-Kind": "invoice"},
		BodyIn:       map[string][]string{"type": {"a", "b"}},
	}
	if err := validateFilterSpec(good); err != nil {
		t.Errorf("valid spec rejected: %v", err)
	}

	bad := map[string]*core.FilterSpec{
		"empty body path":    {BodyEquals: map[string]string{"": "x"}},
		"empty exists path":  {BodyExists: []string{""}},
		"empty header name":  {HeaderEquals: map[string]string{"": "x"}},
		"empty body_in list": {BodyIn: map[string][]string{"type": {}}},
	}
	for name, spec := range bad {
		if err := validateFilterSpec(spec); err == nil {
			t.Errorf("%s: expected error, got nil", name)
		}
	}
}

func TestValidateStatus(t *testing.T) {
	for _, ok := range []string{"", "received", "rejected"} {
		if err := validateStatus(ok); err != nil {
			t.Errorf("validateStatus(%q) = %v, want nil", ok, err)
		}
	}
	for _, bad := range []string{"deleted", "RECEIVED", "pending", "x"} {
		if err := validateStatus(bad); err == nil {
			t.Errorf("validateStatus(%q) = nil, want error", bad)
		}
	}
}

func TestClampLimit(t *testing.T) {
	tests := []struct{ in, def, want int }{
		{0, 50, 50},
		{-5, 50, 50},
		{10, 50, 10},
		{maxPageSize, 50, maxPageSize},
		{maxPageSize + 1, 50, maxPageSize},
		{1_000_000, 50, maxPageSize},
	}
	for _, tt := range tests {
		if got := clampLimit(tt.in, tt.def); got != tt.want {
			t.Errorf("clampLimit(%d, %d) = %d, want %d", tt.in, tt.def, got, tt.want)
		}
	}
}

func TestClampOffset(t *testing.T) {
	if got := clampOffset(-1); got != 0 {
		t.Errorf("clampOffset(-1) = %d, want 0", got)
	}
	if got := clampOffset(25); got != 25 {
		t.Errorf("clampOffset(25) = %d, want 25", got)
	}
}
