package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/elfeddi/dxp/internal/gateway/middleware"
	"github.com/elfeddi/dxp/internal/gateway/rbac"
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

	// Étape 1 — Namespace K8s
	if err := kubectlApply(namespaceManifest(req.Namespace)); err != nil {
		result.Steps["namespace"] = fmt.Sprintf("error: %v", err)
		result.Error = err.Error()
		writeJSON(w, http.StatusInternalServerError, result)
		return
	}
	result.Steps["namespace"] = "created"

	// Étape 2 — Application ArgoCD
	if err := kubectlApply(argocdAppManifest(req.Name, req.Repo, req.Namespace)); err != nil {
		result.Steps["argocd_app"] = fmt.Sprintf("error: %v", err)
		result.Error = err.Error()
		writeJSON(w, http.StatusInternalServerError, result)
		return
	}
	result.Steps["argocd_app"] = "created"

	// Étape 3 — Webhook Tekton (best effort)
	if err := createTektonWebhook(req.Name, req.Namespace); err != nil {
		result.Steps["webhook"] = fmt.Sprintf("warning: %v", err)
	} else {
		result.Steps["webhook"] = "created"
	}

	writeJSON(w, http.StatusOK, result)
}

func kubectlApply(manifest string) error {
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(manifest)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s — %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func namespaceManifest(ns string) string {
	return fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: %s
  labels:
    dxp.io/managed: "true"
`, ns)
}

func argocdAppManifest(name, repo, namespace string) string {
	return fmt.Sprintf(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: %s
  namespace: argocd
spec:
  project: default
  source:
    repoURL: %s
    targetRevision: main
    path: k8s
  destination:
    server: https://kubernetes.default.svc
    namespace: %s
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
`, name, repo, namespace)
}

func createTektonWebhook(name, namespace string) error {
	// POC — à implémenter via GitHub API
	return nil
}