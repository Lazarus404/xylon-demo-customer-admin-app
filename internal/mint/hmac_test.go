package mint_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/mint"
)

const golden = "MjAwMDAwMDAwMDphY21lOnJvb20tMTphbGljZTpbXQ.NRd33yoIPCaZQYezD5p7UFMEYTifGngFm-32o0L5upM"

func TestGoldenVector(t *testing.T) {
	tok, err := mint.Token("test-secret", mint.Params{
		TenantID:      "acme",
		SessionID:     "room-1",
		ParticipantID: "alice",
		Capabilities:  []any{},
		Now:           2000000000 - 300,
		TTLSec:        300,
	})
	if err != nil {
		t.Fatal(err)
	}
	if tok != golden {
		t.Fatalf("got %q\nwant %q", tok, golden)
	}
}

func TestMACRoundTrip(t *testing.T) {
	tok, err := mint.Token("test-secret", mint.Params{
		TenantID:      "acme",
		SessionID:     "room-1",
		ParticipantID: "alice",
		Now:           2000000000 - 300,
		TTLSec:        300,
	})
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(tok, ".")
	if len(parts) != 2 {
		t.Fatal(tok)
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatal(err)
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	mac := hmac.New(sha256.New, []byte("test-secret"))
	_, _ = mac.Write(payload)
	if !hmac.Equal(mac.Sum(nil), sig) {
		t.Fatal("mac mismatch")
	}
}

func TestEmptySecret(t *testing.T) {
	if _, err := mint.Token("", mint.Params{
		TenantID: "a", SessionID: "s", ParticipantID: "p",
	}); err == nil {
		t.Fatal("expected error")
	}
}
