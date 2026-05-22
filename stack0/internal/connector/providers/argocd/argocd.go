package argocd

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// ArgoCDConnector implémente connector.Connector pour ArgoCD
type ArgoCDConnector struct {
	name      string
	url       string
	token     string
	namespace string
	client    *http.Client
}

func New(config map[string]string) (*ArgoCDConnector, error) {
	url, ok := config["url"]
	if !ok {
		return nil, fmt.Errorf("argocd: url obligatoire")
	}
	return &ArgoCDConnector{
		name:      config["name"],
		url:       url,
		token:     config["token"],
		namespace: config["namespace"],
		client:    &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (a *ArgoCDConnector) Name() string { return a.name }
func (a *ArgoCDConnector) Type() string { return "argocd" }

func (a *ArgoCDConnector) Install(ctx context.Context) error {
	// Installation via Helm SDK — implémentation Phase 2
	return fmt.Errorf("not implemented: utiliser helm install argocd argo/argo-cd")
}

func (a *ArgoCDConnector) Configure(ctx context.Context) error {
	// Configuration post-installation (admin password, apiKey capability)
	return nil
}

func (a *ArgoCDConnector) HealthCheck(ctx context.Context) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		a.url+"/healthz", nil)
	if err != nil {
		return false, err
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}

func (a *ArgoCDConnector) GetStatus(ctx context.Context) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		a.url+"/api/v1/applications", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+a.token)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return map[string]interface{}{
		"status": resp.StatusCode,
		"url":    a.url,
	}, nil
}

func (a *ArgoCDConnector) Uninstall(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
