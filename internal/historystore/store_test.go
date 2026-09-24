package historystore_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/historystore"
)

func applyJSON(t *testing.T, s *historystore.Store, obj map[string]any) {
	t.Helper()
	raw, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	var ev map[string]any
	if err := json.Unmarshal(raw, &ev); err != nil {
		t.Fatal(err)
	}
	if err := s.ApplyWebhook(raw, ev); err != nil {
		t.Fatal(err)
	}
}

func TestApplyWebhookLifecycleAndReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.sqlite")
	s, err := historystore.Open(path)
	if err != nil {
		t.Fatal(err)
	}

	applyJSON(t, s, map[string]any{
		"type": "session.created", "tenant_id": "acme", "session_id": "room-1",
		"node": "n1@host", "region": "eu-west",
	})
	applyJSON(t, s, map[string]any{
		"type": "participant.joined", "tenant_id": "acme", "session_id": "room-1",
		"participant_id": "alice", "node": "n1@host", "region": "eu-west",
	})
	applyJSON(t, s, map[string]any{
		"type": "recording.completed", "tenant_id": "acme", "session_id": "room-1",
		"recording_id": "rec-1", "sequence": 1, "mode": "audio",
		"started_at": 100, "ended_at": 200,
	})
	applyJSON(t, s, map[string]any{
		"type": "session.closed", "tenant_id": "acme", "session_id": "room-1",
		"node": "n1@host", "region": "eu-west",
	})

	// Duplicate payload hash is a no-op success.
	dup := []byte(`{"type":"session.closed","tenant_id":"acme","session_id":"room-1","node":"n1@host","region":"eu-west"}`)
	var dupEv map[string]any
	_ = json.Unmarshal(dup, &dupEv)
	if err := s.ApplyWebhook(dup, dupEv); err != nil {
		t.Fatal(err)
	}

	list, err := s.ListSessions("acme", 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("list: %+v", list)
	}
	if list[0].SessionID != "room-1" || list[0].ClosedAt == nil {
		t.Fatalf("session: %+v", list[0])
	}
	if list[0].LastNode != "n1@host" || list[0].LastRegion != "eu-west" {
		t.Fatalf("node/region: %+v", list[0])
	}

	detail, err := s.GetSession("acme", "room-1")
	if err != nil {
		t.Fatal(err)
	}
	if detail == nil {
		t.Fatal("expected detail")
	}
	if len(detail.Participants) != 1 || detail.Participants[0].ParticipantID != "alice" {
		t.Fatalf("participants: %+v", detail.Participants)
	}
	if len(detail.Recordings) != 1 || detail.Recordings[0].RecordingID != "rec-1" {
		t.Fatalf("recordings: %+v", detail.Recordings)
	}
	if detail.Recordings[0].Status != "completed" {
		t.Fatalf("status: %q", detail.Recordings[0].Status)
	}

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	s2, err := historystore.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()

	list2, err := s2.ListSessions("acme", 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(list2) != 1 || list2[0].SessionID != "room-1" || list2[0].ClosedAt == nil {
		t.Fatalf("reopen list: %+v", list2)
	}
}

func TestListSessionsNewestFirst(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.sqlite")
	s, err := historystore.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	applyJSON(t, s, map[string]any{
		"type": "session.created", "tenant_id": "acme", "session_id": "old",
	})
	applyJSON(t, s, map[string]any{
		"type": "session.created", "tenant_id": "acme", "session_id": "new",
	})

	list, err := s.ListSessions("acme", 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("len %d", len(list))
	}
	if list[0].SessionID != "new" || list[1].SessionID != "old" {
		t.Fatalf("order: %+v", list)
	}
}

func TestRecordingFailed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.sqlite")
	s, err := historystore.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	applyJSON(t, s, map[string]any{
		"type": "recording.failed", "tenant_id": "acme", "session_id": "room-2",
		"recording_id": "rec-x", "mode": "audio", "error": "boom",
	})
	detail, err := s.GetSession("acme", "room-2")
	if err != nil {
		t.Fatal(err)
	}
	if detail == nil || len(detail.Recordings) != 1 || detail.Recordings[0].Status != "failed" {
		t.Fatalf("%+v", detail)
	}
}

func TestNodeSnapshotsReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.sqlite")
	s, err := historystore.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	err = s.ReplaceNodeSnapshots([]historystore.NodeSnapshot{
		{NodeID: "a@host", Region: "eu", SessionCount: 1, ParticipantCount: 3, Hub: true, Reachable: true},
		{NodeID: "b@host", Region: "eu", SessionCount: 0, ParticipantCount: 0, Reachable: true},
	}, 100)
	if err != nil {
		t.Fatal(err)
	}
	_ = s.Close()

	s2, err := historystore.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()
	nodes, err := s2.ListNodeSnapshots()
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 2 {
		t.Fatalf("%+v", nodes)
	}
	sum := historystore.FleetSummaryFromSnapshots(nodes)
	if sum["node_count"] != 2 || sum["participant_count"] != int64(3) {
		t.Fatalf("%v", sum)
	}
	if sum["hub_node"] != "a@host" {
		t.Fatalf("%v", sum)
	}
	if sum["healthy_count"] != 2 || sum["unhealthy_count"] != 0 {
		t.Fatalf("health: %v", sum)
	}
}
