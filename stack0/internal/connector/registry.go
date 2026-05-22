package connector

import (
	"fmt"
	"sync"
)

// Factory crée un connecteur à partir d'un type et d'une config
type Factory func(config map[string]string) (Connector, error)

// Registry gère l'enregistrement et la création des connecteurs
type Registry struct {
	mu        sync.RWMutex
	factories map[string]Factory
}

func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]Factory),
	}
}

// Register enregistre une factory pour un type de connecteur
func (r *Registry) Register(connectorType string, factory Factory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[connectorType] = factory
}

// Create crée un connecteur à partir de son type et de sa config
func (r *Registry) Create(connectorType string, config map[string]string) (Connector, error) {
	r.mu.RLock()
	factory, ok := r.factories[connectorType]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("connecteur inconnu: %s — types disponibles: %v",
			connectorType, r.ListTypes())
	}
	return factory(config)
}

// ListTypes retourne la liste des types enregistrés
func (r *Registry) ListTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	types := make([]string, 0, len(r.factories))
	for t := range r.factories {
		types = append(types, t)
	}
	return types
}
