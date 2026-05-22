package connector

import (
	"github.com/elfeddi/dxp/internal/connector/providers/argocd"
	"github.com/elfeddi/dxp/internal/connector/providers/harbor"
	"github.com/elfeddi/dxp/internal/connector/providers/tekton"
	"github.com/elfeddi/dxp/internal/connector/providers/vault"
)

// NewDefaultRegistry crée un registry avec tous les connecteurs enregistrés
func NewDefaultRegistry() *Registry {
	r := NewRegistry()

	r.Register("argocd", func(config map[string]string) (Connector, error) {
		return argocd.New(config)
	})

	r.Register("harbor", func(config map[string]string) (Connector, error) {
		return harbor.New(config)
	})

	r.Register("tekton", func(config map[string]string) (Connector, error) {
		return tekton.New(config)
	})

	r.Register("vault", func(config map[string]string) (Connector, error) {
		return vault.New(config)
	})

	return r
}
