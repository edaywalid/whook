package filter_test

import (
	"testing"

	"github.com/edaywalid/whook/internal/core"
	"github.com/edaywalid/whook/internal/filter"
)

func TestMatches(t *testing.T) {
	event := core.Event{
		Headers: map[string][]string{
			"Content-Type":   {"application/json"},
			"X-Github-Event": {"push"},
		},
		Body: []byte(`{"type":"payment.succeeded","data":{"object":{"amount":4900,"livemode":true}}}`),
	}

	tests := []struct {
		name     string
		spec     *core.FilterSpec
		event    core.Event
		expected bool
	}{
		{
			name:     "nil spec matches all",
			spec:     nil,
			event:    event,
			expected: true,
		},
		{
			name:     "empty spec matches all",
			spec:     &core.FilterSpec{},
			event:    event,
			expected: true,
		},
		{
			name:     "body field equality match",
			spec:     &core.FilterSpec{BodyEquals: map[string]string{"type": "payment.succeeded"}},
			event:    event,
			expected: true,
		},
		{
			name:     "body field equality mismatch",
			spec:     &core.FilterSpec{BodyEquals: map[string]string{"type": "payment.failed"}},
			event:    event,
			expected: false,
		},
		{
			name:     "nested json path",
			spec:     &core.FilterSpec{BodyEquals: map[string]string{"data.object.amount": "4900"}},
			event:    event,
			expected: true,
		},
		{
			name:     "nested boolean value",
			spec:     &core.FilterSpec{BodyEquals: map[string]string{"data.object.livemode": "true"}},
			event:    event,
			expected: true,
		},
		{
			name:     "missing path does not match",
			spec:     &core.FilterSpec{BodyEquals: map[string]string{"data.object.missing": "x"}},
			event:    event,
			expected: false,
		},
		{
			name:     "header equality case-insensitive name",
			spec:     &core.FilterSpec{HeaderEquals: map[string]string{"x-github-event": "push"}},
			event:    event,
			expected: true,
		},
		{
			name:     "header equality mismatch",
			spec:     &core.FilterSpec{HeaderEquals: map[string]string{"X-Github-Event": "pull_request"}},
			event:    event,
			expected: false,
		},
		{
			name: "multiple clauses all must match",
			spec: &core.FilterSpec{
				BodyEquals:   map[string]string{"type": "payment.succeeded"},
				HeaderEquals: map[string]string{"X-Github-Event": "push"},
			},
			event:    event,
			expected: true,
		},
		{
			name: "multiple clauses one fails",
			spec: &core.FilterSpec{
				BodyEquals:   map[string]string{"type": "payment.succeeded"},
				HeaderEquals: map[string]string{"X-Github-Event": "pull_request"},
			},
			event:    event,
			expected: false,
		},
		{
			name:     "non-json body fails body clause",
			spec:     &core.FilterSpec{BodyEquals: map[string]string{"type": "x"}},
			event:    core.Event{Body: []byte("not json")},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filter.Matches(tt.spec, tt.event); got != tt.expected {
				t.Errorf("Matches() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
