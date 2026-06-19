package source_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/edaywalid/whook/internal/crypto"
	"github.com/edaywalid/whook/internal/signature"
	"github.com/edaywalid/whook/internal/source"
	"github.com/edaywalid/whook/internal/storage/sqlite"
)

func newService(t *testing.T) *source.Service {
	t.Helper()
	dsn := "file:" + filepath.Join(t.TempDir(), "test.db")
	st, err := sqlite.Open(context.Background(), dsn)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	sealer, err := crypto.NewSealer("test-key")
	if err != nil {
		t.Fatalf("new sealer: %v", err)
	}
	return source.NewService(st, sealer, signature.NewRegistry())
}

func TestRegisterAndGet(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)

	src, err := svc.Register(ctx, source.Input{
		Name:        "stripe",
		Verifier:    "stripe",
		Secret:      "whsec_test",
		DedupHeader: "Stripe-Id",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if !src.HasSecret {
		t.Fatal("expected HasSecret true")
	}

	got, err := svc.Get(ctx, "stripe")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Verifier != "stripe" || got.DedupHeader != "Stripe-Id" {
		t.Fatalf("unexpected source: %+v", got)
	}
}

func TestRegisterUnknownVerifier(t *testing.T) {
	svc := newService(t)
	_, err := svc.Register(context.Background(), source.Input{Name: "x", Verifier: "twilio"})
	if !errors.Is(err, source.ErrUnknownVerifier) {
		t.Fatalf("expected ErrUnknownVerifier, got %v", err)
	}
}

func TestSecretRoundTrip(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)

	if _, err := svc.Register(ctx, source.Input{Name: "stripe", Verifier: "stripe", Secret: "whsec_test"}); err != nil {
		t.Fatalf("register: %v", err)
	}
	secret, err := svc.Secret(ctx, "stripe")
	if err != nil {
		t.Fatalf("secret: %v", err)
	}
	if string(secret) != "whsec_test" {
		t.Fatalf("expected decrypted secret, got %q", secret)
	}
}

func TestUpdateKeepsSecretWhenOmitted(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)

	if _, err := svc.Register(ctx, source.Input{Name: "stripe", Verifier: "stripe", Secret: "whsec_test"}); err != nil {
		t.Fatalf("register: %v", err)
	}
	if _, err := svc.Update(ctx, source.Input{Name: "stripe", Verifier: "stripe", DedupHeader: "Stripe-Id"}); err != nil {
		t.Fatalf("update: %v", err)
	}

	secret, err := svc.Secret(ctx, "stripe")
	if err != nil {
		t.Fatalf("secret: %v", err)
	}
	if string(secret) != "whsec_test" {
		t.Fatalf("update should keep existing secret, got %q", secret)
	}
}
