package webhooksign_test

import (
	"testing"

	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/webhooksign"
)

func TestSignRoundTrip(t *testing.T) {
	body := []byte(`{"type":"session.created"}`)
	sig := webhooksign.Sign(body, "lab-webhook-secret")
	if !webhooksign.Valid(body, "lab-webhook-secret", sig) {
		t.Fatal("expected valid")
	}
	if webhooksign.Valid(body, "wrong", sig) {
		t.Fatal("wrong secret must fail")
	}
}
