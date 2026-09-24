// Package fleetpoll periodically pulls interface fleet nodes into SQLite.
package fleetpoll

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/historystore"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/interfaceclient"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/opsrecommend"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/webhookstore"
)

// Poller refreshes node_snapshots on an interval.
type Poller struct {
	Interface            *interfaceclient.Client
	History              *historystore.Store
	Recommendations      *webhookstore.Store
	ScaleOutParticipants int
	Interval             time.Duration
	stop                 chan struct{}
	done                 chan struct{}
}

// Start begins the poll loop. Call Stop to end it.
func (p *Poller) Start() {
	if p == nil || p.Interface == nil || p.History == nil {
		return
	}
	if p.Interval < time.Second {
		p.Interval = 5 * time.Second
	}
	p.stop = make(chan struct{})
	p.done = make(chan struct{})
	go p.loop()
}

// Stop ends the poll loop and waits for exit.
func (p *Poller) Stop() {
	if p == nil || p.stop == nil {
		return
	}
	select {
	case <-p.stop:
	default:
		close(p.stop)
	}
	<-p.done
}

func (p *Poller) loop() {
	defer close(p.done)
	p.tick()
	t := time.NewTicker(p.Interval)
	defer t.Stop()
	for {
		select {
		case <-p.stop:
			return
		case <-t.C:
			p.tick()
		}
	}
}

func (p *Poller) tick() {
	if err := RefreshAndRecommend(p.Interface, p.History, p.Recommendations, p.ScaleOutParticipants); err != nil {
		log.Printf("fleetpoll: %v", err)
	}
}

// Refresh pulls fleet nodes from the interface and replaces SQLite snapshots.
func Refresh(iface *interfaceclient.Client, hist *historystore.Store) error {
	return RefreshAndRecommend(iface, hist, nil, 0)
}

// RefreshAndRecommend refreshes snapshots then evaluates scale.recommend into ring.
func RefreshAndRecommend(iface *interfaceclient.Client, hist *historystore.Store, rec *webhookstore.Store, threshold int) error {
	if iface == nil || hist == nil {
		return nil
	}
	res, err := iface.FleetNodes()
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("fleet nodes HTTP %d: %s", res.StatusCode, res.Body)
	}
	nodes, err := ParseNodesBody(res.Body)
	if err != nil {
		return err
	}
	if err := hist.ReplaceNodeSnapshots(nodes, time.Now().Unix()); err != nil {
		return err
	}
	MaybeRecommend(hist, rec, threshold)
	return nil
}

// MaybeRecommend lists snapshots and adds a scale.recommend advisory when warranted.
// Skips Add when the newest ring event matches the same advisory fields (avoids
// poll spam while load stays above threshold).
func MaybeRecommend(hist *historystore.Store, rec *webhookstore.Store, threshold int) {
	if hist == nil || rec == nil || threshold < 1 {
		return
	}
	nodes, err := hist.ListNodeSnapshots()
	if err != nil {
		return
	}
	ev, ok := opsrecommend.Evaluate(nodes, threshold)
	if !ok {
		return
	}
	if prev, has := rec.Latest(); has && sameRecommend(prev, ev) {
		return
	}
	rec.Add(ev)
}

func sameRecommend(a, b map[string]any) bool {
	keys := []string{"type", "action", "reason", "threshold", "node", "participant_count"}
	for _, k := range keys {
		if fmt.Sprint(a[k]) != fmt.Sprint(b[k]) {
			return false
		}
	}
	return true
}

// ParseNodesBody decodes `{"nodes":[...]}` into NodeSnapshot rows.
func ParseNodesBody(raw []byte) ([]historystore.NodeSnapshot, error) {
	var wrap struct {
		Nodes []map[string]any `json:"nodes"`
	}
	if err := json.Unmarshal(raw, &wrap); err != nil {
		return nil, err
	}
	out := make([]historystore.NodeSnapshot, 0, len(wrap.Nodes))
	now := time.Now().Unix()
	for _, m := range wrap.Nodes {
		payload, _ := json.Marshal(m)
		n := historystore.NodeSnapshot{
			NodeID:           str(m, "node"),
			Region:           str(m, "region"),
			SessionCount:     int64Val(m, "session_count"),
			ParticipantCount: int64Val(m, "participant_count"),
			Hub:              boolVal(m, "hub"),
			CohostHubCount:   int64Val(m, "cohost_hub_count"),
			Reachable:        true,
			Accepting:        true,
			PolledAt:         now,
			PayloadJSON:      string(payload),
		}
		if _, ok := m["reachable"]; ok {
			n.Reachable = boolVal(m, "reachable")
		}
		if _, ok := m["accepting"]; ok {
			n.Accepting = boolVal(m, "accepting")
		}
		if n.NodeID == "" {
			continue
		}
		out = append(out, n)
	}
	return out, nil
}

func str(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}

func int64Val(m map[string]any, key string) int64 {
	switch v := m[key].(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	case json.Number:
		n, _ := v.Int64()
		return n
	default:
		return 0
	}
}

func boolVal(m map[string]any, key string) bool {
	v, ok := m[key].(bool)
	return ok && v
}
