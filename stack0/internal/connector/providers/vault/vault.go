package vault

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type VaultConnector struct {
	name   string
	url    string
	token  string
	client *http.Client
}

func New(config map[string]string) (*VaultConnector, error) {
	url, ok := config["url"]
	if !ok {
		return nil, fmt.Errorf("vault: url obligatoire")
	}
	return &VaultConnector{
		name:   config["name"],
		url:    url,
		token:  config["token"],
		client: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (v *VaultConnector) Name() string { return v.name }
func (v *VaultConnector) Type() string { return "vault" }

func (v *VaultConnector) Install(ctx context.Context) error {
	return fmt.Errorf("not implemented: helm install vault hashicorp/vault")
}

func (v *VaultConnector) Configure(ctx context.Context) error   { return nil }
func (v *VaultConnector) Uninstall(ctx context.Context) error   { return fmt.Errorf("not implemented") }

func (v *VaultConnector) HealthCheck(ctx context.Context) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", v.url+"/v1/sys/health", nil)
	if err != nil {
		return false, err
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}

func (v *VaultConnector) GetStatus(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{"type": "vault", "url": v.url}, nil
}
