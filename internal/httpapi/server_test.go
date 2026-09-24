package httpapi_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/config"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/historystore"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/httpapi"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/interfaceclient"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/webhooksign"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/webhookstore"
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

func TestRecordForwardsToInterface(t *testing.T) {
	var gotPath, gotBody, gotMethod string
	iface := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"recording_id":"rec-1"}`))
	}))
	defer iface.Close()

	srv := &httpapi.Server{
		Cfg:       config.Config{CustomerID: "acme", HMACSecret: "x"},
		Interface: interfaceclient.New(iface.URL, "acme"),
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	startBody := `{"mode":"audio"}`
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/v1/sessions/room-1/record/start", strings.NewReader(startBody))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("start %d %s", res.StatusCode, raw)
	}
	if gotMethod != http.MethodPost || gotPath != "/v1/customers/acme/sessions/room-1/record/start" {
		t.Fatalf("start method=%s path=%q", gotMethod, gotPath)
	}
	if gotBody != startBody {
		t.Fatalf("start body: %q", gotBody)
	}
	if !strings.Contains(string(raw), `"rec-1"`) {
		t.Fatalf("passthrough: %s", raw)
	}

	req, _ = http.NewRequest(http.MethodPost, ts.URL+"/v1/sessions/room-1/record/stop", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if gotPath != "/v1/customers/acme/sessions/room-1/record/stop" {
		t.Fatalf("stop path: %q", gotPath)
	}

	res, err = http.Get(ts.URL + "/v1/sessions/room-1/record")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if gotMethod != http.MethodGet || gotPath != "/v1/customers/acme/sessions/room-1/record" {
		t.Fatalf("get method=%s path=%q", gotMethod, gotPath)
	}
}

func TestRecordPassthroughForbidden(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
	}{
		{"suspended", http.StatusForbidden, `{"error":"customer_suspended"}`},
		{"feature", http.StatusForbidden, `{"error":"feature_denied","feature":"record"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			iface := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/customers/acme/sessions/room-1/record/start" {
					t.Fatalf("path: %q", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer iface.Close()

			srv := &httpapi.Server{
				Cfg:       config.Config{CustomerID: "acme", HMACSecret: "x"},
				Interface: interfaceclient.New(iface.URL, "acme"),
			}
			ts := httptest.NewServer(srv.Handler())
			defer ts.Close()

			req, _ := http.NewRequest(http.MethodPost, ts.URL+"/v1/sessions/room-1/record/start", strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			raw, _ := io.ReadAll(res.Body)
			if res.StatusCode != tc.status {
				t.Fatalf("status %d body %s", res.StatusCode, raw)
			}
			var body map[string]any
			if err := json.Unmarshal(raw, &body); err != nil {
				t.Fatal(err)
			}
			wantErr := "customer_suspended"
			if tc.name == "feature" {
				wantErr = "feature_denied"
			}
			if body["error"] != wantErr {
				t.Fatalf("error: %v", body)
			}
		})
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

func TestSessionsListForwardsToInterface(t *testing.T) {
	var gotPath string
	var deny bool
	iface := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if deny {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"feature_denied","feature":"sessions"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"sessions":[{"session_id":"room-1"}]}`))
	}))
	defer iface.Close()

	srv := &httpapi.Server{
		Cfg:       config.Config{CustomerID: "acme", HMACSecret: "x"},
		Interface: interfaceclient.New(iface.URL, "acme"),
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	res, err := http.Get(ts.URL + "/v1/sessions")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d %s", res.StatusCode, raw)
	}
	if gotPath != "/v1/customers/acme/sessions" {
		t.Fatalf("path: %q", gotPath)
	}
	if !strings.Contains(string(raw), `"room-1"`) {
		t.Fatalf("body: %s", raw)
	}

	deny = true
	res, err = http.Get(ts.URL + "/v1/sessions")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ = io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("deny status %d %s", res.StatusCode, raw)
	}
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	if body["error"] != "feature_denied" {
		t.Fatalf("deny body: %v", body)
	}
}

func TestXylonWebhookIngest(t *testing.T) {
	dir := t.TempDir()
	hist, err := historystore.Open(filepath.Join(dir, "h.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer hist.Close()

	srv := &httpapi.Server{
		Cfg: config.Config{
			CustomerID:    "acme",
			HMACSecret:    "x",
			WebhookSecret: "lab-webhook-secret",
		},
		History: hist,
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	raw := []byte(`{"type":"session.closed","session_id":"room-1","tenant_id":"acme"}`)
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/v1/webhooks/xylon", strings.NewReader(string(raw)))
	req.Header.Set("x-xylon-signature", "sha256=deadbeef")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("bad sig %d", res.StatusCode)
	}

	sig := webhooksign.Sign(raw, "lab-webhook-secret")
	req, _ = http.NewRequest(http.MethodPost, ts.URL+"/v1/webhooks/xylon", strings.NewReader(string(raw)))
	req.Header.Set("x-xylon-signature", sig)
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("good sig %d", res.StatusCode)
	}

	res, err = http.Get(ts.URL + "/v1/webhooks/xylon/latest")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var body map[string]any
	_ = json.NewDecoder(res.Body).Decode(&body)
	if res.StatusCode != http.StatusOK || body["type"] != "session.closed" {
		t.Fatalf("latest: %d %v", res.StatusCode, body)
	}

	res, err = http.Get(ts.URL + "/v1/history/sessions")
	if err != nil {
		t.Fatal(err)
	}
	rawHist, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("history %d %s", res.StatusCode, rawHist)
	}
	if !strings.Contains(string(rawHist), `"room-1"`) {
		t.Fatalf("history body: %s", rawHist)
	}

	res, err = http.Get(ts.URL + "/v1/history/sessions/room-1")
	if err != nil {
		t.Fatal(err)
	}
	rawDetail, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK || !strings.Contains(string(rawDetail), `"session"`) {
		t.Fatalf("detail %d %s", res.StatusCode, rawDetail)
	}

	res, err = http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	rawUI, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK || !strings.Contains(string(rawUI), "html") {
		t.Fatalf("ui %d %s", res.StatusCode, rawUI)
	}
}

func TestXylonWebhookUnconfigured(t *testing.T) {
	srv := &httpapi.Server{
		Cfg: config.Config{CustomerID: "acme", HMACSecret: "x"},
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/v1/webhooks/xylon", strings.NewReader(`{}`))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var body map[string]any
	_ = json.NewDecoder(res.Body).Decode(&body)
	if res.StatusCode != http.StatusServiceUnavailable || body["error"] != "webhook_unconfigured" {
		t.Fatalf("got %d %v", res.StatusCode, body)
	}
}

func TestHistorySurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hist.sqlite")
	hist, err := historystore.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	srv := &httpapi.Server{
		Cfg: config.Config{
			CustomerID:    "acme",
			HMACSecret:    "x",
			WebhookSecret: "lab-webhook-secret",
		},
		History: hist,
	}
	ts := httptest.NewServer(srv.Handler())

	raw := []byte(`{"type":"session.created","session_id":"persist-1","tenant_id":"acme","node":"n1","region":"eu"}`)
	sig := webhooksign.Sign(raw, "lab-webhook-secret")
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/v1/webhooks/xylon", strings.NewReader(string(raw)))
	req.Header.Set("x-xylon-signature", sig)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("ingest %d", res.StatusCode)
	}
	ts.Close()
	_ = hist.Close()

	hist2, err := historystore.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer hist2.Close()
	srv2 := &httpapi.Server{
		Cfg:     config.Config{CustomerID: "acme", HMACSecret: "x"},
		History: hist2,
	}
	ts2 := httptest.NewServer(srv2.Handler())
	defer ts2.Close()

	res, err = http.Get(ts2.URL + "/v1/history/sessions")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK || !strings.Contains(string(body), `"persist-1"`) {
		t.Fatalf("reopen history %d %s", res.StatusCode, body)
	}
}

func TestFleetNodesRefreshAndCache(t *testing.T) {
	var hits int
	iface := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if r.URL.Path != "/v1/customers/acme/fleet/nodes" {
			t.Fatalf("path: %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"nodes":[{"node":"a@host","region":"eu","session_count":1,"participant_count":2,"hub":true,"reachable":true}]}`))
	}))
	defer iface.Close()

	hist, err := historystore.Open(filepath.Join(t.TempDir(), "f.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer hist.Close()

	srv := &httpapi.Server{
		Cfg:       config.Config{CustomerID: "acme", HMACSecret: "x"},
		Interface: interfaceclient.New(iface.URL, "acme"),
		History:   hist,
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	res, err := http.Get(ts.URL + "/v1/fleet/nodes")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK || !strings.Contains(string(raw), `"nodes"`) {
		t.Fatalf("empty cache %d %s", res.StatusCode, raw)
	}
	if hits != 0 {
		t.Fatalf("cache should not hit interface: %d", hits)
	}

	res, err = http.Get(ts.URL + "/v1/fleet/nodes?refresh=1")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ = io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK || !strings.Contains(string(raw), `"a@host"`) {
		t.Fatalf("refresh %d %s", res.StatusCode, raw)
	}
	if hits != 1 {
		t.Fatalf("hits %d", hits)
	}

	res, err = http.Get(ts.URL + "/v1/fleet/summary")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ = io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK || !strings.Contains(string(raw), `"participant_count"`) {
		t.Fatalf("summary %d %s", res.StatusCode, raw)
	}
	if !strings.Contains(string(raw), `"healthy_count"`) {
		t.Fatalf("missing healthy_count: %s", raw)
	}
	if strings.Contains(string(raw), `"sdp"`) || strings.Contains(string(raw), `"media"`) {
		t.Fatalf("media leak: %s", raw)
	}
}

func TestOpsMatchesFleetCounts(t *testing.T) {
	hist, err := historystore.Open(filepath.Join(t.TempDir(), "ops.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer hist.Close()
	if err := hist.ReplaceNodeSnapshots([]historystore.NodeSnapshot{
		{NodeID: "a@host", Region: "eu", SessionCount: 1, ParticipantCount: 2, Hub: true, Reachable: true, PolledAt: 1},
		{NodeID: "b@host", Region: "eu", SessionCount: 0, ParticipantCount: 0, Reachable: true, PolledAt: 1},
	}, 1); err != nil {
		t.Fatal(err)
	}

	srv := &httpapi.Server{
		Cfg:     config.Config{CustomerID: "acme", HMACSecret: "x", ScaleOutParticipants: 100},
		History: hist,
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	fleetSum, err := http.Get(ts.URL + "/v1/fleet/summary")
	if err != nil {
		t.Fatal(err)
	}
	fleetRaw, _ := io.ReadAll(fleetSum.Body)
	fleetSum.Body.Close()
	opsSum, err := http.Get(ts.URL + "/v1/ops/summary")
	if err != nil {
		t.Fatal(err)
	}
	opsRaw, _ := io.ReadAll(opsSum.Body)
	opsSum.Body.Close()
	if fleetSum.StatusCode != http.StatusOK || opsSum.StatusCode != http.StatusOK {
		t.Fatalf("status fleet=%d ops=%d", fleetSum.StatusCode, opsSum.StatusCode)
	}
	if string(fleetRaw) != string(opsRaw) {
		t.Fatalf("summary mismatch\nfleet=%s\nops=%s", fleetRaw, opsRaw)
	}

	fleetNodes, err := http.Get(ts.URL + "/v1/fleet/nodes")
	if err != nil {
		t.Fatal(err)
	}
	fleetNRaw, _ := io.ReadAll(fleetNodes.Body)
	fleetNodes.Body.Close()
	opsNodes, err := http.Get(ts.URL + "/v1/ops/nodes")
	if err != nil {
		t.Fatal(err)
	}
	opsNRaw, _ := io.ReadAll(opsNodes.Body)
	opsNodes.Body.Close()
	if string(fleetNRaw) != string(opsNRaw) {
		t.Fatalf("nodes mismatch\nfleet=%s\nops=%s", fleetNRaw, opsNRaw)
	}
	if !strings.Contains(string(opsNRaw), `"a@host"`) || !strings.Contains(string(opsRaw), `"participant_count"`) {
		t.Fatalf("ops bodies: %s %s", opsRaw, opsNRaw)
	}
	if strings.Contains(string(opsNRaw), `"payload_json"`) || strings.Contains(string(opsNRaw), `"sdp"`) {
		t.Fatalf("ops nodes must be counts/ids only: %s", opsNRaw)
	}
	if !strings.Contains(string(opsRaw), `"healthy_count"`) {
		t.Fatalf("ops summary missing healthy_count: %s", opsRaw)
	}
}

func TestOpsRecommendationsAfterRefresh(t *testing.T) {
	iface := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"nodes":[{"node":"hot@host","region":"eu","session_count":1,"participant_count":5,"hub":true,"reachable":true}]}`))
	}))
	defer iface.Close()

	hist, err := historystore.Open(filepath.Join(t.TempDir(), "rec.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer hist.Close()
	recs := webhookstore.New(8)

	srv := &httpapi.Server{
		Cfg: config.Config{
			CustomerID:           "acme",
			HMACSecret:           "x",
			ScaleOutParticipants: 3,
		},
		Interface:       interfaceclient.New(iface.URL, "acme"),
		History:         hist,
		Recommendations: recs,
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	res, err := http.Get(ts.URL + "/v1/ops/recommendations/latest")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("empty want 404 got %d", res.StatusCode)
	}

	res, err = http.Get(ts.URL + "/v1/ops/nodes?refresh=1")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK || !strings.Contains(string(raw), `"hot@host"`) {
		t.Fatalf("refresh %d %s", res.StatusCode, raw)
	}

	res, err = http.Get(ts.URL + "/v1/ops/recommendations/latest")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var body map[string]any
	_ = json.NewDecoder(res.Body).Decode(&body)
	if res.StatusCode != http.StatusOK || body["type"] != "scale.recommend" {
		t.Fatalf("latest: %d %v", res.StatusCode, body)
	}
	if body["action"] != "scale_out" || body["node"] != "hot@host" {
		t.Fatalf("%v", body)
	}
}

func TestFleetAcceptingForward(t *testing.T) {
	var acceptingPath, acceptingBody, acceptingMethod string
	iface := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/fleet/accepting") {
			acceptingPath = r.URL.Path
			acceptingMethod = r.Method
			b, _ := io.ReadAll(r.Body)
			acceptingBody = string(b)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"node":"a@host","accepting":false}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"nodes":[{"node":"a@host","accepting":false,"reachable":true}]}`))
	}))
	defer iface.Close()

	hist, err := historystore.Open(filepath.Join(t.TempDir(), "acc.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer hist.Close()

	srv := &httpapi.Server{
		Cfg:       config.Config{CustomerID: "acme", HMACSecret: "x", ScaleOutParticipants: 100},
		Interface: interfaceclient.New(iface.URL, "acme"),
		History:   hist,
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/v1/ops/accepting", strings.NewReader(`{"node":"a@host","accepting":false}`))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("%d %s", res.StatusCode, raw)
	}
	if acceptingMethod != http.MethodPost || acceptingPath != "/v1/customers/acme/fleet/accepting" {
		t.Fatalf("method=%s path=%q", acceptingMethod, acceptingPath)
	}
	if acceptingBody != `{"node":"a@host","accepting":false}` {
		t.Fatalf("body %q", acceptingBody)
	}
	nodes, err := hist.ListNodeSnapshots()
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 || nodes[0].Accepting {
		t.Fatalf("refresh after drain: %+v", nodes)
	}
}

