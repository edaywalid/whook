package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
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

// Shopify verifies the X-Shopify-Hmac-Sha256 header, which is the base64-encoded
// HMAC-SHA256 of the raw body keyed by the app's webhook secret.
type Shopify struct{}

func (Shopify) Verify(secret []byte, req Request) error {
	sig := header(req.Headers, "X-Shopify-Hmac-Sha256")
	if sig == "" {
		return ErrInvalid
	}
	if !hmacEqualBase64(secret, req.Body, sig) {
		return ErrInvalid
	}
	return nil
}

// Slack verifies the X-Slack-Signature header. The signed payload is
// "v0:{timestamp}:{body}" and the signature is "v0=" followed by its hex
// HMAC-SHA256, with the timestamp read from X-Slack-Request-Timestamp.
type Slack struct{}

func (Slack) Verify(secret []byte, req Request) error {
	hexSig, ok := strings.CutPrefix(header(req.Headers, "X-Slack-Signature"), "v0=")
	if !ok || hexSig == "" {
		return ErrInvalid
	}
	ts := header(req.Headers, "X-Slack-Request-Timestamp")
	if ts == "" {
		return ErrInvalid
	}
	signed := "v0:" + ts + ":" + string(req.Body)
	if !hmacEqual(secret, []byte(signed), hexSig) {
		return ErrInvalid
	}
	return nil
}

// Linear verifies the Linear-Signature header, which is the hex HMAC-SHA256 of
// the raw body keyed by the webhook signing secret.
type Linear struct{}

func (Linear) Verify(secret []byte, req Request) error {
	sig := header(req.Headers, "Linear-Signature")
	if sig == "" {
		return ErrInvalid
	}
	if !hmacEqual(secret, req.Body, sig) {
		return ErrInvalid
	}
	return nil
}

// hmacEqual reports whether the hex-encoded HMAC-SHA256 of message under key
// equals wantHex, using a constant-time comparison.
func hmacEqual(key, message []byte, wantHex string) bool {
	want, err := hex.DecodeString(wantHex)
	if err != nil {
		return false
	}
	return hmacEqualRaw(key, message, want)
}

// hmacEqualBase64 is hmacEqual for a base64-encoded (standard encoding) digest.
func hmacEqualBase64(key, message []byte, wantB64 string) bool {
	want, err := base64.StdEncoding.DecodeString(wantB64)
	if err != nil {
		return false
	}
	return hmacEqualRaw(key, message, want)
}

// hmacEqualRaw reports whether the HMAC-SHA256 of message under key equals the
// raw digest want, using a constant-time comparison.
func hmacEqualRaw(key, message, want []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	return hmac.Equal(mac.Sum(nil), want)
}
