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

func TestMatchesArrayWildcard(t *testing.T) {
	event := core.Event{
		Body: []byte(`{
			"action": "assigned",
			"data": {
				"assignees": [
					{"id": "u_1", "login": "ada"},
					{"id": "u_2", "login": "grace", "teams": [{"slug": "core"}, {"slug": "infra"}]}
				],
				"labels": ["bug", "urgent"],
				"reviewers": [],
				"meta": {"*": "literal"}
			}
		}`),
	}

	tests := []struct {
		name     string
		spec     *core.FilterSpec
		expected bool
	}{
		{
			name:     "equals matches when any element matches",
			spec:     &core.FilterSpec{BodyEquals: map[string]string{"data.assignees.*.id": "u_2"}},
			expected: true,
		},
		{
			name:     "equals fails when no element matches",
			spec:     &core.FilterSpec{BodyEquals: map[string]string{"data.assignees.*.id": "u_9"}},
			expected: false,
		},
		{
			name:     "trailing wildcard matches scalar elements",
			spec:     &core.FilterSpec{BodyEquals: map[string]string{"data.labels.*": "urgent"}},
			expected: true,
		},
		{
			name:     "nested wildcards",
			spec:     &core.FilterSpec{BodyEquals: map[string]string{"data.assignees.*.teams.*.slug": "infra"}},
			expected: true,
		},
		{
			name:     "empty array never matches equals",
			spec:     &core.FilterSpec{BodyEquals: map[string]string{"data.reviewers.*.id": "u_1"}},
			expected: false,
		},
		{
			name:     "wildcard over a non-array does not match",
			spec:     &core.FilterSpec{BodyEquals: map[string]string{"action.*": "assigned"}},
			expected: false,
		},
		{
			name:     "wildcard on an object is a literal key",
			spec:     &core.FilterSpec{BodyEquals: map[string]string{"data.meta.*": "literal"}},
			expected: true,
		},
		{
			name:     "exists matches when any element has the field",
			spec:     &core.FilterSpec{BodyExists: []string{"data.assignees.*.teams"}},
			expected: true,
		},
		{
			name:     "exists fails when no element has the field",
			spec:     &core.FilterSpec{BodyExists: []string{"data.assignees.*.email"}},
			expected: false,
		},
		{
			name:     "exists fails on an empty array",
			spec:     &core.FilterSpec{BodyExists: []string{"data.reviewers.*"}},
			expected: false,
		},
		{
			name:     "in matches when any element is in the list",
			spec:     &core.FilterSpec{BodyIn: map[string][]string{"data.assignees.*.login": {"linus", "grace"}}},
			expected: true,
		},
		{
			name:     "in fails when no element is in the list",
			spec:     &core.FilterSpec{BodyIn: map[string][]string{"data.assignees.*.login": {"linus", "ken"}}},
			expected: false,
		},
		{
			name:     "prefix matches when any element has the prefix",
			spec:     &core.FilterSpec{BodyPrefix: map[string]string{"data.labels.*": "urg"}},
			expected: true,
		},
		{
			name:     "prefix fails when no element has the prefix",
			spec:     &core.FilterSpec{BodyPrefix: map[string]string{"data.labels.*": "feat"}},
			expected: false,
		},
		{
			name:     "not fails when any element equals the value",
			spec:     &core.FilterSpec{BodyNot: map[string]string{"data.assignees.*.id": "u_2"}},
			expected: false,
		},
		{
			name:     "not passes when no element equals the value",
			spec:     &core.FilterSpec{BodyNot: map[string]string{"data.assignees.*.id": "u_9"}},
			expected: true,
		},
		{
			name:     "not passes on an empty array",
			spec:     &core.FilterSpec{BodyNot: map[string]string{"data.reviewers.*.id": "u_1"}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filter.Matches(tt.spec, event); got != tt.expected {
				t.Errorf("Matches() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
