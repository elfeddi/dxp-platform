package harbor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/elfeddi/dxp/internal/gateway"
)

// Backend implémente BackendConnector pour Harbor.
// Auth : Authorization: Robot <token>
// Endpoint : http://localhost:9091 (NodePort interne VM→k3d)
type Backend struct {
	baseURL string
	token   string
	client  *http.Client
}

func New(baseURL, token string) *Backend {
	return &Backend{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (b *Backend) Name() string { return "harbor" }

func (b *Backend) GetStatus(ctx context.Context) (*gateway.ComponentStatus, error) {
	resp, err := b.get(ctx, "/api/v2.0/systeminfo")
	if err != nil {
		return &gateway.ComponentStatus{
			Name: b.Name(), Healthy: false, Ready: false,
			Message: err.Error(), UpdatedAt: time.Now().Unix(),
		}, nil
	}
	defer resp.Body.Close()

	var info struct {
		AuthMode      string `json:"auth_mode"`
		HarborVersion string `json:"harbor_version"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&info)

	healthy := resp.StatusCode == http.StatusOK
	version := info.HarborVersion
	if version == "" {
		version = "v2.15.x"
	}

	return &gateway.ComponentStatus{
		Name:      b.Name(),
		Healthy:   healthy,
		Ready:     healthy,
		Version:   version,
		Message:   fmt.Sprintf("Harbor operational — auth: %s", info.AuthMode),
		UpdatedAt: time.Now().Unix(),
	}, nil
}

func (b *Backend) GetMetrics(ctx context.Context) (string, error) {
	resp, err := b.get(ctx, "/metrics")
	if err != nil {
		return "", fmt.Errorf("harbor: metrics unavailable: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("harbor: failed to read metrics: %w", err)
	}
	return string(body), nil
}

func (b *Backend) GetLogs(_ context.Context, req *gateway.LogRequest) ([]*gateway.LogEntry, error) {
	return []*gateway.LogEntry{{
		Timestamp: time.Now().UnixMilli(),
		Level:     "info",
		Message:   fmt.Sprintf("Harbor logs stub — connect C5 to Loki (requested: %d lines)", req.Lines),
		Source:    "harbor/gateway-stub",
	}}, nil
}

func (b *Backend) ExecuteAction(ctx context.Context, req *gateway.ActionRequest) (*gateway.ActionResult, error) {
	switch req.Action {
	case "gc":
		return b.triggerGC(ctx, req)
	default:
		return &gateway.ActionResult{
			Success: false,
			Message: fmt.Sprintf("unknown action: %s", req.Action),
		}, nil
	}
}

func (b *Backend) ListActions(_ context.Context) ([]*gateway.ActionDescriptor, error) {
	return []*gateway.ActionDescriptor{
		{
			Name:        "gc",
			Description: "Déclenche un Garbage Collection sur Harbor.",
			Dangerous:   false,
		},
	}, nil
}

func (b *Backend) triggerGC(ctx context.Context, req *gateway.ActionRequest) (*gateway.ActionResult, error) {
	if req.DryRun {
		return &gateway.ActionResult{Success: true, Message: "[dry-run] would trigger Harbor GC"}, nil
	}
	body := strings.NewReader(`{"schedule":{"type":"Manual"}}`)
	resp, err := b.post(ctx, "/api/v2.0/system/gc/schedule", body)
	if err != nil {
		return &gateway.ActionResult{Success: false, Message: err.Error()}, nil
	}
	defer resp.Body.Close()
	return &gateway.ActionResult{
		Success: resp.StatusCode < 300,
		Message: fmt.Sprintf("Harbor GC triggered (status %d)", resp.StatusCode),
	}, nil
}

func (b *Backend) get(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Robot "+b.token)
	return b.client.Do(req)
}

func (b *Backend) post(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Robot "+b.token)
	req.Header.Set("Content-Type", "application/json")
	return b.client.Do(req)
}
