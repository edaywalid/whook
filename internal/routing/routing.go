// Package routing resolves a captured event to the destinations that should
// receive it. It is pure logic with no I/O.
package routing

import (
	"github.com/edaywalid/whook/internal/core"
	"github.com/edaywalid/whook/internal/filter"
)

// Resolve returns the subset of destinations that are enabled and whose filter
// matches the event. Order is preserved. The result is never nil.
func Resolve(e core.Event, destinations []core.Destination) []core.Destination {
	matched := make([]core.Destination, 0, len(destinations))
	for _, d := range destinations {
		if !d.Enabled {
			continue
		}
		if filter.Matches(d.FilterSpec, e) {
			matched = append(matched, d)
		}
	}
	return matched
}
