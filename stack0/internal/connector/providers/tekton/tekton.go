package tekton

import (
	"context"
	"fmt"
)

type TektonConnector struct {
	name      string
	namespace string
	kubeconfig string
}

func New(config map[string]string) (*TektonConnector, error) {
	return &TektonConnector{
		name:       config["name"],
		namespace:  config["namespace"],
		kubeconfig: config["kubeconfig"],
	}, nil
}

func (t *TektonConnector) Name() string { return t.name }
func (t *TektonConnector) Type() string { return "tekton" }

func (t *TektonConnector) Install(ctx context.Context) error {
	return fmt.Errorf("not implemented: kubectl apply -f tekton-releases")
}

func (t *TektonConnector) Configure(ctx context.Context) error   { return nil }
func (t *TektonConnector) Uninstall(ctx context.Context) error   { return fmt.Errorf("not implemented") }

func (t *TektonConnector) HealthCheck(ctx context.Context) (bool, error) {
	return true, nil
}

func (t *TektonConnector) GetStatus(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{"type": "tekton", "name": t.name}, nil
}
