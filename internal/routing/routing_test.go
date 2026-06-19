package routing_test

import (
	"testing"

	"github.com/edaywalid/whook/internal/core"
	"github.com/edaywalid/whook/internal/routing"
)

func TestResolve(t *testing.T) {
	event := core.Event{
		Body: []byte(`{"type":"order.created"}`),
	}

	dest := func(id string, enabled bool, spec *core.FilterSpec) core.Destination {
		return core.Destination{ID: id, Enabled: enabled, FilterSpec: spec}
	}

	tests := []struct {
		name         string
		destinations []core.Destination
		expectedIDs  []string
	}{
		{
			name:         "single destination no filter",
			destinations: []core.Destination{dest("a", true, nil)},
			expectedIDs:  []string{"a"},
		},
		{
			name: "fan out to multiple destinations",
			destinations: []core.Destination{
				dest("a", true, nil),
				dest("b", true, nil),
				dest("c", true, nil),
			},
			expectedIDs: []string{"a", "b", "c"},
		},
		{
			name: "filter includes and excludes",
			destinations: []core.Destination{
				dest("match", true, &core.FilterSpec{BodyEquals: map[string]string{"type": "order.created"}}),
				dest("nomatch", true, &core.FilterSpec{BodyEquals: map[string]string{"type": "order.deleted"}}),
			},
			expectedIDs: []string{"match"},
		},
		{
			name:         "disabled destination excluded",
			destinations: []core.Destination{dest("a", false, nil), dest("b", true, nil)},
			expectedIDs:  []string{"b"},
		},
		{
			name: "no match yields empty set",
			destinations: []core.Destination{
				dest("a", true, &core.FilterSpec{BodyEquals: map[string]string{"type": "nope"}}),
			},
			expectedIDs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := routing.Resolve(event, tt.destinations)
			if len(got) != len(tt.expectedIDs) {
				t.Fatalf("got %d destinations, expected %d", len(got), len(tt.expectedIDs))
			}
			for i, d := range got {
				if d.ID != tt.expectedIDs[i] {
					t.Errorf("position %d: got %q, expected %q", i, d.ID, tt.expectedIDs[i])
				}
			}
		})
	}
}
