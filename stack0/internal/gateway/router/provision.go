package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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

	// Client K8s via ServiceAccount du pod
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
			Name: req.Namespace,
			Labels: map[string]string{
				"dxp.io/managed": "true",
			},
		},
	}
	_, err = clientset.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
	if err != nil {
		// Ignorer si déjà existant
		if !isAlreadyExists(err) {
			result.Steps["namespace"] = fmt.Sprintf("error: %v", err)
			result.Error = err.Error()
			writeJSON(w, http.StatusInternalServerError, result)
			return
		}
	}
	result.Steps["namespace"] = "created"

	// Étape 2 — Application ArgoCD (backlog — dynamic client)
	result.Steps["argocd_app"] = "pending"

	// Étape 3 — Webhook (backlog)
	result.Steps["webhook"] = "pending"

	writeJSON(w, http.StatusOK, result)
}

func isAlreadyExists(err error) bool {
	return err != nil && len(err.Error()) > 0 &&
		(contains(err.Error(), "already exists"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > 0 && len(substr) > 0 &&
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
