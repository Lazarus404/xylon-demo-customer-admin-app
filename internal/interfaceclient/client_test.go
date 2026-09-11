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
