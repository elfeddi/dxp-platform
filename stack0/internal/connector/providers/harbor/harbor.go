package harbor

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type HarborConnector struct {
	name      string
	url       string
	username  string
	password  string
	namespace string
	client    *http.Client
}

func New(config map[string]string) (*HarborConnector, error) {
	url, ok := config["url"]
	if !ok {
		return nil, fmt.Errorf("harbor: url obligatoire")
	}
	return &HarborConnector{
		name:      config["name"],
		url:       url,
		username:  config["username"],
		password:  config["password"],
		namespace: config["namespace"],
		client:    &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (h *HarborConnector) Name() string { return h.name }
func (h *HarborConnector) Type() string { return "harbor" }

func (h *HarborConnector) Install(ctx context.Context) error {
	return fmt.Errorf("not implemented: utiliser helm install harbor harbor/harbor")
}

func (h *HarborConnector) Configure(ctx context.Context) error {
	return nil
}

func (h *HarborConnector) HealthCheck(ctx context.Context) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		h.url+"/api/v2.0/ping", nil)
	if err != nil {
		return false, err
	}
	req.SetBasicAuth(h.username, h.password)
	resp, err := h.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}

func (h *HarborConnector) GetStatus(ctx context.Context) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		h.url+"/api/v2.0/systeminfo", nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(h.username, h.password)
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return map[string]interface{}{
		"status": resp.StatusCode,
		"url":    h.url,
	}, nil
}

func (h *HarborConnector) Uninstall(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
