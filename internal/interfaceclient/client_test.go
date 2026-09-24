package interfaceclient_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/interfaceclient"
)

func TestPresenceUsesCustomerPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"participants":{}}`))
	}))
	defer srv.Close()

	c := interfaceclient.New(srv.URL, "acme")
	res, err := c.Presence("room-1")
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/v1/customers/acme/sessions/room-1/presence" {
		t.Fatalf("path: %q", gotPath)
	}
	if res.StatusCode != http.StatusOK || string(res.Body) != `{"participants":{}}` {
		t.Fatalf("res: %+v", res)
	}
}

func TestModeratePostsBody(t *testing.T) {
	var gotPath, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"feature_denied"}`))
	}))
	defer srv.Close()

	c := interfaceclient.New(srv.URL, "acme")
	body := []byte(`{"action":"kick","participant_id":"p2"}`)
	res, err := c.Moderate("room-1", body)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/v1/customers/acme/sessions/room-1/moderate" {
		t.Fatalf("path: %q", gotPath)
	}
	if gotBody != string(body) {
		t.Fatalf("body: %q", gotBody)
	}
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func TestRecordStartStopGet(t *testing.T) {
	var gotPath, gotBody, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"recording_id":"rec-1"}`))
	}))
	defer srv.Close()

	c := interfaceclient.New(srv.URL, "acme")
	body := []byte(`{"mode":"audio"}`)
	res, err := c.RecordStart("room-1", body)
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodPost || gotPath != "/v1/customers/acme/sessions/room-1/record/start" {
		t.Fatalf("start method=%s path=%q", gotMethod, gotPath)
	}
	if gotBody != string(body) {
		t.Fatalf("start body: %q", gotBody)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("start status %d", res.StatusCode)
	}

	_, err = c.RecordStop("room-1", []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/v1/customers/acme/sessions/room-1/record/stop" {
		t.Fatalf("stop path: %q", gotPath)
	}

	_, err = c.RecordGet("room-1")
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodGet || gotPath != "/v1/customers/acme/sessions/room-1/record" {
		t.Fatalf("get method=%s path=%q", gotMethod, gotPath)
	}
}

func TestHistoryQuery(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"messages":[]}`))
	}))
	defer srv.Close()

	c := interfaceclient.New(srv.URL, "acme")
	_, err := c.History("room-1", "chat")
	if err != nil {
		t.Fatal(err)
	}
	if gotQuery != "label=chat" {
		t.Fatalf("query: %q", gotQuery)
	}
}

func TestSessionsList(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"sessions":[]}`))
	}))
	defer srv.Close()

	c := interfaceclient.New(srv.URL, "acme")
	res, err := c.SessionsList()
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/v1/customers/acme/sessions" {
		t.Fatalf("path: %q", gotPath)
	}
	if res.StatusCode != http.StatusOK || string(res.Body) != `{"sessions":[]}` {
		t.Fatalf("res: %+v", res)
	}
}
