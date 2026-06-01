package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elfeddi/dxp/internal/gateway"
	"github.com/elfeddi/dxp/internal/gateway/middleware"
	"github.com/elfeddi/dxp/internal/gateway/rbac"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ProvisionRequest struct {
	Name      string `json:"name"`
	Repo      string `json:"repo"`
	Namespace string `json:"namespace"`
	Language  string `json:"language"`
}

type ProvisionResult struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Steps     map[string]string `json:"steps"`
	Error     string            `json:"error,omitempty"`
}

func (s *Server) handleProvision(w http.ResponseWriter, r *http.Request) {
	role, ok := middleware.RoleFromContext(r.Context())
	if !ok {
		middleware.WriteError(w, http.StatusForbidden, "no role in context")
		return
	}
	if err := s.rbac.Check(role, rbac.OpExecuteAction, false); err != nil {
		middleware.WriteError(w, http.StatusForbidden, err.Error())
		return
	}

	var req ProvisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %s", err))
		return
	}
	if req.Name == "" || req.Repo == "" || req.Namespace == "" {
		middleware.WriteError(w, http.StatusBadRequest, "name, repo and namespace are required")
		return
	}

	result := ProvisionResult{
		Name:      req.Name,
		Namespace: req.Namespace,
		Steps:     make(map[string]string),
	}

	// Client K8s via ServiceAccount
	config, err := rest.InClusterConfig()
	if err != nil {
		result.Error = fmt.Sprintf("k8s config error: %v", err)
		writeJSON(w, http.StatusInternalServerError, result)
		return
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		result.Error = fmt.Sprintf("k8s client error: %v", err)
		writeJSON(w, http.StatusInternalServerError, result)
		return
	}

	// Étape 1 — Namespace K8s
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   req.Namespace,
			Labels: map[string]string{"dxp.io/managed": "true"},
		},
	}
	_, err = clientset.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
	if err != nil && !isAlreadyExists(err) {
		result.Steps["namespace"] = fmt.Sprintf("error: %v", err)
		result.Error = err.Error()
		writeJSON(w, http.StatusInternalServerError, result)
		return
	}
	result.Steps["namespace"] = "created"

	// Étape 2 — Application ArgoCD via backend C4
	argoBackend, err := s.registry.Get("argocd-main")
	if err != nil {
		argoBackend, err = s.registry.Get("argocd")
	}
	if err != nil {
		result.Steps["argocd_app"] = "error: argocd backend not found"
		result.Error = "argocd backend not found"
		writeJSON(w, http.StatusInternalServerError, result)
		return
	}

	actionReq := &gateway.ActionRequest{
		Action: "create-app",
		Target: req.Name,
		Params: map[string]string{
			"repo":      req.Repo,
			"namespace": req.Namespace,
			"path":      "k8s",
		},
	}

	actionResult, err := argoBackend.ExecuteAction(r.Context(), actionReq)
	if err != nil {
		result.Steps["argocd_app"] = fmt.Sprintf("error: %v", err)
	} else if !actionResult.Success {
		result.Steps["argocd_app"] = fmt.Sprintf("error: %s", actionResult.Message)
	} else {
		result.Steps["argocd_app"] = "created"
	}

	// Étape 3 — Webhook (backlog)
	result.Steps["webhook"] = "pending"

	writeJSON(w, http.StatusOK, result)
}

func isAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	return contains(err.Error(), "already exists")
}

func contains(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
