package resolver

import (
	"testing"
)

func TestParseBytes_Valid(t *testing.T) {
	yaml := []byte(`
version: "1.0"
name: test-projet
provider: azure
stacks:
  - name: devops
    enabled: true
    connectors:
      - type: argocd
        name: argocd-main
        config:
          url: http://localhost:9090
`)
	p := NewParser()
	cfg, err := p.ParseBytes(yaml)
	if err != nil {
		t.Fatalf("ParseBytes erreur inattendue: %v", err)
	}
	if cfg.Name != "test-projet" {
		t.Errorf("Name attendu %q, obtenu %q", "test-projet", cfg.Name)
	}
	if cfg.Version != "1.0" {
		t.Errorf("Version attendue %q, obtenue %q", "1.0", cfg.Version)
	}
	if len(cfg.Stacks) != 1 {
		t.Errorf("Stacks attendues 1, obtenues %d", len(cfg.Stacks))
	}
	if len(cfg.Stacks[0].Connectors) != 1 {
		t.Errorf("Connectors attendus 1, obtenus %d", len(cfg.Stacks[0].Connectors))
	}
}

func TestParseBytes_InvalidYAML(t *testing.T) {
	p := NewParser()
	_, err := p.ParseBytes([]byte("not: valid: yaml: :::"))
	if err == nil {
		t.Error("ParseBytes devrait retourner une erreur sur YAML invalide")
	}
}

func TestParseBytes_EmptyConfig(t *testing.T) {
	p := NewParser()
	cfg, err := p.ParseBytes([]byte(""))
	if err != nil {
		t.Fatalf("ParseBytes erreur inattendue sur config vide: %v", err)
	}
	if cfg.Name != "" {
		t.Errorf("Name attendu vide, obtenu %q", cfg.Name)
	}
}

func TestParseBytes_MultipleStacks(t *testing.T) {
	yaml := []byte(`
version: "1.0"
name: multi-stack
provider: azure
stacks:
  - name: devops
    enabled: true
    connectors:
      - type: argocd
        name: argocd-main
  - name: dataops
    enabled: false
    connectors:
      - type: airflow
        name: airflow-main
`)
	p := NewParser()
	cfg, err := p.ParseBytes(yaml)
	if err != nil {
		t.Fatalf("ParseBytes erreur inattendue: %v", err)
	}
	if len(cfg.Stacks) != 2 {
		t.Errorf("2 stacks attendues, obtenues %d", len(cfg.Stacks))
	}
	if cfg.Stacks[1].Enabled {
		t.Error("Stack dataops devrait être disabled")
	}
}
