package resolver

import (
	"testing"

	"github.com/elfeddi/dxp/pkg/types"
)

func makeConfigWithDeps() *types.DxPConfig {
	return &types.DxPConfig{
		Version: "1.0",
		Name:    "test",
		Stacks: []types.Stack{
			{
				Name:    "devops",
				Enabled: true,
				Connectors: []types.Connector{
					{Name: "argocd", Type: "argocd"},
					{Name: "harbor", Type: "harbor", DependsOn: []string{"argocd"}},
					{Name: "tekton", Type: "tekton", DependsOn: []string{"harbor"}},
				},
			},
		},
	}
}

func TestDAGResolve_NoDeps(t *testing.T) {
	cfg := &types.DxPConfig{
		Version: "1.0",
		Name:    "test",
		Stacks: []types.Stack{
			{
				Name:    "devops",
				Enabled: true,
				Connectors: []types.Connector{
					{Name: "argocd", Type: "argocd"},
					{Name: "harbor", Type: "harbor"},
				},
			},
		},
	}
	d := NewDAGResolver()
	ordered, err := d.Resolve(cfg)
	if err != nil {
		t.Fatalf("Resolve erreur inattendue: %v", err)
	}
	if len(ordered) != 2 {
		t.Errorf("2 connecteurs attendus, obtenus %d", len(ordered))
	}
}

func TestDAGResolve_WithDeps(t *testing.T) {
	cfg := makeConfigWithDeps()
	d := NewDAGResolver()
	ordered, err := d.Resolve(cfg)
	if err != nil {
		t.Fatalf("Resolve erreur inattendue: %v", err)
	}
	if len(ordered) != 3 {
		t.Errorf("3 connecteurs attendus, obtenus %d", len(ordered))
	}
	// argocd doit être avant harbor
	pos := map[string]int{}
	for i, c := range ordered {
		pos[c.Name] = i
	}
	if pos["argocd"] > pos["harbor"] {
		t.Error("argocd doit être installé avant harbor")
	}
	if pos["harbor"] > pos["tekton"] {
		t.Error("harbor doit être installé avant tekton")
	}
}

func TestDAGResolve_UnknownDep(t *testing.T) {
	cfg := &types.DxPConfig{
		Version: "1.0",
		Name:    "test",
		Stacks: []types.Stack{
			{
				Name:    "devops",
				Enabled: true,
				Connectors: []types.Connector{
					{Name: "harbor", Type: "harbor", DependsOn: []string{"inexistant"}},
				},
			},
		},
	}
	d := NewDAGResolver()
	_, err := d.Resolve(cfg)
	if err == nil {
		t.Error("Erreur attendue pour dépendance inconnue")
	}
}

func TestDAGResolve_DisabledStack(t *testing.T) {
	cfg := &types.DxPConfig{
		Version: "1.0",
		Name:    "test",
		Stacks: []types.Stack{
			{Name: "devops", Enabled: false, Connectors: []types.Connector{
				{Name: "argocd", Type: "argocd"},
			}},
		},
	}
	d := NewDAGResolver()
	ordered, err := d.Resolve(cfg)
	if err != nil {
		t.Fatalf("Resolve erreur inattendue: %v", err)
	}
	if len(ordered) != 0 {
		t.Errorf("0 connecteur attendu (stack disabled), obtenu %d", len(ordered))
	}
}
