package opsrecommend_test

import (
	"testing"

	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/historystore"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/opsrecommend"
)

func TestEvaluateScaleOut(t *testing.T) {
	nodes := []historystore.NodeSnapshot{
		{NodeID: "a@host", ParticipantCount: 1, Reachable: true},
		{NodeID: "b@host", ParticipantCount: 3, Reachable: true},
	}
	ev, ok := opsrecommend.Evaluate(nodes, 2)
	if !ok {
		t.Fatal("expected advisory")
	}
	if ev["type"] != "scale.recommend" || ev["action"] != "scale_out" {
		t.Fatalf("%v", ev)
	}
	if ev["node"] != "b@host" || ev["participant_count"] != int64(3) {
		t.Fatalf("%v", ev)
	}
	if ev["threshold"] != 2 {
		t.Fatalf("%v", ev)
	}
}

func TestEvaluateBelowThreshold(t *testing.T) {
	nodes := []historystore.NodeSnapshot{
		{NodeID: "a@host", ParticipantCount: 1, Reachable: true},
	}
	_, ok := opsrecommend.Evaluate(nodes, 2)
	if ok {
		t.Fatal("expected no advisory")
	}
}

func TestEvaluateIgnoresUnreachable(t *testing.T) {
	nodes := []historystore.NodeSnapshot{
		{NodeID: "a@host", ParticipantCount: 99, Reachable: false},
	}
	_, ok := opsrecommend.Evaluate(nodes, 2)
	if ok {
		t.Fatal("unreachable must not trigger")
	}
}

func TestEvaluateTieBreakLexNode(t *testing.T) {
	nodes := []historystore.NodeSnapshot{
		{NodeID: "b@host", ParticipantCount: 5, Reachable: true},
		{NodeID: "a@host", ParticipantCount: 5, Reachable: true},
	}
	ev, ok := opsrecommend.Evaluate(nodes, 2)
	if !ok || ev["node"] != "a@host" {
		t.Fatalf("%v ok=%v", ev, ok)
	}
}
