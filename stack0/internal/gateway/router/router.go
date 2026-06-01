package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/elfeddi/dxp/internal/gateway"
	"github.com/elfeddi/dxp/internal/gateway/middleware"
	"github.com/elfeddi/dxp/internal/gateway/rbac"
	"github.com/elfeddi/dxp/internal/aigateway"
	"github.com/elfeddi/dxp/internal/observer"
)

// Server est le serveur HTTP C4 Gateway.
type Server struct {
	registry *gateway.Registry
	rbac     *rbac.Checker
	obs      *observer.Observer
	ai       *aigateway.Gateway
	mux      *http.ServeMux
}

// New crée un Server et enregistre toutes les routes.
func New(reg *gateway.Registry, obs *observer.Observer, ai *aigateway.Gateway) *Server {
	s := &Server{
		registry: reg,
		rbac:     rbac.NewChecker(),
		obs:      obs,
		ai:       ai,
		mux:      http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

// Handler retourne le http.Handler avec les middlewares appliqués.
func (s *Server) Handler() http.Handler {
	chain := middleware.Chain(
		middleware.Recovery,
		middleware.Logger,
		middleware.CORS,
		middleware.Auth,
	)
	return chain(s.mux)
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /healthz", s.handleHealthz)
	s.mux.HandleFunc("GET /readyz", s.handleReadyz)
	s.mux.HandleFunc("GET /api/dxp/status", s.handlePlatformStatus)
	s.mux.HandleFunc("POST /api/dxp/generate", s.handleGenerate)
	s.mux.HandleFunc("POST /api/dxp/diagnose", s.handleDiagnose)
	s.mux.HandleFunc("POST /api/dxp/provision", s.handleProvision)
	s.mux.HandleFunc("GET /api/dxp/metrics", s.handlePlatformMetrics)
	s.mux.HandleFunc("GET /api/dxp/backends", s.handleListBackends)
	s.mux.HandleFunc("GET /api/dxp/backends/{name}/status", s.handleGetStatus)
	s.mux.HandleFunc("GET /api/dxp/backends/{name}/metrics", s.handleGetMetrics)
	s.mux.HandleFunc("GET /api/dxp/backends/{name}/logs", s.handleGetLogs)
	s.mux.HandleFunc("GET /api/dxp/backends/{name}/actions", s.handleListActions)
	s.mux.HandleFunc("POST /api/dxp/backends/{name}/actions", s.handleExecuteAction)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if s.obs == nil {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	}
	health := s.obs.Health.Get()
	// Toujours retourner 200 — K8s gère le restart via failureThreshold
	writeJSON(w, http.StatusOK, map[string]any{
		"status":     "ok",
		"healthy":    health.Healthy,
		"ready":      health.Ready,
		"checked_at": health.CheckedAt,
	})
}

func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	if s.obs == nil {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
		return
	}
	health := s.obs.Health.Get()
	if !health.Ready {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not ready"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (s *Server) handlePlatformStatus(w http.ResponseWriter, r *http.Request) {
	if s.obs == nil {
		middleware.WriteError(w, http.StatusServiceUnavailable, "C5 observer not initialized")
		return
	}
	writeJSON(w, http.StatusOK, s.obs.Status.Get())
}

func (s *Server) handlePlatformMetrics(w http.ResponseWriter, r *http.Request) {
	if s.obs == nil {
		middleware.WriteError(w, http.StatusServiceUnavailable, "C5 observer not initialized")
		return
	}
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(s.obs.Metrics.Collect()))
}

func (s *Server) handleListBackends(w http.ResponseWriter, r *http.Request) {
	role, ok := middleware.RoleFromContext(r.Context())
	if !ok {
		middleware.WriteError(w, http.StatusForbidden, "no role in context")
		return
	}
	if err := s.rbac.Check(role, rbac.OpGetStatus, false); err != nil {
		middleware.WriteError(w, http.StatusForbidden, err.Error())
		return
	}
	backends := s.registry.List()
	statuses := make([]*gateway.ComponentStatus, 0, len(backends))
	for _, name := range backends {
		b, _ := s.registry.Get(name)
		status, err := b.GetStatus(r.Context())
		if err != nil {
			status = &gateway.ComponentStatus{Name: name, Healthy: false, Message: err.Error()}
		}
		statuses = append(statuses, status)
	}
	writeJSON(w, http.StatusOK, map[string]any{"backends": statuses, "total": len(statuses)})
}

func (s *Server) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	_, b, ok := s.resolveBackend(w, r, rbac.OpGetStatus, false)
	if !ok {
		return
	}
	status, err := b.GetStatus(r.Context())
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	_, b, ok := s.resolveBackend(w, r, rbac.OpGetMetrics, false)
	if !ok {
		return
	}
	metrics, err := b.GetMetrics(r.Context())
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(metrics))
}

