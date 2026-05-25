package connector

import (
	"context"
	"testing"
)

// mockConnector implémente Connector pour les tests
type mockConnector struct {
	connName string
	connType string
}

func (m *mockConnector) Install(ctx context.Context) error                        { return nil }
func (m *mockConnector) Configure(ctx context.Context) error                      { return nil }
func (m *mockConnector) HealthCheck(ctx context.Context) (bool, error)            { return true, nil }
func (m *mockConnector) GetStatus(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{"status": "ok"}, nil
}
func (m *mockConnector) Uninstall(ctx context.Context) error { return nil }
func (m *mockConnector) Name() string                        { return m.connName }
func (m *mockConnector) Type() string                        { return m.connType }

func TestRegistry_RegisterAndCreate(t *testing.T) {
	r := NewRegistry()
	r.Register("argocd", func(cfg map[string]string) (Connector, error) {
		return &mockConnector{connName: "argocd-main", connType: "argocd"}, nil
	})

	c, err := r.Create("argocd", map[string]string{"url": "http://localhost:9090"})
	if err != nil {
		t.Fatalf("Create erreur inattendue: %v", err)
	}
	if c == nil {
		t.Error("Connector attendu, obtenu nil")
	}
	if c.Type() != "argocd" {
		t.Errorf("Type attendu %q, obtenu %q", "argocd", c.Type())
	}
}

func TestRegistry_CreateUnknown(t *testing.T) {
	r := NewRegistry()
	_, err := r.Create("inexistant", map[string]string{})
	if err == nil {
		t.Error("Erreur attendue pour connector inconnu")
	}
}

func TestRegistry_ListTypes(t *testing.T) {
	r := NewRegistry()
	types := []string{"argocd", "harbor", "grafana", "tekton", "vault"}
	for _, ct := range types {
		ctCopy := ct
		r.Register(ctCopy, func(cfg map[string]string) (Connector, error) {
			return &mockConnector{connName: ctCopy + "-main", connType: ctCopy}, nil
		})
	}

	listed := r.ListTypes()
	if len(listed) != len(types) {
		t.Errorf("%d types attendus, obtenus %d", len(types), len(listed))
	}
}

func TestRegistry_RegisterMultipleCreate(t *testing.T) {
	r := NewRegistry()
	r.Register("harbor", func(cfg map[string]string) (Connector, error) {
		return &mockConnector{connName: cfg["name"], connType: "harbor"}, nil
	})

	c, err := r.Create("harbor", map[string]string{"name": "harbor-main", "url": "http://localhost:9091"})
	if err != nil {
		t.Fatalf("Create erreur inattendue: %v", err)
	}
	if c.Name() != "harbor-main" {
		t.Errorf("Name attendu %q, obtenu %q", "harbor-main", c.Name())
	}
}
