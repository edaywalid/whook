// Package crypto seals provider signing secrets at rest with AES-256-GCM.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
)

// ErrNoKey indicates a secret operation was attempted without a configured key.
var ErrNoKey = errors.New("crypto: no secret key configured")

// Sealer encrypts and decrypts secrets. A nil Sealer is valid and returns
// ErrNoKey from Seal/Open, so callers can construct one unconditionally and
// fail only when a secret is actually used.
type Sealer struct {
	aead cipher.AEAD
}

// NewSealer derives an AES-256 key from the passphrase. An empty passphrase
// yields a nil Sealer that rejects secret operations.
func NewSealer(passphrase string) (*Sealer, error) {
	if passphrase == "" {
		return nil, nil
	}
	key := sha256.Sum256([]byte(passphrase))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating gcm: %w", err)
	}
	return &Sealer{aead: aead}, nil
}

// Seal returns nonce||ciphertext for the plaintext.
func (s *Sealer) Seal(plaintext []byte) ([]byte, error) {
	if s == nil {
		return nil, ErrNoKey
	}
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("reading nonce: %w", err)
	}
	return s.aead.Seal(nonce, nonce, plaintext, nil), nil
}

// Open reverses Seal.
func (s *Sealer) Open(sealed []byte) ([]byte, error) {
	if s == nil {
		return nil, ErrNoKey
	}
	size := s.aead.NonceSize()
	if len(sealed) < size {
		return nil, errors.New("crypto: ciphertext too short")
	}
	nonce, ciphertext := sealed[:size], sealed[size:]
	plaintext, err := s.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypting secret: %w", err)
	}
	return plaintext, nil
}
