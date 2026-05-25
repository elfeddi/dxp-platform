package resolver

import (
	"testing"

	"github.com/elfeddi/dxp/pkg/types"
)

func makeValidConfig() *types.DxPConfig {
	return &types.DxPConfig{
		Version:  "1.0",
		Name:     "test",
		Provider: "azure",
		Stacks: []types.Stack{
			{
				Name:    "devops",
				Enabled: true,
				Connectors: []types.Connector{
					{Type: "argocd", Name: "argocd-main"},
				},
			},
		},
	}
}

func TestValidate_Valid(t *testing.T) {
	v := NewValidator()
	errs := v.Validate(makeValidConfig())
	if len(errs) != 0 {
		t.Errorf("Aucune erreur attendue, obtenu %d: %v", len(errs), errs)
	}
}

func TestValidate_MissingVersion(t *testing.T) {
	cfg := makeValidConfig()
	cfg.Version = ""
	v := NewValidator()
	errs := v.Validate(cfg)
	if len(errs) == 0 {
		t.Error("Erreur attendue pour version manquante")
	}
	found := false
	for _, e := range errs {
		if ve, ok := e.(*ValidationError); ok && ve.Field == "version" {
			found = true
		}
	}
	if !found {
		t.Error("Erreur attendue sur le champ version")
	}
}

func TestValidate_MissingName(t *testing.T) {
	cfg := makeValidConfig()
	cfg.Name = ""
	v := NewValidator()
	errs := v.Validate(cfg)
	if len(errs) == 0 {
		t.Error("Erreur attendue pour name manquant")
	}
}

func TestValidate_EmptyStacks(t *testing.T) {
	cfg := makeValidConfig()
	cfg.Stacks = []types.Stack{}
	v := NewValidator()
	errs := v.Validate(cfg)
	if len(errs) == 0 {
		t.Error("Erreur attendue pour stacks vides")
	}
}

func TestValidate_MissingStackName(t *testing.T) {
	cfg := makeValidConfig()
	cfg.Stacks[0].Name = ""
	v := NewValidator()
	errs := v.Validate(cfg)
	if len(errs) == 0 {
		t.Error("Erreur attendue pour stack name manquant")
	}
}

func TestValidate_MissingConnectorType(t *testing.T) {
	cfg := makeValidConfig()
	cfg.Stacks[0].Connectors[0].Type = ""
	v := NewValidator()
	errs := v.Validate(cfg)
	if len(errs) == 0 {
		t.Error("Erreur attendue pour connector type manquant")
	}
}

func TestValidate_MultipleErrors(t *testing.T) {
	cfg := &types.DxPConfig{}
	v := NewValidator()
	errs := v.Validate(cfg)
	if len(errs) < 2 {
		t.Errorf("Au moins 2 erreurs attendues, obtenu %d", len(errs))
	}
}

func TestValidationError_Message(t *testing.T) {
	e := &ValidationError{Field: "name", Message: "champ obligatoire"}
	msg := e.Error()
	if msg == "" {
		t.Error("ValidationError.Error() ne doit pas être vide")
	}
}
