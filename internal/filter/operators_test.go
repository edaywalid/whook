package filter_test

import (
	"testing"

	"github.com/edaywalid/whook/internal/core"
	"github.com/edaywalid/whook/internal/filter"
)

func TestMatchesOperators(t *testing.T) {
	event := core.Event{
		Body: []byte(`{"type":"payment.succeeded","data":{"object":{"amount":4900,"livemode":true}}}`),
	}

	tests := []struct {
		name     string
		spec     *core.FilterSpec
		expected bool
	}{
		{"exists matches present path", &core.FilterSpec{BodyExists: []string{"data.object.amount"}}, true},
		{"exists fails on missing path", &core.FilterSpec{BodyExists: []string{"data.object.missing"}}, false},
		{"in matches a listed value", &core.FilterSpec{BodyIn: map[string][]string{"type": {"payment.succeeded", "payment.failed"}}}, true},
		{"in fails when not listed", &core.FilterSpec{BodyIn: map[string][]string{"type": {"order.created"}}}, false},
		{"prefix matches", &core.FilterSpec{BodyPrefix: map[string]string{"type": "payment."}}, true},
		{"prefix fails", &core.FilterSpec{BodyPrefix: map[string]string{"type": "order."}}, false},
		{"not passes when different", &core.FilterSpec{BodyNot: map[string]string{"type": "payment.failed"}}, true},
		{"not fails when equal", &core.FilterSpec{BodyNot: map[string]string{"type": "payment.succeeded"}}, false},
		{"not passes when missing", &core.FilterSpec{BodyNot: map[string]string{"data.object.missing": "x"}}, true},
		{
			name: "operators combine with AND",
			spec: &core.FilterSpec{
				BodyPrefix: map[string]string{"type": "payment."},
				BodyExists: []string{"data.object.livemode"},
				BodyNot:    map[string]string{"type": "payment.failed"},
			},
			expected: true,
		},
		{
			name: "one failing operator fails the set",
			spec: &core.FilterSpec{
				BodyPrefix: map[string]string{"type": "payment."},
				BodyExists: []string{"data.object.missing"},
			},
			expected: false,
		},
		{"operator on non-json body fails", &core.FilterSpec{BodyExists: []string{"type"}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := event
			if tt.name == "operator on non-json body fails" {
				ev = core.Event{Body: []byte("not json")}
			}
			if got := filter.Matches(tt.spec, ev); got != tt.expected {
				t.Errorf("Matches() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
