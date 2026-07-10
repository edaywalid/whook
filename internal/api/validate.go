package api

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/edaywalid/whook/internal/core"
)

// maxPageSize caps how many rows a list endpoint will return in one page, so a
// caller cannot force an unbounded query and result slice with a huge ?limit.
const maxPageSize = 500

// validateURL rejects destination URLs that are not absolute http or https URLs.
// A non-absolute or non-HTTP URL always fails at delivery time, so we reject it
// at the API boundary to give the caller a clear error instead of a silent
// dead-letter later.
func validateURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return errors.New("destination url must be an absolute http or https url")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("destination url scheme must be http or https")
	}
	return nil
}

// validateDestinationInput validates every destination field that is present.
// A non-empty URL must be a valid http/https URL; the backoff and filter, when
// supplied, must be well formed. It suits both create and update: fields left
// unset (empty URL, nil backoff/filter) are skipped so a partial update only
// validates what it changes.
func validateDestinationInput(in destinationInput) error {
	if in.URL != "" {
		if err := validateURL(in.URL); err != nil {
			return err
		}
	}
	if err := validateBackoff(in.Backoff); err != nil {
		return err
	}
	return validateFilterSpec(in.FilterSpec)
}

// validateBackoff rejects per-destination retry overrides that would misbehave
// in the delivery worker: a non-positive MaxAttempts dead-letters on the first
// failure, and non-positive delays or a shrinking multiplier produce a tight
// redelivery loop. A nil override means "use the gateway default" and is fine.
func validateBackoff(b *core.BackoffConfig) error {
	if b == nil {
		return nil
	}
	if b.MaxAttempts < 1 {
		return errors.New("backoff max_attempts must be at least 1")
	}
	if b.BaseDelay <= 0 {
		return errors.New("backoff base_delay must be positive")
	}
	if b.MaxDelay <= 0 {
		return errors.New("backoff max_delay must be positive")
	}
	if b.MaxDelay < b.BaseDelay {
		return errors.New("backoff max_delay must be greater than or equal to base_delay")
	}
	if b.Multiplier < 1 {
		return errors.New("backoff multiplier must be at least 1")
	}
	return nil
}

// validateFilterSpec rejects filters that can never match, so a typo surfaces as
// a 400 at configuration time rather than a destination that silently receives
// nothing. Empty field paths, empty header names, and empty allow-lists are the
// cases the matcher treats as unsatisfiable.
func validateFilterSpec(spec *core.FilterSpec) error {
	if spec == nil {
		return nil
	}
	for _, m := range []map[string]string{spec.BodyEquals, spec.BodyPrefix, spec.BodyNot} {
		for path := range m {
			if strings.TrimSpace(path) == "" {
				return errors.New("filter body paths must not be empty")
			}
		}
	}
	for _, path := range spec.BodyExists {
		if strings.TrimSpace(path) == "" {
			return errors.New("filter body paths must not be empty")
		}
	}
	for name := range spec.HeaderEquals {
		if strings.TrimSpace(name) == "" {
			return errors.New("filter header names must not be empty")
		}
	}
	for path, allowed := range spec.BodyIn {
		if strings.TrimSpace(path) == "" {
			return errors.New("filter body paths must not be empty")
		}
		if len(allowed) == 0 {
			return fmt.Errorf("filter body_in for %q must list at least one value", path)
		}
	}
	return nil
}

// validateStatus rejects an event-status filter that is not a known status. An
// empty value means "no status filter" and is allowed.
func validateStatus(status string) error {
	switch core.EventStatus(status) {
	case "", core.StatusReceived, core.StatusRejected:
		return nil
	default:
		return fmt.Errorf("status must be %q or %q", core.StatusReceived, core.StatusRejected)
	}
}

// clampLimit floors a page size to the default when non-positive and caps it at
// maxPageSize so a caller cannot request an unbounded result set.
func clampLimit(n, def int) int {
	if n <= 0 {
		n = def
	}
	if n > maxPageSize {
		return maxPageSize
	}
	return n
}

// clampOffset floors a negative offset to zero. A negative offset is a client
// error and, on Postgres, an outright query failure.
func clampOffset(n int) int {
	if n < 0 {
		return 0
	}
	return n
}
