package httpapi

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/config"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/interfaceclient"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/mint"
)

// Server is the mint + Control-admin-forward HTTP API.
type Server struct {
	Cfg       config.Config
	Interface *interfaceclient.Client
}

// Handler returns the root mux with CORS.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("POST /v1/token", s.handleToken)
	mux.HandleFunc("OPTIONS /v1/token", s.handleOptions)
	mux.HandleFunc("GET /v1/sessions/{sid}/presence", s.handlePresence)
	mux.HandleFunc("GET /v1/sessions/{sid}/history", s.handleHistory)
	mux.HandleFunc("POST /v1/sessions/{sid}/data", s.handleData)
	mux.HandleFunc("POST /v1/sessions/{sid}/moderate", s.handleModerate)
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
