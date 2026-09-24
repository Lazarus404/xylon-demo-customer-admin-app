// Package opsrecommend evaluates advisory scale.recommend events from fleet snapshots.
package opsrecommend

import (
	"time"

	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/historystore"
)

// Evaluate returns a scale_out advisory when any reachable node meets threshold.
// When several exceed, picks highest participant_count (tie: lexicographic node).
func Evaluate(nodes []historystore.NodeSnapshot, threshold int) (map[string]any, bool) {
	if threshold < 1 {
		return nil, false
	}
	var best *historystore.NodeSnapshot
	for i := range nodes {
		n := &nodes[i]
		if !n.Reachable {
			continue
		}
		if n.ParticipantCount < int64(threshold) {
			continue
		}
		if best == nil {
			best = n
			continue
		}
		if n.ParticipantCount > best.ParticipantCount {
			best = n
			continue
		}
		if n.ParticipantCount == best.ParticipantCount && n.NodeID < best.NodeID {
			best = n
		}
	}
	if best == nil {
		return nil, false
	}
	return map[string]any{
		"type":              "scale.recommend",
		"action":            "scale_out",
		"reason":            "participant_count_threshold",
		"threshold":         threshold,
		"node":              best.NodeID,
		"participant_count": best.ParticipantCount,
		"at":                time.Now().Unix(),
	}, true
}
