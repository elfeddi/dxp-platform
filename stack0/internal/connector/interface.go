package connector

import "context"

// Connector est l'interface que tout connecteur DxP doit implémenter
type Connector interface {
	// Install installe le composant
	Install(ctx context.Context) error

	// Configure applique la configuration post-installation
	Configure(ctx context.Context) error

	// HealthCheck vérifie que le composant est opérationnel
	HealthCheck(ctx context.Context) (bool, error)

	// GetStatus retourne l'état courant du composant
	GetStatus(ctx context.Context) (map[string]interface{}, error)

	// Uninstall désinstalle le composant
	Uninstall(ctx context.Context) error

	// Name retourne le nom du connecteur
	Name() string

	// Type retourne le type du connecteur (argocd, harbor, tekton...)
	Type() string
}
