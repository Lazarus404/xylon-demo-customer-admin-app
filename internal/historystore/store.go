// Package historystore persists Control webhook events in SQLite for the demo admin console.
package historystore

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Store is a SQLite-backed history ingest for lifecycle and recording webhooks.
type Store struct {
	db *sql.DB
}

// SessionRow is one row from the sessions table.
type SessionRow struct {
	TenantID  string  `json:"tenant_id"`
	SessionID string  `json:"session_id"`
	CreatedAt *int64  `json:"created_at"`
	ClosedAt  *int64  `json:"closed_at"`
	LastNode  string  `json:"last_node"`
	LastRegion string `json:"last_region"`
}

// ParticipantEvent is a join/leave row.
type ParticipantEvent struct {
	ID            int64  `json:"id"`
	TenantID      string `json:"tenant_id"`
	SessionID     string `json:"session_id"`
	ParticipantID string `json:"participant_id"`
	EventType     string `json:"event_type"`
	Node          string `json:"node"`
	Region        string `json:"region"`
	OccurredAt    int64  `json:"occurred_at"`
}

// RecordingSegment is one recording.completed / recording.failed upsert.
type RecordingSegment struct {
	TenantID    string `json:"tenant_id"`
	RecordingID string `json:"recording_id"`
	SessionID   string `json:"session_id"`
	Sequence    int64  `json:"sequence"`
	Mode        string `json:"mode"`
	Status      string `json:"status"`
	StartedAt   *int64 `json:"started_at"`
	EndedAt     *int64 `json:"ended_at"`
	PayloadJSON string `json:"payload_json"`
}

// SessionDetail is a session plus related participant and recording rows.
type SessionDetail struct {
	Session      SessionRow          `json:"session"`
	Participants []ParticipantEvent  `json:"participants"`
	Recordings   []RecordingSegment  `json:"recordings"`
}

