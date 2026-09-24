package httpapi

import (
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"strconv"

	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/config"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/fleetpoll"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/historystore"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/interfaceclient"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/mint"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/webhooksign"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/webhookstore"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/web"
)

// Server is the mint + Control-admin-forward HTTP API.
type Server struct {
	Cfg             config.Config
	Interface       *interfaceclient.Client
	Webhooks        *webhookstore.Store
	Recommendations *webhookstore.Store
	History         *historystore.Store
}

// Handler returns the root mux with CORS.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("POST /v1/token", s.handleToken)
	mux.HandleFunc("OPTIONS /v1/token", s.handleOptions)
	mux.HandleFunc("GET /v1/sessions", s.handleSessionsList)
	mux.HandleFunc("GET /v1/sessions/{sid}/presence", s.handlePresence)
	mux.HandleFunc("GET /v1/sessions/{sid}/history", s.handleHistory)
	mux.HandleFunc("POST /v1/sessions/{sid}/data", s.handleData)
	mux.HandleFunc("POST /v1/sessions/{sid}/moderate", s.handleModerate)
	mux.HandleFunc("POST /v1/sessions/{sid}/record/start", s.handleRecordStart)
	mux.HandleFunc("POST /v1/sessions/{sid}/record/stop", s.handleRecordStop)
	mux.HandleFunc("GET /v1/sessions/{sid}/record", s.handleRecordGet)
	mux.HandleFunc("POST /v1/webhooks/xylon", s.handleXylonWebhook)
	mux.HandleFunc("GET /v1/webhooks/xylon/latest", s.handleXylonWebhookLatest)
	mux.HandleFunc("GET /v1/history/sessions", s.handleHistorySessions)
	mux.HandleFunc("GET /v1/history/sessions/{sid}", s.handleHistorySession)
	mux.HandleFunc("GET /v1/fleet/summary", s.handleFleetSummary)
	mux.HandleFunc("GET /v1/fleet/nodes", s.handleFleetNodes)
	mux.HandleFunc("POST /v1/fleet/accepting", s.handleFleetAccepting)
	mux.HandleFunc("GET /v1/ops/summary", s.handleOpsSummary)
	mux.HandleFunc("GET /v1/ops/nodes", s.handleOpsNodes)
	mux.HandleFunc("POST /v1/ops/accepting", s.handleOpsAccepting)
	mux.HandleFunc("GET /v1/ops/recommendations/latest", s.handleOpsRecommendationsLatest)
	if distFS, err := fs.Sub(web.Dist, "dist"); err == nil {
		mux.Handle("GET /assets/", http.FileServer(http.FS(distFS)))
	}
	mux.HandleFunc("GET /{$}", s.handleUIIndex)
	mux.HandleFunc("GET /index.html", s.handleUIIndex)
	return s.withCORS(mux)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleOptions(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleToken(w http.ResponseWriter, r *http.Request) {
	if s.Cfg.HMACSecret == "" {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "no_secret"})
		return
	}
	var body struct {
		SessionID     string `json:"session_id"`
		ParticipantID string `json:"participant_id"`
		Capabilities  any    `json:"capabilities"`
		TTLSec        *int   `json:"ttl_sec"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "bad_json"})
		return
	}
	if body.SessionID == "" || body.ParticipantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "session_id and participant_id required"})
		return
	}
	ttl := s.Cfg.DefaultTTLSec
	if body.TTLSec != nil {
		ttl = *body.TTLSec
	}
	if ttl < 1 {
		ttl = s.Cfg.DefaultTTLSec
	}
	if ttl > s.Cfg.MaxTTLSec {
		ttl = s.Cfg.MaxTTLSec
	}

	tok, err := mint.Token(s.Cfg.HMACSecret, mint.Params{
		TenantID:      s.Cfg.CustomerID,
		SessionID:     body.SessionID,
		ParticipantID: body.ParticipantID,
		Capabilities:  body.Capabilities,
		TTLSec:        ttl,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"token": tok})
}

func (s *Server) handlePresence(w http.ResponseWriter, r *http.Request) {
	s.forward(w, func() (interfaceclient.Result, error) {
		return s.Interface.Presence(r.PathValue("sid"))
	})
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	s.forward(w, func() (interfaceclient.Result, error) {
		return s.Interface.History(r.PathValue("sid"), r.URL.Query().Get("label"))
	})
}

func (s *Server) handleData(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "read_body"})
		return
	}
	s.forward(w, func() (interfaceclient.Result, error) {
		return s.Interface.Inject(r.PathValue("sid"), body)
	})
}

func (s *Server) handleModerate(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "read_body"})
		return
	}
	s.forward(w, func() (interfaceclient.Result, error) {
		return s.Interface.Moderate(r.PathValue("sid"), body)
	})
}

func (s *Server) handleRecordStart(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "read_body"})
		return
	}
	s.forward(w, func() (interfaceclient.Result, error) {
		return s.Interface.RecordStart(r.PathValue("sid"), body)
	})
}

func (s *Server) handleRecordStop(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "read_body"})
		return
	}
	s.forward(w, func() (interfaceclient.Result, error) {
		return s.Interface.RecordStop(r.PathValue("sid"), body)
	})
}

func (s *Server) handleRecordGet(w http.ResponseWriter, r *http.Request) {
	s.forward(w, func() (interfaceclient.Result, error) {
		return s.Interface.RecordGet(r.PathValue("sid"))
	})
}

func (s *Server) handleSessionsList(w http.ResponseWriter, r *http.Request) {
	s.forward(w, func() (interfaceclient.Result, error) {
		return s.Interface.SessionsList()
	})
}

func (s *Server) handleXylonWebhook(w http.ResponseWriter, r *http.Request) {
	if s.Cfg.WebhookSecret == "" {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "webhook_unconfigured"})
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "read_body"})
		return
	}
	sig := r.Header.Get(webhooksign.Header)
	if !webhooksign.Valid(body, s.Cfg.WebhookSecret, sig) {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "bad_signature"})
		return
	}
	var ev map[string]any
	if err := json.Unmarshal(body, &ev); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_json"})
		return
	}
	if s.History != nil {
		if err := s.History.ApplyWebhook(body, ev); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "history_ingest", "detail": err.Error()})
			return
		}
	}
	if s.Webhooks == nil {
		s.Webhooks = webhookstore.New(32)
	}
	s.Webhooks.Add(ev)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleHistorySessions(w http.ResponseWriter, r *http.Request) {
	if s.History == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "history_unconfigured"})
		return
	}
	limit := 50
	if q := r.URL.Query().Get("limit"); q != "" {
		if n, err := strconv.Atoi(q); err == nil {
			limit = n
		}
	}
	list, err := s.History.ListSessions(s.Cfg.CustomerID, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "history_list", "detail": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"sessions": list})
}

func (s *Server) handleHistorySession(w http.ResponseWriter, r *http.Request) {
	if s.History == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "history_unconfigured"})
		return
	}
	sid := r.PathValue("sid")
	detail, err := s.History.GetSession(s.Cfg.CustomerID, sid)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "history_get", "detail": err.Error()})
		return
	}
	if detail == nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "not_found"})
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (s *Server) handleFleetNodes(w http.ResponseWriter, r *http.Request) {
	s.writeFleetNodes(w, r)
}

func (s *Server) handleFleetSummary(w http.ResponseWriter, r *http.Request) {
	s.writeFleetSummary(w, r)
}

func (s *Server) handleOpsNodes(w http.ResponseWriter, r *http.Request) {
	s.writeFleetNodes(w, r)
}

func (s *Server) handleOpsSummary(w http.ResponseWriter, r *http.Request) {
	s.writeFleetSummary(w, r)
}

func (s *Server) handleFleetAccepting(w http.ResponseWriter, r *http.Request) {
	s.forwardAccepting(w, r)
}

func (s *Server) handleOpsAccepting(w http.ResponseWriter, r *http.Request) {
	s.forwardAccepting(w, r)
}

func (s *Server) forwardAccepting(w http.ResponseWriter, r *http.Request) {
	if s.Interface == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "interface_unconfigured"})
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "read_body"})
		return
	}
	res, err := s.Interface.FleetAccepting(body)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": "interface_unreachable", "detail": err.Error()})
		return
	}
	if res.StatusCode >= 200 && res.StatusCode < 300 && s.History != nil {
		_ = fleetpoll.RefreshAndRecommend(s.Interface, s.History, s.Recommendations, s.Cfg.ScaleOutParticipants)
	}
	if res.ContentType != "" {
		w.Header().Set("Content-Type", res.ContentType)
	} else {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(res.StatusCode)
	_, _ = w.Write(res.Body)
}

func (s *Server) handleOpsRecommendationsLatest(w http.ResponseWriter, r *http.Request) {
	if s.Recommendations == nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "empty"})
		return
	}
	ev, ok := s.Recommendations.Latest()
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "empty"})
		return
	}
	writeJSON(w, http.StatusOK, ev)
}

func (s *Server) writeFleetNodes(w http.ResponseWriter, r *http.Request) {
	if s.History == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "history_unconfigured"})
		return
	}
	if err := s.maybeFleetRefresh(r); err != nil {
		writeJSON(w, err.status, err.body)
		return
	}
	nodes, err := s.History.ListNodeSnapshots()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "fleet_list", "detail": err.Error()})
		return
	}
	// Counts/ids only for fleet + ops machine API (omit internal payload_json).
	for i := range nodes {
		nodes[i].PayloadJSON = ""
	}
	writeJSON(w, http.StatusOK, map[string]any{"nodes": nodes})
}

func (s *Server) writeFleetSummary(w http.ResponseWriter, r *http.Request) {
	if s.History == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "history_unconfigured"})
		return
	}
	if err := s.maybeFleetRefresh(r); err != nil {
		writeJSON(w, err.status, err.body)
		return
	}
	nodes, err := s.History.ListNodeSnapshots()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "fleet_list", "detail": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"summary": historystore.FleetSummaryFromSnapshots(nodes),
	})
}

type httpErr struct {
	status int
	body   map[string]any
}

func (s *Server) maybeFleetRefresh(r *http.Request) *httpErr {
	if r.URL.Query().Get("refresh") != "1" {
		return nil
	}
	if s.Interface == nil {
		return &httpErr{http.StatusServiceUnavailable, map[string]any{"error": "interface_unconfigured"}}
	}
	if err := fleetpoll.RefreshAndRecommend(s.Interface, s.History, s.Recommendations, s.Cfg.ScaleOutParticipants); err != nil {
		return &httpErr{http.StatusBadGateway, map[string]any{"error": "fleet_refresh", "detail": err.Error()}}
	}
	return nil
}

func (s *Server) handleUIIndex(w http.ResponseWriter, r *http.Request) {
	data, err := web.Dist.ReadFile("dist/index.html")
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "ui_missing"})
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (s *Server) handleXylonWebhookLatest(w http.ResponseWriter, r *http.Request) {
	if s.Webhooks == nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "empty"})
		return
	}
	ev, ok := s.Webhooks.Latest()
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "empty"})
		return
	}
	writeJSON(w, http.StatusOK, ev)
}

func (s *Server) forward(w http.ResponseWriter, call func() (interfaceclient.Result, error)) {
	if s.Interface == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "interface_unconfigured"})
		return
	}
	res, err := call()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": "interface_unreachable", "detail": err.Error()})
		return
	}
	if res.ContentType != "" {
		w.Header().Set("Content-Type", res.ContentType)
	} else {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(res.StatusCode)
	_, _ = w.Write(res.Body)
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	allowed := map[string]struct{}{}
	for _, o := range s.Cfg.CORSOrigins {
		allowed[o] = struct{}{}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if _, ok := allowed[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
