package observer

import (
	"time"
)

// StackStatus représente létat dune stack DxP.
type StackStatus struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	Healthy bool   `json:"healthy"`
}

// PlatformStatus est un snapshot complet de létat de la plateforme.
type PlatformStatus struct {
	Platform  string                   `json:"platform"`
	Version   string                   `json:"version"`
	Healthy   bool                     `json:"healthy"`
	Ready     bool                     `json:"ready"`
	Backends  map[string]*HealthStatus `json:"backends"`
	Stacks    []StackStatus            `json:"stacks"`
	UpdatedAt int64                    `json:"updated_at"`
}

// StatusAggregator agrège létat complet de la plateforme depuis C5.
type StatusAggregator struct {
	checker *HealthChecker
	version string
}

// NewStatusAggregator crée un StatusAggregator.
func NewStatusAggregator(checker *HealthChecker, version string) *StatusAggregator {
	return &StatusAggregator{checker: checker, version: version}
}

// Get retourne le statut agrégé courant.
func (s *StatusAggregator) Get() *PlatformStatus {
	health := s.checker.Get()

	stacks := []StackStatus{
		{Name: "devops", Enabled: true, Healthy: health.Healthy},
		{Name: "observability", Enabled: true, Healthy: true},
		{Name: "security", Enabled: true, Healthy: true},
		{Name: "dataops", Enabled: false, Healthy: false},
		{Name: "mlops", Enabled: false, Healthy: false},
		{Name: "llmops", Enabled: false, Healthy: false},
	}

	return &PlatformStatus{
		Platform:  "DxP",
		Version:   s.version,
		Healthy:   health.Healthy,
		Ready:     health.Ready,
		Backends:  health.Backends,
		Stacks:    stacks,
		UpdatedAt: time.Now().Unix(),
	}
}
