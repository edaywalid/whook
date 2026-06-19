package crypto_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/edaywalid/whook/internal/crypto"
)

func TestSealOpenRoundTrip(t *testing.T) {
	sealer, err := crypto.NewSealer("a strong passphrase")
	if err != nil {
		t.Fatalf("new sealer: %v", err)
	}

	plaintext := []byte("whsec_supersecret")
	sealed, err := sealer.Seal(plaintext)
	if err != nil {
		t.Fatalf("seal: %v", err)
	}
	if bytes.Equal(sealed, plaintext) {
		t.Fatal("sealed output should not equal plaintext")
	}

	opened, err := sealer.Open(sealed)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if !bytes.Equal(opened, plaintext) {
		t.Fatalf("round trip mismatch: got %q", opened)
	}
}

func TestNilSealerRejectsSecrets(t *testing.T) {
	sealer, err := crypto.NewSealer("")
	if err != nil {
		t.Fatalf("new sealer: %v", err)
	}
	if sealer != nil {
		t.Fatal("empty passphrase should yield a nil sealer")
	}

	if _, err := sealer.Seal([]byte("x")); !errors.Is(err, crypto.ErrNoKey) {
		t.Fatalf("expected ErrNoKey, got %v", err)
	}
}

func TestOpenRejectsTamperedCiphertext(t *testing.T) {
	sealer, _ := crypto.NewSealer("key")
	sealed, _ := sealer.Seal([]byte("secret"))
	sealed[len(sealed)-1] ^= 0xff

	if _, err := sealer.Open(sealed); err == nil {
		t.Fatal("expected error opening tampered ciphertext")
	}
}
