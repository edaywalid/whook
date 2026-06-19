package signature_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/edaywalid/whook/internal/signature"
)

func hmacHex(secret, message []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write(message)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestStripeVerify(t *testing.T) {
	secret := []byte("whsec_test")
	body := []byte(`{"id":"evt_1","type":"payment.succeeded"}`)
	ts := "1700000000"
	good := hmacHex(secret, []byte(ts+"."+string(body)))

	tests := []struct {
		name    string
		sig     string
		body    []byte
		wantErr bool
	}{
		{name: "valid signature", sig: "t=" + ts + ",v1=" + good, body: body},
		{name: "modified body rejected", sig: "t=" + ts + ",v1=" + good, body: []byte(`{"tampered":true}`), wantErr: true},
		{name: "missing header rejected", sig: "", body: body, wantErr: true},
		{name: "malformed header rejected", sig: "garbage", body: body, wantErr: true},
		{name: "wrong signature rejected", sig: "t=" + ts + ",v1=deadbeef", body: body, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := signature.Request{
				Headers: map[string][]string{"Stripe-Signature": {tt.sig}},
				Body:    tt.body,
			}
			if tt.sig == "" {
				req.Headers = map[string][]string{}
			}
			err := signature.Stripe{}.Verify(secret, req)
			assertErr(t, err, tt.wantErr)
		})
	}
}

func TestGitHubVerify(t *testing.T) {
	secret := []byte("ghsecret")
	body := []byte(`{"action":"opened"}`)
	good := "sha256=" + hmacHex(secret, body)

	tests := []struct {
		name    string
		sig     string
		body    []byte
		wantErr bool
	}{
		{name: "valid signature", sig: good, body: body},
		{name: "modified body rejected", sig: good, body: []byte(`{"action":"closed"}`), wantErr: true},
		{name: "missing prefix rejected", sig: hmacHex(secret, body), body: body, wantErr: true},
		{name: "empty signature rejected", sig: "", body: body, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := signature.Request{
				Headers: map[string][]string{"X-Hub-Signature-256": {tt.sig}},
				Body:    tt.body,
			}
			err := signature.GitHub{}.Verify(secret, req)
			assertErr(t, err, tt.wantErr)
		})
	}
}

func TestRegistry(t *testing.T) {
	reg := signature.NewRegistry()

	t.Run("empty provider accepts", func(t *testing.T) {
		if err := reg.Verify("", nil, signature.Request{}); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("unknown provider errors", func(t *testing.T) {
		err := reg.Verify("twilio", nil, signature.Request{})
		if !errors.Is(err, signature.ErrUnknownProvider) {
			t.Fatalf("expected ErrUnknownProvider, got %v", err)
		}
	})
}

func assertErr(t *testing.T, err error, wantErr bool) {
	t.Helper()
	if wantErr && err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !wantErr && err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}
