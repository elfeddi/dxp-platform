package gateway

import (
	"fmt"
	"sync"
)

// Registry maintient la liste des BackendConnectors enregistrés.
// Thread-safe — peut être interrogé depuis plusieurs goroutines HTTP.
type Registry struct {
	mu       sync.RWMutex
	backends map[string]BackendConnector
}

// NewRegistry crée un registry vide.
func NewRegistry() *Registry {
	return &Registry{
		backends: make(map[string]BackendConnector),
	}
}

// Register enregistre un backend. Retourne une erreur si le nom est déjà pris.
func (r *Registry) Register(b BackendConnector) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := b.Name()
	if _, exists := r.backends[name]; exists {
		return fmt.Errorf("gateway: backend %q already registered", name)
	}
	r.backends[name] = b
	return nil
}

// MustRegister comme Register mais panique en cas d'erreur.
func (r *Registry) MustRegister(b BackendConnector) {
	if err := r.Register(b); err != nil {
		panic(err)
	}
}

// Get retourne un backend par son nom.
func (r *Registry) Get(name string) (BackendConnector, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	b, ok := r.backends[name]
	if !ok {
		return nil, fmt.Errorf("gateway: backend %q not found", name)
	}
	return b, nil
}

// List retourne tous les noms de backends enregistrés.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.backends))
	for name := range r.backends {
		names = append(names, name)
	}
	return names
}

// All retourne tous les backends (copie de la map).
func (r *Registry) All() map[string]BackendConnector {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make(map[string]BackendConnector, len(r.backends))
	for k, v := range r.backends {
		out[k] = v
	}
	return out
}

// Unregister supprime un backend du registry.
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.backends, name)
}
