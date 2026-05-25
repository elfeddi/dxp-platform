package observer

import (
	"strings"
	"testing"
)

// mockChecker simule un HealthChecker pour les tests
type mockChecker struct {
	health *PlatformHealth
}

func (m *mockChecker) Get() *PlatformHealth {
	return m.health
}

func newMockCollector(healthy bool, backends map[string]*HealthStatus) *MetricsCollector {
	// On crée un MetricsCollector avec un checker mocké via composition
	// en testant directement Collect() via un checker réel initialisé
	checker := &HealthChecker{}
	checker.cache = &PlatformHealth{
		Healthy:  healthy,
		Ready:    healthy,
		Backends: backends,
	}
	return &MetricsCollector{checker: checker}
}

func TestCollect_Format(t *testing.T) {
	backends := map[string]*HealthStatus{
		"argocd": {Name: "argocd", Healthy: true, Ready: true},
		"harbor": {Name: "harbor", Healthy: false, Ready: false},
	}
	mc := newMockCollector(true, backends)
	output := mc.Collect()

	// Vérifier la présence des métriques clés
	checks := []string{
		"dxp_platform_healthy",
		"dxp_platform_ready",
		"dxp_backend_healthy",
		"dxp_backend_ready",
		"dxp_backends_total",
		"dxp_backends_healthy_total",
		"# HELP",
		"# TYPE",
		"gauge",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("Métrique %q absente de la sortie", check)
		}
	}
}

func TestCollect_Values(t *testing.T) {
	backends := map[string]*HealthStatus{
		"argocd": {Name: "argocd", Healthy: true, Ready: true},
		"harbor": {Name: "harbor", Healthy: true, Ready: true},
	}
	mc := newMockCollector(true, backends)
	output := mc.Collect()

	if !strings.Contains(output, "dxp_platform_healthy 1") {
		t.Error("dxp_platform_healthy devrait être 1")
	}
	if !strings.Contains(output, "dxp_backends_total 2") {
		t.Error("dxp_backends_total devrait être 2")
	}
	if !strings.Contains(output, "dxp_backends_healthy_total 2") {
		t.Error("dxp_backends_healthy_total devrait être 2")
	}
}

func TestCollect_UnhealthyPlatform(t *testing.T) {
	backends := map[string]*HealthStatus{
		"argocd": {Name: "argocd", Healthy: false, Ready: false},
	}
	mc := newMockCollector(false, backends)
	output := mc.Collect()

	if !strings.Contains(output, "dxp_platform_healthy 0") {
		t.Error("dxp_platform_healthy devrait être 0")
	}
	if !strings.Contains(output, "dxp_backends_healthy_total 0") {
		t.Error("dxp_backends_healthy_total devrait être 0")
	}
}
