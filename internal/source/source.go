// Package source manages registered inbound sources and their signing secrets.
package source

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/edaywalid/whook/internal/core"
	"github.com/edaywalid/whook/internal/crypto"
	"github.com/edaywalid/whook/internal/id"
	"github.com/edaywalid/whook/internal/signature"
	"github.com/edaywalid/whook/internal/storage"
)

// ErrUnknownVerifier means the requested verifier is not registered.
var ErrUnknownVerifier = errors.New("source: unknown verifier")

// Service is the source registry.
type Service struct {
	store     storage.Store
	sealer    *crypto.Sealer
	verifiers *signature.Registry
	now       func() time.Time
}

// NewService wires the registry.
func NewService(store storage.Store, sealer *crypto.Sealer, verifiers *signature.Registry) *Service {
	return &Service{
		store:     store,
		sealer:    sealer,
		verifiers: verifiers,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

// Input describes a source to register or update. Secret is the plaintext
// signing secret; an empty Secret leaves an existing secret unchanged on update
// and means "no secret" on register.
type Input struct {
	Name        string
	Verifier    string
	Secret      string
	DedupHeader string
}

// Register creates a new source.
func (s *Service) Register(ctx context.Context, in Input) (core.Source, error) {
	if in.Name == "" {
		return core.Source{}, errors.New("source: name is required")
	}
	if err := s.checkVerifier(in.Verifier); err != nil {
		return core.Source{}, err
	}

	secret, err := s.sealSecret(in.Secret)
	if err != nil {
		return core.Source{}, err
	}

	src := core.Source{
		ID:          id.New("src"),
		Name:        in.Name,
		Verifier:    in.Verifier,
		DedupHeader: in.DedupHeader,
		HasSecret:   secret != nil,
		CreatedAt:   s.now(),
	}
	if err := s.store.CreateSource(ctx, src, secret); err != nil {
		return core.Source{}, err
	}
	return src, nil
}

// Update changes an existing source. A non-empty Secret replaces the stored one.
func (s *Service) Update(ctx context.Context, in Input) (core.Source, error) {
	if err := s.checkVerifier(in.Verifier); err != nil {
		return core.Source{}, err
	}

	existing, err := s.store.GetSource(ctx, in.Name)
	if err != nil {
		return core.Source{}, err
	}

	secret, err := s.sealSecret(in.Secret)
	if err != nil {
		return core.Source{}, err
	}
	if secret == nil {
		// Keep the current secret when none was supplied.
		secret, err = s.store.SourceSecret(ctx, in.Name)
		if err != nil {
			return core.Source{}, err
		}
	}

	existing.Verifier = in.Verifier
	existing.DedupHeader = in.DedupHeader
	existing.HasSecret = secret != nil
	if err := s.store.UpdateSource(ctx, existing, secret); err != nil {
		return core.Source{}, err
	}
	return existing, nil
}

// Get returns a source by name.
func (s *Service) Get(ctx context.Context, name string) (core.Source, error) {
	return s.store.GetSource(ctx, name)
}

// List returns all sources.
func (s *Service) List(ctx context.Context) ([]core.Source, error) {
	return s.store.ListSources(ctx)
}

// Secret returns the decrypted signing secret for a source, or nil when none is
// configured.
func (s *Service) Secret(ctx context.Context, name string) ([]byte, error) {
	sealed, err := s.store.SourceSecret(ctx, name)
	if err != nil {
		return nil, err
	}
	if len(sealed) == 0 {
		return nil, nil
	}
	return s.sealer.Open(sealed)
}

func (s *Service) sealSecret(plaintext string) ([]byte, error) {
	if plaintext == "" {
		return nil, nil
	}
	sealed, err := s.sealer.Seal([]byte(plaintext))
	if err != nil {
		return nil, fmt.Errorf("sealing secret: %w", err)
	}
	return sealed, nil
}

func (s *Service) checkVerifier(name string) error {
	if name == "" {
		return nil
	}
	if !s.verifiers.Has(name) {
		return fmt.Errorf("%w: %q", ErrUnknownVerifier, name)
	}
	return nil
}
