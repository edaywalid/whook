// Package signature verifies provider webhook signatures. Verifiers are
// registered by provider name and selected per source.
package signature

import "errors"

// ErrInvalid means the request signature did not verify.
var ErrInvalid = errors.New("signature: verification failed")

// ErrUnknownProvider means no verifier is registered under the given name.
var ErrUnknownProvider = errors.New("signature: unknown provider")

// Request is the part of an inbound request a verifier inspects.
type Request struct {
	Headers map[string][]string
	Body    []byte
}

// Verifier checks a request against a shared secret. It returns nil when the
// signature is valid and ErrInvalid otherwise.
type Verifier interface {
	Verify(secret []byte, req Request) error
}

// Registry maps provider names to verifiers.
type Registry struct {
	verifiers map[string]Verifier
}

// NewRegistry returns a registry preloaded with the built-in verifiers.
func NewRegistry() *Registry {
	return &Registry{
		verifiers: map[string]Verifier{
			"stripe": Stripe{},
			"github": GitHub{},
		},
	}
}

// Register adds or replaces a verifier.
func (r *Registry) Register(name string, v Verifier) {
	r.verifiers[name] = v
}

// Has reports whether a verifier is registered under the given name.
func (r *Registry) Has(name string) bool {
	_, ok := r.verifiers[name]
	return ok
}

// Verify runs the named verifier. An empty provider name means the source opted
// out of verification, so the request is accepted.
func (r *Registry) Verify(provider string, secret []byte, req Request) error {
	if provider == "" {
		return nil
	}
	v, ok := r.verifiers[provider]
	if !ok {
		return ErrUnknownProvider
	}
	return v.Verify(secret, req)
}

// header returns the first value of a header by case-insensitive name.
func header(headers map[string][]string, name string) string {
	for k, v := range headers {
		if len(v) > 0 && equalFold(k, name) {
			return v[0]
		}
	}
	return ""
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range len(a) {
		ca, cb := a[i], b[i]
		if 'A' <= ca && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if 'A' <= cb && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
