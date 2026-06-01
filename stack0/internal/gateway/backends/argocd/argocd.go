package argocd

import (
	"bytes"
	"encoding/json"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/elfeddi/dxp/internal/gateway"
)

// Backend implémente BackendConnector pour ArgoCD.
// Utilise un token API statique — généré via : argocd account generate-token
// Accès via nginx Ingress HTTPS avec cert self-signed → InsecureSkipVerify.
type Backend struct {
	baseURL string
	token   string
	client  *http.Client
}

func New(baseURL, token string) *Backend {
	return &Backend{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

func (b *Backend) Name() string { return "argocd" }

// GetStatus utilise /api/v1/applications pour vérifier qu'ArgoCD répond.
func (b *Backend) GetStatus(ctx context.Context) (*gateway.ComponentStatus, error) {
	resp, err := b.get(ctx, "/api/v1/applications?limit=1")
	if err != nil {
		return &gateway.ComponentStatus{
			Name: b.Name(), Healthy: false, Ready: false,
			Message: err.Error(), UpdatedAt: time.Now().Unix(),
		}, nil
	}
	defer resp.Body.Close()

	healthy := resp.StatusCode == http.StatusOK
	msg := "ArgoCD operational"
	if !healthy {
		body, _ := io.ReadAll(resp.Body)
		msg = fmt.Sprintf("ArgoCD returned %d: %s", resp.StatusCode, string(body))
	}

	return &gateway.ComponentStatus{
		Name:      b.Name(),
		Healthy:   healthy,
		Ready:     healthy,
		Version:   "v3.4.x",
		Message:   msg,
		UpdatedAt: time.Now().Unix(),
	}, nil
}

func (b *Backend) GetMetrics(ctx context.Context) (string, error) {
	resp, err := b.get(ctx, "/metrics")
	if err != nil {
		return "", fmt.Errorf("argocd: metrics unavailable: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("argocd: failed to read metrics: %w", err)
	}
	return string(body), nil
}

func (b *Backend) GetLogs(_ context.Context, req *gateway.LogRequest) ([]*gateway.LogEntry, error) {
	return []*gateway.LogEntry{{
		Timestamp: time.Now().UnixMilli(),
		Level:     "info",
		Message:   fmt.Sprintf("ArgoCD logs stub — connect C5 to Loki (requested: %d lines)", req.Lines),
		Source:    "argocd/gateway-stub",
	}}, nil
}

func (b *Backend) ExecuteAction(ctx context.Context, req *gateway.ActionRequest) (*gateway.ActionResult, error) {
	switch req.Action {
	case "sync":
		return b.syncApp(ctx, req)
	case "refresh":
		return b.refreshApp(ctx, req)
	case "create-app":
		return b.createApp(ctx, req)
	default:
		return &gateway.ActionResult{
			Success: false,
			Message: fmt.Sprintf("unknown action: %s", req.Action),
		}, nil
	}
}

func (b *Backend) ListActions(_ context.Context) ([]*gateway.ActionDescriptor, error) {
	return []*gateway.ActionDescriptor{
		{Name: "sync", Description: "Synchronise une application ArgoCD avec le repo Git.",
			Params: map[string]string{"app": "nom de l'application ArgoCD"}, Dangerous: false},
		{Name: "refresh", Description: "Force le refresh d'une application depuis le repo Git.",
			Params: map[string]string{"app": "nom de l'application ArgoCD"}, Dangerous: false},
	}, nil
}

func (b *Backend) syncApp(ctx context.Context, req *gateway.ActionRequest) (*gateway.ActionResult, error) {
	appName := req.Target
	if appName == "" {
		appName = req.Params["app"]
	}
	if appName == "" {
		return &gateway.ActionResult{Success: false, Message: "missing target app name"}, nil
	}
	if req.DryRun {
		return &gateway.ActionResult{Success: true,
			Message: fmt.Sprintf("[dry-run] would sync ArgoCD app: %s", appName)}, nil
	}
	resp, err := b.post(ctx, fmt.Sprintf("/api/v1/applications/%s/sync", appName), nil)
	if err != nil {
		return &gateway.ActionResult{Success: false, Message: err.Error()}, nil
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return &gateway.ActionResult{Success: true,
			Message: fmt.Sprintf("ArgoCD app %q sync triggered", appName)}, nil
	}
	body, _ := io.ReadAll(resp.Body)
	return &gateway.ActionResult{Success: false,
		Message: fmt.Sprintf("sync failed (%d): %s", resp.StatusCode, string(body))}, nil
}

func (b *Backend) refreshApp(ctx context.Context, req *gateway.ActionRequest) (*gateway.ActionResult, error) {
	appName := req.Target
	if appName == "" {
		appName = req.Params["app"]
	}
	if appName == "" {
		return &gateway.ActionResult{Success: false, Message: "missing target app name"}, nil
	}
	if req.DryRun {
		return &gateway.ActionResult{Success: true,
			Message: fmt.Sprintf("[dry-run] would refresh ArgoCD app: %s", appName)}, nil
	}
	resp, err := b.get(ctx, fmt.Sprintf("/api/v1/applications/%s?refresh=normal", appName))
	if err != nil {
		return &gateway.ActionResult{Success: false, Message: err.Error()}, nil
	}
	defer resp.Body.Close()
	return &gateway.ActionResult{
		Success: resp.StatusCode < 300,
		Message: fmt.Sprintf("ArgoCD app %q refresh triggered", appName),
	}, nil
}

func (b *Backend) get(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+b.token)
	return b.client.Do(req)
}

func (b *Backend) post(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+b.token)
	req.Header.Set("Content-Type", "application/json")
	return b.client.Do(req)
}

func (b *Backend) createApp(ctx context.Context, req *gateway.ActionRequest) (*gateway.ActionResult, error) {
	name := req.Target
	repo := req.Params["repo"]
	namespace := req.Params["namespace"]
	path := req.Params["path"]
	if path == "" {
		path = "."
	}
	path = "."

	body := map[string]interface{}{
		"apiVersion": "argoproj.io/v1alpha1",
		"kind":       "Application",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": "argocd",
			"labels":    map[string]interface{}{"dxp.io/managed": "true"},
		},
		"spec": map[string]interface{}{
			"project": "default",
			"source": map[string]interface{}{
				"repoURL":        repo,
				"targetRevision": "main",
				"path":           path,
			},
			"destination": map[string]interface{}{
				"server":    "https://kubernetes.default.svc",
				"namespace": namespace,
			},
			"syncPolicy": map[string]interface{}{
				"automated": map[string]interface{}{
					"prune":    true,
					"selfHeal": true,
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return &gateway.ActionResult{Success: false, Message: err.Error()}, nil
	}

	resp, err := b.post(ctx, "/api/v1/applications", bytes.NewReader(bodyBytes))
	if err != nil {
		return &gateway.ActionResult{Success: false, Message: err.Error()}, nil
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
		return &gateway.ActionResult{
			Success: true,
			Message: fmt.Sprintf("ArgoCD app %q created", name),
		}, nil
	}

	return &gateway.ActionResult{
		Success: false,
		Message: fmt.Sprintf("ArgoCD API returned %d: %s", resp.StatusCode, string(respBody)),
	}, nil
}
