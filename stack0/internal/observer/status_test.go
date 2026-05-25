package observer

import (
	"testing"
	"time"
)

func makeHealthChecker(healthy bool) *HealthChecker {
	hc := &HealthChecker{}
	hc.cache = &PlatformHealth{
		Healthy:   healthy,
		Ready:     healthy,
		Backends:  map[string]*HealthStatus{
			"argocd": {Name: "argocd", Healthy: healthy, Ready: healthy, CheckedAt: time.Now().Unix()},
		},
		CheckedAt: time.Now().Unix(),
	}
	return hc
}

func TestStatusAggregator_Get_Healthy(t *testing.T) {
	checker := makeHealthChecker(true)
	agg := NewStatusAggregator(checker, "v0.2.0")
	status := agg.Get()

	if status.Platform != "DxP" {
		t.Errorf("Platform attendu %q, obtenu %q", "DxP", status.Platform)
	}
	if status.Version != "v0.2.0" {
		t.Errorf("Version attendue %q, obtenue %q", "v0.2.0", status.Version)
	}
	if !status.Healthy {
		t.Error("Status devrait être healthy")
	}
	if len(status.Stacks) == 0 {
		t.Error("Stacks ne devrait pas être vide")
	}
	if status.UpdatedAt == 0 {
		t.Error("UpdatedAt devrait être défini")
	}
}

func TestStatusAggregator_Get_Unhealthy(t *testing.T) {
	checker := makeHealthChecker(false)
	agg := NewStatusAggregator(checker, "v0.2.0")
	status := agg.Get()

	if status.Healthy {
		t.Error("Status devrait être unhealthy")
	}
}

func TestStatusAggregator_Stacks(t *testing.T) {
	checker := makeHealthChecker(true)
	agg := NewStatusAggregator(checker, "v0.2.0")
	status := agg.Get()

	// Vérifier que les stacks actives sont présentes
	activeCount := 0
	for _, s := range status.Stacks {
		if s.Enabled {
			activeCount++
		}
	}
	if activeCount == 0 {
		t.Error("Au moins une stack active attendue")
	}
}

func TestStackStatus_Fields(t *testing.T) {
	ss := StackStatus{Name: "devops", Enabled: true, Healthy: true}
	if ss.Name != "devops" {
		t.Errorf("Name attendu %q, obtenu %q", "devops", ss.Name)
	}
	if !ss.Enabled {
		t.Error("Enabled devrait être true")
	}
}
