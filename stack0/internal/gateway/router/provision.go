package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

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

	// Étape 3 — Webhook + désactivation protection branche
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		result.Steps["webhook"] = "skipped: GITHUB_TOKEN not set"
	} else {
		owner, repoName, whErr := parseGitHubRepo(req.Repo)
		if whErr != nil {
			result.Steps["webhook"] = fmt.Sprintf("skipped: %v", whErr)
		} else {
			webhookURL := os.Getenv("DXP_WEBHOOK_URL")
			if webhookURL == "" {
				webhookURL = "http://158.158.8.131:30092/"
			}
			if err := createGitHubWebhook(githubToken, owner, repoName, webhookURL); err != nil {
				result.Steps["webhook"] = fmt.Sprintf("error: %v", err)
			} else {
				result.Steps["webhook"] = "created"
				// Désactiver la protection de branche pour que le SE puisse git push
				if err := disableGitHubBranchProtection(githubToken, owner, repoName, "main"); err != nil {
					result.Steps["branch_protection"] = fmt.Sprintf("warning: %v", err)
				} else {
					result.Steps["branch_protection"] = "disabled"
				}
			}
		}
	}

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

func parseGitHubRepo(repoURL string) (owner, repo string, err error) {
	u := repoURL
	for _, prefix := range []string{"https://github.com/", "http://github.com/", "github.com/"} {
		u = strings.TrimPrefix(u, prefix)
	}
	u = strings.TrimSuffix(u, ".git")
	parts := strings.Split(u, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("expected github.com/owner/repo, got: %s", repoURL)
	}
	return parts[0], parts[1], nil
}

func createGitHubWebhook(token, owner, repo, webhookURL string) error {
	payload, _ := json.Marshal(map[string]interface{}{
		"name":   "web",
		"active": true,
		"events": []string{"push"},
		"config": map[string]string{
			"url":          webhookURL,
			"content_type": "json",
			"insecure_ssl": "0",
		},
	})
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/hooks", owner, repo)
	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// 201 = créé, 422 = déjà existant → les deux sont ok
	if resp.StatusCode == 201 || resp.StatusCode == 422 {
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("github api %d: %s", resp.StatusCode, string(body))
}

func disableGitHubBranchProtection(token, owner, repo, branch string) error {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches/%s/protection", owner, repo, branch)
	req, err := http.NewRequest("DELETE", apiURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// 204 = supprimée, 404 = pas de protection active → les deux sont ok
	if resp.StatusCode == 204 || resp.StatusCode == 404 {
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("github api %d: %s", resp.StatusCode, string(body))
}

func createGitHubRepo(token, owner, repoName string) error {
	apiURL := "https://api.github.com/user/repos"
	payload, _ := json.Marshal(map[string]interface{}{
		"name":        repoName,
		"private":     false,
		"auto_init":   false,
		"description": "Created by DxP",
	})
	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// 201 = créé, 422 = déjà existant → ok
	if resp.StatusCode == 201 || resp.StatusCode == 422 {
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("github api %d: %s", resp.StatusCode, string(body))
}

// handleProvisionArgoCD crée une app ArgoCD pour un service déjà buildé
// Appelé par Tekton après le build — le repo est non vide à ce stade
func (s *Server) handleProvisionArgoCD(w http.ResponseWriter, r *http.Request) {
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

	// Créer l'app ArgoCD via backend C2
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
		result.Error = err.Error()
		writeJSON(w, http.StatusInternalServerError, result)
		return
	}

	if !actionResult.Success {
		result.Steps["argocd_app"] = fmt.Sprintf("error: %s", actionResult.Message)
		result.Error = actionResult.Message
		writeJSON(w, http.StatusInternalServerError, result)
		return
	}

	result.Steps["argocd_app"] = "created"
	writeJSON(w, http.StatusOK, result)
}
