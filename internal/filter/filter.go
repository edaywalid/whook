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
		if !anyValue(body, path, func(got any) bool { return valueString(got) == want }) {
			return false
		}
	}
	for _, path := range spec.BodyExists {
		if !anyValue(body, path, func(any) bool { return true }) {
			return false
		}
	}
	for path, allowed := range spec.BodyIn {
		if !anyValue(body, path, func(got any) bool { return slices.Contains(allowed, valueString(got)) }) {
			return false
		}
	}
	for path, prefix := range spec.BodyPrefix {
		if !anyValue(body, path, func(got any) bool { return strings.HasPrefix(valueString(got), prefix) }) {
			return false
		}
	}
	for path, notWant := range spec.BodyNot {
		// A field that equals the value fails. A missing field passes, since it
		// does not equal anything.
		if anyValue(body, path, func(got any) bool { return valueString(got) == notWant }) {
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

// anyValue reports whether pred holds for any value a dot-separated path
// resolves to. A "*" segment fans out over every element of an array, so one
// path can resolve to several values; without one it resolves to at most one.
func anyValue(obj map[string]any, path string, pred func(any) bool) bool {
	return anyValueAt(obj, strings.Split(path, "."), pred)
}

func anyValueAt(node any, parts []string, pred func(any) bool) bool {
	if len(parts) == 0 {
		return pred(node)
	}
	part, rest := parts[0], parts[1:]
	if items, ok := node.([]any); ok && part == "*" {
		for _, item := range items {
			if anyValueAt(item, rest, pred) {
				return true
			}
		}
		return false
	}
	m, ok := node.(map[string]any)
	if !ok {
		return false
	}
	child, ok := m[part]
	if !ok {
		return false
	}
	return anyValueAt(child, rest, pred)
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