func (s *Server) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	_, b, ok := s.resolveBackend(w, r, rbac.OpGetLogs, false)
	if !ok {
		return
	}
	req := &gateway.LogRequest{
		Lines:     parseIntQuery(r, "lines", 100),
		Since:     r.URL.Query().Get("since"),
		Filter:    r.URL.Query().Get("filter"),
		Namespace: r.URL.Query().Get("namespace"),
	}
	logs, err := b.GetLogs(r.Context(), req)
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"logs": logs, "count": len(logs)})
}

func (s *Server) handleListActions(w http.ResponseWriter, r *http.Request) {
	_, b, ok := s.resolveBackend(w, r, rbac.OpListActions, false)
	if !ok {
		return
	}
	actions, err := b.ListActions(r.Context())
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"actions": actions})
}

func (s *Server) handleExecuteAction(w http.ResponseWriter, r *http.Request) {
	var req gateway.ActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %s", err))
		return
	}
	name := r.PathValue("name")
	b, err := s.registry.Get(name)
	if err != nil {
		middleware.WriteError(w, http.StatusNotFound, fmt.Sprintf("backend %q not found", name))
		return
	}
	dangerous := false
	actions, err := b.ListActions(r.Context())
	if err == nil {
		for _, a := range actions {
			if a.Name == req.Action {
				dangerous = a.Dangerous
				break
			}
		}
	}
	_, _, ok := s.resolveBackend(w, r, rbac.OpExecuteAction, dangerous)
	if !ok {
		return
	}
	result, err := b.ExecuteAction(r.Context(), &req)
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	code := http.StatusOK
	if !result.Success {
		code = http.StatusUnprocessableEntity
	}
	writeJSON(w, code, result)
}

func (s *Server) resolveBackend(
	w http.ResponseWriter,
	r *http.Request,
	op rbac.Op,
	dangerous bool,
) (rbac.Role, gateway.BackendConnector, bool) {
	role, ok := middleware.RoleFromContext(r.Context())
	if !ok {
		middleware.WriteError(w, http.StatusForbidden, "no role in context")
		return "", nil, false
	}
	if err := s.rbac.Check(role, op, dangerous); err != nil {
		middleware.WriteError(w, http.StatusForbidden, err.Error())
		return "", nil, false
	}
	name := r.PathValue("name")
	b, err := s.registry.Get(name)
	if err != nil {
		middleware.WriteError(w, http.StatusNotFound, fmt.Sprintf("backend %q not found", name))
		return "", nil, false
	}
	return role, b, true
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	if s.ai == nil {
		middleware.WriteError(w, http.StatusServiceUnavailable, "C6 AIGateway not initialized")
		return
	}

	var req struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %s", err))
		return
	}
	if strings.TrimSpace(req.Prompt) == "" {
		middleware.WriteError(w, http.StatusBadRequest, "prompt is required")
		return
	}

	result, err := s.ai.Generate(r.Context(), req.Prompt)
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	code := http.StatusOK
	if !result.Valid {
		code = http.StatusUnprocessableEntity
	}
	writeJSON(w, code, result)
}

func (s *Server) handleDiagnose(w http.ResponseWriter, r *http.Request) {
	if s.ai == nil {
		middleware.WriteError(w, http.StatusServiceUnavailable, "C6 AIGateway not initialized")
		return
	}

	var req struct {
		Incident string `json:"incident"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %s", err))
		return
	}
	if strings.TrimSpace(req.Incident) == "" {
		middleware.WriteError(w, http.StatusBadRequest, "incident is required")
		return
	}

	result, err := s.ai.Diagnose(r.Context(), req.Incident)
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func parseIntQuery(r *http.Request, key string, defaultVal int) int {
	val := strings.TrimSpace(r.URL.Query().Get(key))
	if val == "" {
		return defaultVal
	}
	var n int
	if _, err := fmt.Sscanf(val, "%d", &n); err != nil || n <= 0 {
		return defaultVal
	}
	return n
}
