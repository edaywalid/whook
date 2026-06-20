// Package filter evaluates a destination's routing filter against an event.
package filter

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/edaywalid/whook/internal/core"
)

// Matches reports whether the event satisfies every clause in the spec. A nil
// or empty spec matches all events. Body clauses require the body to parse as a
// JSON object; a non-JSON body fails any body clause.
func Matches(spec *core.FilterSpec, e core.Event) bool {
	if spec == nil {
		return true
	}

	for name, want := range spec.HeaderEquals {
		if !headerEquals(e.Headers, name, want) {
			return false
		}
	}

	if !needsBody(spec) {
		return true
	}
	body, ok := parseJSONObject(e.Body)
	if !ok {
		return false
	}
	return matchesBody(spec, body)
}

func matchesBody(spec *core.FilterSpec, body map[string]any) bool {
	for path, want := range spec.BodyEquals {
		if got, found := lookup(body, path); !found || valueString(got) != want {
			return false
		}
	}
	for _, path := range spec.BodyExists {
		if _, found := lookup(body, path); !found {
			return false
		}
	}
	for path, allowed := range spec.BodyIn {
		got, found := lookup(body, path)
		if !found || !slices.Contains(allowed, valueString(got)) {
			return false
		}
	}
	for path, prefix := range spec.BodyPrefix {
		got, found := lookup(body, path)
		if !found || !strings.HasPrefix(valueString(got), prefix) {
			return false
		}
	}
	for path, notWant := range spec.BodyNot {
		// A field that equals the value fails. A missing field passes, since it
		// does not equal anything.
		if got, found := lookup(body, path); found && valueString(got) == notWant {
			return false
		}
	}
	return true
}

func needsBody(spec *core.FilterSpec) bool {
	return len(spec.BodyEquals) > 0 || len(spec.BodyExists) > 0 ||
		len(spec.BodyIn) > 0 || len(spec.BodyPrefix) > 0 || len(spec.BodyNot) > 0
}

func headerEquals(headers map[string][]string, name, want string) bool {
	for k, v := range headers {
		if strings.EqualFold(k, name) && len(v) > 0 {
			return v[0] == want
		}
	}
	return false
}

func parseJSONObject(body []byte) (map[string]any, bool) {
	var obj map[string]any
	if err := json.Unmarshal(body, &obj); err != nil {
		return nil, false
	}
	return obj, true
}

// lookup walks a dot-separated path through nested JSON objects.
func lookup(obj map[string]any, path string) (any, bool) {
	parts := strings.Split(path, ".")
	var current any = obj
	for _, part := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = m[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

// valueString renders a JSON leaf for comparison against the expected string.
// Strings compare directly; numbers and booleans use their natural rendering.
func valueString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case bool:
		return fmt.Sprintf("%t", val)
	case float64:
		return fmt.Sprintf("%g", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}