// Open opens or creates the SQLite database at path (creates parent dirs).
func Open(path string) (*Store, error) {
	if path == "" {
		return nil, fmt.Errorf("sqlite path required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("mkdir: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		_ = db.Close()
		return nil, err
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// Close closes the underlying database.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS events_raw (
  payload_hash TEXT PRIMARY KEY,
  event_type TEXT NOT NULL,
  tenant_id TEXT NOT NULL,
  session_id TEXT NOT NULL,
  payload_json TEXT NOT NULL,
  received_at INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS sessions (
  tenant_id TEXT NOT NULL,
  session_id TEXT NOT NULL,
  created_at INTEGER,
  closed_at INTEGER,
  last_node TEXT NOT NULL DEFAULT '',
  last_region TEXT NOT NULL DEFAULT '',
  PRIMARY KEY (tenant_id, session_id)
);
CREATE TABLE IF NOT EXISTS participant_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  tenant_id TEXT NOT NULL,
  session_id TEXT NOT NULL,
  participant_id TEXT NOT NULL,
  event_type TEXT NOT NULL,
  node TEXT NOT NULL DEFAULT '',
  region TEXT NOT NULL DEFAULT '',
  occurred_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_participant_events_session
  ON participant_events (tenant_id, session_id, occurred_at);
CREATE TABLE IF NOT EXISTS recording_segments (
  tenant_id TEXT NOT NULL,
  recording_id TEXT NOT NULL,
  session_id TEXT NOT NULL,
  sequence INTEGER NOT NULL DEFAULT 0,
  mode TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL,
  started_at INTEGER,
  ended_at INTEGER,
  payload_json TEXT NOT NULL,
  PRIMARY KEY (tenant_id, recording_id)
);
CREATE INDEX IF NOT EXISTS idx_sessions_list
  ON sessions (tenant_id, created_at DESC);
CREATE TABLE IF NOT EXISTS node_snapshots (
  node_id TEXT PRIMARY KEY,
  region TEXT NOT NULL DEFAULT '',
  session_count INTEGER NOT NULL DEFAULT 0,
  participant_count INTEGER NOT NULL DEFAULT 0,
  hub INTEGER NOT NULL DEFAULT 0,
  cohost_hub_count INTEGER NOT NULL DEFAULT 0,
  reachable INTEGER NOT NULL DEFAULT 1,
  accepting INTEGER NOT NULL DEFAULT 1,
  polled_at INTEGER NOT NULL,
  payload_json TEXT NOT NULL
);
`)
	if err != nil {
		return err
	}
	// Existing DBs created before accepting column.
	_, _ = s.db.Exec(`ALTER TABLE node_snapshots ADD COLUMN accepting INTEGER NOT NULL DEFAULT 1`)
	return nil
}

// ApplyWebhook records a verified webhook body. Duplicate payload_hash is a no-op.
func (s *Store) ApplyWebhook(rawBody []byte, ev map[string]any) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("store closed")
	}
	sum := sha256.Sum256(rawBody)
	hash := hex.EncodeToString(sum[:])
	typ := strField(ev, "type")
	tenant := strField(ev, "tenant_id")
	sid := strField(ev, "session_id")
	now := time.Now().Unix()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.Exec(
		`INSERT OR IGNORE INTO events_raw (payload_hash, event_type, tenant_id, session_id, payload_json, received_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		hash, typ, tenant, sid, string(rawBody), now,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return tx.Commit()
	}

	switch typ {
	case "session.created":
		if err := ensureSession(tx, tenant, sid, strField(ev, "node"), strField(ev, "region"), &now, nil); err != nil {
			return err
		}
	case "session.closed":
		if err := ensureSession(tx, tenant, sid, strField(ev, "node"), strField(ev, "region"), nil, &now); err != nil {
			return err
		}
		if _, err := tx.Exec(
			`UPDATE sessions SET closed_at = COALESCE(closed_at, ?), last_node = CASE WHEN ? != '' THEN ? ELSE last_node END,
			 last_region = CASE WHEN ? != '' THEN ? ELSE last_region END
			 WHERE tenant_id = ? AND session_id = ?`,
			now, strField(ev, "node"), strField(ev, "node"), strField(ev, "region"), strField(ev, "region"), tenant, sid,
		); err != nil {
			return err
		}
	case "participant.joined", "participant.left":
		node, region := strField(ev, "node"), strField(ev, "region")
		if err := ensureSession(tx, tenant, sid, node, region, nil, nil); err != nil {
			return err
		}
		if _, err := tx.Exec(
			`INSERT INTO participant_events (tenant_id, session_id, participant_id, event_type, node, region, occurred_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			tenant, sid, strField(ev, "participant_id"), typ, node, region, now,
		); err != nil {
			return err
		}
	case "recording.completed", "recording.failed":
		node, region := strField(ev, "node"), strField(ev, "region")
		if err := ensureSession(tx, tenant, sid, node, region, nil, nil); err != nil {
			return err
		}
		status := "completed"
		if typ == "recording.failed" {
			status = "failed"
		}
		seq := int64Field(ev, "sequence")
		started := int64PtrField(ev, "started_at")
		ended := int64PtrField(ev, "ended_at")
		payload, _ := json.Marshal(ev)
		if _, err := tx.Exec(
			`INSERT INTO recording_segments (tenant_id, recording_id, session_id, sequence, mode, status, started_at, ended_at, payload_json)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			 ON CONFLICT(tenant_id, recording_id) DO UPDATE SET
			   session_id = excluded.session_id,
			   sequence = excluded.sequence,
			   mode = excluded.mode,
			   status = excluded.status,
			   started_at = excluded.started_at,
			   ended_at = excluded.ended_at,
			   payload_json = excluded.payload_json`,
			tenant, strField(ev, "recording_id"), sid, seq, strField(ev, "mode"), status, started, ended, string(payload),
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func ensureSession(tx *sql.Tx, tenant, sid, node, region string, created, closed *int64) error {
	if tenant == "" || sid == "" {
		return nil
	}
	_, err := tx.Exec(
		`INSERT INTO sessions (tenant_id, session_id, created_at, closed_at, last_node, last_region)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(tenant_id, session_id) DO UPDATE SET
		   created_at = COALESCE(sessions.created_at, excluded.created_at),
		   closed_at = COALESCE(sessions.closed_at, excluded.closed_at),
		   last_node = CASE WHEN excluded.last_node != '' THEN excluded.last_node ELSE sessions.last_node END,
		   last_region = CASE WHEN excluded.last_region != '' THEN excluded.last_region ELSE sessions.last_region END`,
		tenant, sid, created, closed, node, region,
	)
	return err
}

// ListSessions returns sessions for tenant, newest created_at first.
func (s *Store) ListSessions(tenantID string, limit int) ([]SessionRow, error) {
	if limit < 1 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	rows, err := s.db.Query(
		`SELECT tenant_id, session_id, created_at, closed_at, last_node, last_region
		 FROM sessions WHERE tenant_id = ?
		 ORDER BY COALESCE(created_at, 0) DESC, rowid DESC
		 LIMIT ?`,
		tenantID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SessionRow
	for rows.Next() {
		var r SessionRow
		var created, closed sql.NullInt64
		if err := rows.Scan(&r.TenantID, &r.SessionID, &created, &closed, &r.LastNode, &r.LastRegion); err != nil {
			return nil, err
		}
		if created.Valid {
			v := created.Int64
			r.CreatedAt = &v
		}
		if closed.Valid {
			v := closed.Int64
			r.ClosedAt = &v
		}
		out = append(out, r)
	}
	if out == nil {
		out = []SessionRow{}
	}
	return out, rows.Err()
}

// GetSession returns session detail or nil if missing.
func (s *Store) GetSession(tenantID, sessionID string) (*SessionDetail, error) {
	var r SessionRow
	var created, closed sql.NullInt64
	err := s.db.QueryRow(
		`SELECT tenant_id, session_id, created_at, closed_at, last_node, last_region
		 FROM sessions WHERE tenant_id = ? AND session_id = ?`,
		tenantID, sessionID,
	).Scan(&r.TenantID, &r.SessionID, &created, &closed, &r.LastNode, &r.LastRegion)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if created.Valid {
		v := created.Int64
		r.CreatedAt = &v
	}
	if closed.Valid {
		v := closed.Int64
		r.ClosedAt = &v
	}

	parts, err := s.listParticipants(tenantID, sessionID)
	if err != nil {
		return nil, err
	}
	recs, err := s.listRecordings(tenantID, sessionID)
	if err != nil {
		return nil, err
	}
	return &SessionDetail{Session: r, Participants: parts, Recordings: recs}, nil
}

func (s *Store) listParticipants(tenantID, sessionID string) ([]ParticipantEvent, error) {
	rows, err := s.db.Query(
		`SELECT id, tenant_id, session_id, participant_id, event_type, node, region, occurred_at
		 FROM participant_events WHERE tenant_id = ? AND session_id = ?
		 ORDER BY occurred_at ASC, id ASC LIMIT 500`,
		tenantID, sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ParticipantEvent
	for rows.Next() {
		var p ParticipantEvent
		if err := rows.Scan(&p.ID, &p.TenantID, &p.SessionID, &p.ParticipantID, &p.EventType, &p.Node, &p.Region, &p.OccurredAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if out == nil {
		out = []ParticipantEvent{}
	}
	return out, rows.Err()
}

func (s *Store) listRecordings(tenantID, sessionID string) ([]RecordingSegment, error) {
	rows, err := s.db.Query(
		`SELECT tenant_id, recording_id, session_id, sequence, mode, status, started_at, ended_at, payload_json
		 FROM recording_segments WHERE tenant_id = ? AND session_id = ?
		 ORDER BY sequence ASC, recording_id ASC LIMIT 200`,
		tenantID, sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RecordingSegment
	for rows.Next() {
		var r RecordingSegment
		var started, ended sql.NullInt64
		if err := rows.Scan(&r.TenantID, &r.RecordingID, &r.SessionID, &r.Sequence, &r.Mode, &r.Status, &started, &ended, &r.PayloadJSON); err != nil {
			return nil, err
		}
		if started.Valid {
			v := started.Int64
			r.StartedAt = &v
		}
		if ended.Valid {
			v := ended.Int64
			r.EndedAt = &v
		}
		out = append(out, r)
	}
	if out == nil {
		out = []RecordingSegment{}
	}
	return out, rows.Err()
}

// NodeSnapshot is one cached fleet node row.
type NodeSnapshot struct {
	NodeID           string `json:"node"`
	Region           string `json:"region"`
	SessionCount     int64  `json:"session_count"`
	ParticipantCount int64  `json:"participant_count"`
	Hub              bool   `json:"hub"`
	CohostHubCount   int64  `json:"cohost_hub_count"`
	Reachable        bool   `json:"reachable"`
	Accepting        bool   `json:"accepting"`
	PolledAt         int64  `json:"polled_at"`
	PayloadJSON      string `json:"payload_json,omitempty"`
}

// ReplaceNodeSnapshots replaces all node_snapshots with the given rows (same polled_at).
func (s *Store) ReplaceNodeSnapshots(nodes []NodeSnapshot, polledAt int64) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("store closed")
	}
	if polledAt < 1 {
		polledAt = time.Now().Unix()
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Exec(`DELETE FROM node_snapshots`); err != nil {
		return err
	}
	for _, n := range nodes {
		hub, reachable, accepting := 0, 0, 0
		if n.Hub {
			hub = 1
		}
		if n.Reachable {
			reachable = 1
		}
		if n.Accepting {
			accepting = 1
		}
		payload := n.PayloadJSON
		if payload == "" {
			b, _ := json.Marshal(n)
			payload = string(b)
		}
		if _, err := tx.Exec(
			`INSERT INTO node_snapshots (node_id, region, session_count, participant_count, hub, cohost_hub_count, reachable, accepting, polled_at, payload_json)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			n.NodeID, n.Region, n.SessionCount, n.ParticipantCount, hub, n.CohostHubCount, reachable, accepting, polledAt, payload,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ListNodeSnapshots returns cached fleet nodes newest polled first.
func (s *Store) ListNodeSnapshots() ([]NodeSnapshot, error) {
	rows, err := s.db.Query(
		`SELECT node_id, region, session_count, participant_count, hub, cohost_hub_count, reachable, accepting, polled_at, payload_json
		 FROM node_snapshots ORDER BY node_id ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []NodeSnapshot
	for rows.Next() {
		var n NodeSnapshot
		var hub, reachable, accepting int
		if err := rows.Scan(&n.NodeID, &n.Region, &n.SessionCount, &n.ParticipantCount, &hub, &n.CohostHubCount, &reachable, &accepting, &n.PolledAt, &n.PayloadJSON); err != nil {
			return nil, err
		}
		n.Hub = hub != 0
		n.Reachable = reachable != 0
		n.Accepting = accepting != 0
		out = append(out, n)
	}
	if out == nil {
		out = []NodeSnapshot{}
	}
	return out, rows.Err()
}

// FleetSummaryFromSnapshots derives a summary map from cached node rows.
func FleetSummaryFromSnapshots(nodes []NodeSnapshot) map[string]any {
	regions := map[string]struct{}{}
	var sessions, participants int64
	healthy, unhealthy := 0, 0
	hubNode := ""
	for _, n := range nodes {
		if n.Region != "" {
			regions[n.Region] = struct{}{}
		}
		sessions += n.SessionCount
		participants += n.ParticipantCount
		if n.Reachable {
			healthy++
		} else {
			unhealthy++
		}
		if n.Hub && hubNode == "" {
			hubNode = n.NodeID
		}
	}
	regionList := make([]string, 0, len(regions))
	for r := range regions {
		regionList = append(regionList, r)
	}
	sort.Strings(regionList)
	return map[string]any{
		"regions":           regionList,
		"node_count":        len(nodes),
		"healthy_count":     healthy,
		"unhealthy_count":   unhealthy,
		"session_count":     sessions,
		"participant_count": participants,
		"hub_node":          hubNode,
		"connected":         healthy > 0,
	}
}

func strField(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatInt(int64(t), 10)
	case json.Number:
		return t.String()
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

func int64Field(m map[string]any, key string) int64 {
	p := int64PtrField(m, key)
	if p == nil {
		return 0
	}
	return *p
}

func int64PtrField(m map[string]any, key string) *int64 {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	switch t := v.(type) {
	case float64:
		n := int64(t)
		return &n
	case int64:
		return &t
	case int:
		n := int64(t)
		return &n
	case json.Number:
		n, err := t.Int64()
		if err != nil {
			return nil
		}
		return &n
	case string:
		n, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			return nil
		}
		return &n
	default:
		return nil
	}
}
