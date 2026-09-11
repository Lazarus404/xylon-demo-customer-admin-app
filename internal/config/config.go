package config

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

// Config for the demo mint service.
type Config struct {
	CustomerID       string   `json:"customer_id"`
	HMACSecret       string   `json:"hmac_secret"`
	ListenAddr       string   `json:"listen_addr"`
	InterfaceBaseURL string   `json:"interface_base_url"`
	DefaultTTLSec    int      `json:"default_ttl_sec"`
	MaxTTLSec        int      `json:"max_ttl_sec"`
	CORSOrigins      []string `json:"cors_origins"`
}

// Load reads JSON config; customer_id and hmac_secret are required.
func Load(path string) (Config, error) {
	if path == "" {
		return Config{}, fmt.Errorf("config path required")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	cfg := Config{
		ListenAddr:       "127.0.0.1:8092",
		InterfaceBaseURL: "http://127.0.0.1:8093",
		DefaultTTLSec:    300,
		MaxTTLSec:        3600,
		CORSOrigins: []string{
			"http://127.0.0.1:8080",
			"http://127.0.0.1:5173",
		},
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = "127.0.0.1:8092"
	}
	if cfg.InterfaceBaseURL == "" {
		cfg.InterfaceBaseURL = "http://127.0.0.1:8093"
	}
	if cfg.DefaultTTLSec < 1 {
		cfg.DefaultTTLSec = 300
	}
	if cfg.MaxTTLSec < 1 {
		cfg.MaxTTLSec = 3600
	}
	if cfg.CustomerID == "" {
		return Config{}, fmt.Errorf("customer_id required")
	}
	if cfg.HMACSecret == "" {
		return Config{}, fmt.Errorf("hmac_secret required")
	}
	if err := requireLoopbackListen(cfg.ListenAddr); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func requireLoopbackListen(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("listen_addr: %w", err)
	}
	host = strings.TrimSpace(host)
	switch strings.ToLower(host) {
	case "", "0.0.0.0", "::", "[::]":
		return fmt.Errorf("listen_addr must be loopback (got %q)", addr)
	case "127.0.0.1", "localhost", "::1":
		return nil
	}
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		return nil
	}
	return fmt.Errorf("listen_addr must be loopback (got %q)", addr)
}
