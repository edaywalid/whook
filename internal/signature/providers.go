package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// Stripe verifies the Stripe-Signature header. The signed payload is
// "{timestamp}.{body}" and v1 is its HMAC-SHA256 keyed by the signing secret.
type Stripe struct{}

func (Stripe) Verify(secret []byte, req Request) error {
	sig := header(req.Headers, "Stripe-Signature")
	if sig == "" {
		return ErrInvalid
	}

	var timestamp, want string
	for part := range strings.SplitSeq(sig, ",") {
		k, v, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		switch strings.TrimSpace(k) {
		case "t":
			timestamp = v
		case "v1":
			want = v
		}
	}
	if timestamp == "" || want == "" {
		return ErrInvalid
	}

	signed := timestamp + "." + string(req.Body)
	if !hmacEqual(secret, []byte(signed), want) {
		return ErrInvalid
	}
	return nil
}

// GitHub verifies the X-Hub-Signature-256 header, which is "sha256=" followed by
// the hex HMAC-SHA256 of the raw body keyed by the webhook secret.
type GitHub struct{}

func (GitHub) Verify(secret []byte, req Request) error {
	sig := header(req.Headers, "X-Hub-Signature-256")
	hexSig, ok := strings.CutPrefix(sig, "sha256=")
	if !ok || hexSig == "" {
		return ErrInvalid
	}
	if !hmacEqual(secret, req.Body, hexSig) {
		return ErrInvalid
	}
	return nil
}

// hmacEqual reports whether the hex-encoded HMAC-SHA256 of message under key
// equals wantHex, using a constant-time comparison.
func hmacEqual(key, message []byte, wantHex string) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	got := mac.Sum(nil)

	want, err := hex.DecodeString(wantHex)
	if err != nil {
		return false
	}
	return hmac.Equal(got, want)
}
