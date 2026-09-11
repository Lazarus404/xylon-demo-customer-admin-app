package httpapi_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/config"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/httpapi"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/interfaceclient"
)

func TestHealthzTokenCORS(t *testing.T) {
	srv := &httpapi.Server{Cfg: config.Config{
		CustomerID:    "acme",
		HMACSecret:    "test-secret",
		DefaultTTLSec: 300,
		MaxTTLSec:     3600,
		CORSOrigins:   []string{"http://127.0.0.1:8080"},
	}}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	res, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("healthz %d", res.StatusCode)
	}

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/v1/token", strings.NewReader(
		`{"session_id":"room-1","participant_id":"alice"}`,
	))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://127.0.0.1:8080")
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("token %d", res.StatusCode)
	}
	if res.Header.Get("Access-Control-Allow-Origin") != "http://127.0.0.1:8080" {
		t.Fatalf("cors: %q", res.Header.Get("Access-Control-Allow-Origin"))
	}
	var body map[string]any
	_ = json.NewDecoder(res.Body).Decode(&body)
	tok, _ := body["token"].(string)
	if tok == "" || !strings.Contains(tok, ".") {
		t.Fatalf("token: %v", body)
	}

	req, _ = http.NewRequest(http.MethodPost, ts.URL+"/v1/token", strings.NewReader(`{"participant_id":"alice"}`))
	req.Header.Set("Content-Type", "application/json")
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}

	req, _ = http.NewRequest(http.MethodOptions, ts.URL+"/v1/token", nil)
	req.Header.Set("Origin", "http://127.0.0.1:8080")
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.Header.Get("Access-Control-Allow-Origin") != "http://127.0.0.1:8080" {
		t.Fatalf("preflight cors: %q", res.Header.Get("Access-Control-Allow-Origin"))
	}
}

func TestSessionAdminForwardsToInterface(t *testing.T) {
	var gotPath string
	iface := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"participants":{"p1":{}}}`))
	}))
	defer iface.Close()

	srv := &httpapi.Server{
		Cfg: config.Config{
			CustomerID:    "acme",
			HMACSecret:    "test-secret",
			DefaultTTLSec: 300,
			MaxTTLSec:     3600,
		},
		Interface: interfaceclient.New(iface.URL, "acme"),
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	res, err := http.Get(ts.URL + "/v1/sessions/room-1/presence")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d body %s", res.StatusCode, raw)
	}
	if gotPath != "/v1/customers/acme/sessions/room-1/presence" {
		t.Fatalf("path: %q", gotPath)
	}
	if !strings.Contains(string(raw), `"p1"`) {
		t.Fatalf("body: %s", raw)
	}
}

func TestSessionAdminPassthroughForbidden(t *testing.T) {
	iface := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"customer_suspended"}`))
	}))
	defer iface.Close()

	srv := &httpapi.Server{
		Cfg:       config.Config{CustomerID: "acme", HMACSecret: "x"},
		Interface: interfaceclient.New(iface.URL, "acme"),
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/v1/sessions/room-1/moderate",
		strings.NewReader(`{"action":"kick","participant_id":"p2"}`))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var body map[string]any
	_ = json.NewDecoder(res.Body).Decode(&body)
	if res.StatusCode != http.StatusForbidden || body["error"] != "customer_suspended" {
		t.Fatalf("got %d %v", res.StatusCode, body)
	}
}

func TestSessionAdminInterfaceUnreachable(t *testing.T) {
	srv := &httpapi.Server{
		Cfg:       config.Config{CustomerID: "acme", HMACSecret: "x"},
		Interface: interfaceclient.New("http://127.0.0.1:1", "acme"),
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	res, err := http.Get(ts.URL + "/v1/sessions/room-1/presence")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var body map[string]any
	_ = json.NewDecoder(res.Body).Decode(&body)
	if res.StatusCode != http.StatusBadGateway || body["error"] != "interface_unreachable" {
		t.Fatalf("got %d %v", res.StatusCode, body)
	}
}
