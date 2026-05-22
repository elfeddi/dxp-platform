package resolver

import (
	"fmt"

	"github.com/elfeddi/dxp/pkg/types"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error [%s]: %s", e.Field, e.Message)
}

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) Validate(config *types.DxPConfig) []error {
	var errs []error

	if config.Version == "" {
		errs = append(errs, &ValidationError{"version", "champ obligatoire"})
	}
	if config.Name == "" {
		errs = append(errs, &ValidationError{"name", "champ obligatoire"})
	}
	if len(config.Stacks) == 0 {
		errs = append(errs, &ValidationError{"stacks", "au moins une stack requise"})
	}
	for i, stack := range config.Stacks {
		if stack.Name == "" {
			errs = append(errs, &ValidationError{
				fmt.Sprintf("stacks[%d].name", i), "champ obligatoire",
			})
		}
		for j, connector := range stack.Connectors {
			if connector.Type == "" {
				errs = append(errs, &ValidationError{
					fmt.Sprintf("stacks[%d].connectors[%d].type", i, j),
					"champ obligatoire",
				})
			}
		}
	}
	return errs
}
