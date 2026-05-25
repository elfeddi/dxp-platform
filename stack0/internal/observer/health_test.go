package observer

import (
	"testing"
	"time"
)

func TestHealthStatus_Fields(t *testing.T) {
	hs := &HealthStatus{
		Name:      "argocd",
		Healthy:   true,
		Ready:     true,
		Message:   "ArgoCD operational",
		CheckedAt: time.Now().Unix(),
	}
	if hs.Name != "argocd" {
		t.Errorf("Name attendu %q, obtenu %q", "argocd", hs.Name)
	}
	if !hs.Healthy {
		t.Error("Healthy devrait être true")
	}
}

func TestPlatformHealth_AllHealthy(t *testing.T) {
	ph := &PlatformHealth{
		Healthy: true,
		Ready:   true,
		Backends: map[string]*HealthStatus{
			"argocd": {Name: "argocd", Healthy: true, Ready: true},
			"harbor": {Name: "harbor", Healthy: true, Ready: true},
		},
		CheckedAt: time.Now().Unix(),
	}
	if !ph.Healthy {
		t.Error("PlatformHealth devrait être healthy")
	}
	if len(ph.Backends) != 2 {
		t.Errorf("2 backends attendus, obtenus %d", len(ph.Backends))
	}
}

func TestPlatformHealth_OneUnhealthy(t *testing.T) {
	ph := &PlatformHealth{
		Healthy: false,
		Ready:   false,
		Backends: map[string]*HealthStatus{
			"argocd": {Name: "argocd", Healthy: true, Ready: true},
			"harbor": {Name: "harbor", Healthy: false, Ready: false, Message: "connection refused"},
		},
		CheckedAt: time.Now().Unix(),
	}
	if ph.Healthy {
		t.Error("PlatformHealth devrait être unhealthy")
	}
	if ph.Backends["harbor"].Message == "" {
		t.Error("Message d'erreur attendu pour harbor")
	}
}
