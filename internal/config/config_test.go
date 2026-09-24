package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/config"
)

func TestLoadOK(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cfg.json")
	body := `{
	  "customer_id":"acme",
	  "hmac_secret":"s3cret",
	  "listen_addr":"127.0.0.1:8092"
	}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.CustomerID != "acme" || cfg.HMACSecret != "s3cret" {
		t.Fatalf("%+v", cfg)
	}
	if cfg.DefaultTTLSec != 300 {
		t.Fatalf("ttl %d", cfg.DefaultTTLSec)
	}
	if cfg.InterfaceBaseURL != "http://127.0.0.1:8093" {
		t.Fatalf("interface: %q", cfg.InterfaceBaseURL)
	}
	if cfg.SQLitePath != "./data/admin-history.sqlite" {
		t.Fatalf("sqlite: %q", cfg.SQLitePath)
	}
	if cfg.FleetPollSec != 5 {
		t.Fatalf("fleet_poll: %d", cfg.FleetPollSec)
	}
	if cfg.ScaleOutParticipants != 100 {
		t.Fatalf("scale_out: %d", cfg.ScaleOutParticipants)
	}
}

func TestLoadRejectsEmptySecret(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cfg.json")
	_ = os.WriteFile(path, []byte(`{"customer_id":"acme","hmac_secret":""}`), 0o600)
	if _, err := config.Load(path); err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadRejectsNonLoopback(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cfg.json")
	_ = os.WriteFile(path, []byte(`{"customer_id":"acme","hmac_secret":"x","listen_addr":"0.0.0.0:8092"}`), 0o600)
	if _, err := config.Load(path); err == nil {
		t.Fatal("expected error")
	}
}
