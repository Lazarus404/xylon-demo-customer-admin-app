package fleetpoll_test

import (
	"path/filepath"
	"testing"

	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/fleetpoll"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/historystore"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/webhookstore"
)

func TestParseNodesBody(t *testing.T) {
	raw := []byte(`{"nodes":[{"node":"a@host","region":"eu","session_count":1,"participant_count":2,"hub":true,"cohost_hub_count":1,"reachable":true,"accepting":false}]}`)
	nodes, err := fleetpoll.ParseNodesBody(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 {
		t.Fatalf("%+v", nodes)
	}
	n := nodes[0]
	if n.NodeID != "a@host" || n.Region != "eu" || n.SessionCount != 1 || n.ParticipantCount != 2 {
		t.Fatalf("%+v", n)
	}
	if !n.Hub || n.CohostHubCount != 1 || !n.Reachable || n.Accepting {
		t.Fatalf("%+v", n)
	}
}

func TestMaybeRecommendDedupesIdentical(t *testing.T) {
	hist, err := historystore.Open(filepath.Join(t.TempDir(), "rec.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer hist.Close()
	if err := hist.ReplaceNodeSnapshots([]historystore.NodeSnapshot{
		{NodeID: "hot@host", ParticipantCount: 5, Reachable: true},
	}, 1); err != nil {
		t.Fatal(err)
	}
	recs := webhookstore.New(8)
	fleetpoll.MaybeRecommend(hist, recs, 3)
	fleetpoll.MaybeRecommend(hist, recs, 3)
	ev, ok := recs.Latest()
	if !ok || ev["action"] != "scale_out" {
		t.Fatalf("%v ok=%v", ev, ok)
	}
	// Second call must not append a duplicate advisories entry.
	// Store has no Len; re-Latest and ensure only one distinct Add by
	// forcing a different count then checking Latest updates once.
	if err := hist.ReplaceNodeSnapshots([]historystore.NodeSnapshot{
		{NodeID: "hot@host", ParticipantCount: 9, Reachable: true},
	}, 2); err != nil {
		t.Fatal(err)
	}
	fleetpoll.MaybeRecommend(hist, recs, 3)
	ev2, ok := recs.Latest()
	if !ok || ev2["participant_count"] != int64(9) {
		t.Fatalf("want updated advisory got %v", ev2)
	}
	fleetpoll.MaybeRecommend(hist, recs, 3)
	ev3, _ := recs.Latest()
	if ev3["at"] != ev2["at"] {
		t.Fatalf("identical should not re-Add: %v vs %v", ev2, ev3)
	}
}
