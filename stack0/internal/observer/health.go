package observer

import (
	"context"
	"sync"
	"time"

	"github.com/elfeddi/dxp/internal/gateway"
)

// HealthStatus représente létat de santé dun composant.
type HealthStatus struct {
	Name      string `json:"name"`
	Healthy   bool   `json:"healthy"`
	Ready     bool   `json:"ready"`
	Message   string `json:"message,omitempty"`
	CheckedAt int64  `json:"checked_at"`
}

// PlatformHealth représente létat de santé global de la plateforme DxP.
type PlatformHealth struct {
	Healthy   bool                     `json:"healthy"`
	Ready     bool                     `json:"ready"`
	Backends  map[string]*HealthStatus `json:"backends"`
	CheckedAt int64                    `json:"checked_at"`
}

// HealthChecker interroge périodiquement tous les backends C4.
type HealthChecker struct {
	registry *gateway.Registry
	mu       sync.RWMutex
	cache    *PlatformHealth
	interval time.Duration
}

// NewHealthChecker crée un HealthChecker avec un intervalle de refresh.
func NewHealthChecker(reg *gateway.Registry, interval time.Duration) *HealthChecker {
	return &HealthChecker{
		registry: reg,
		interval: interval,
	}
}

// Start lance la boucle de health check en arrière-plan.
func (h *HealthChecker) Start(ctx context.Context) {
	h.check(ctx)
	go func() {
		ticker := time.NewTicker(h.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				h.check(ctx)
			}
		}
	}()
}

// Get retourne le dernier état de santé connu (depuis le cache).
func (h *HealthChecker) Get() *PlatformHealth {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.cache == nil {
		return &PlatformHealth{
			Healthy:   false,
			Ready:     false,
			Backends:  map[string]*HealthStatus{},
			CheckedAt: time.Now().Unix(),
		}
	}
	return h.cache
}

// check interroge tous les backends et met à jour le cache.
func (h *HealthChecker) check(ctx context.Context) {
	backends := h.registry.All()
	statuses := make(map[string]*HealthStatus, len(backends))
	allHealthy := true
	allReady := true

	for name, b := range backends {
		ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
		status, err := b.GetStatus(ctx2)
		cancel()

		hs := &HealthStatus{
			Name:      name,
			CheckedAt: time.Now().Unix(),
		}
		if err != nil {
			hs.Healthy = false
			hs.Ready = false
			hs.Message = err.Error()
		} else {
			hs.Healthy = status.Healthy
			hs.Ready = status.Ready
			hs.Message = status.Message
		}

		if !hs.Healthy {
			allHealthy = false
		}
		if !hs.Ready {
			allReady = false
		}
		statuses[name] = hs
	}

	h.mu.Lock()
	h.cache = &PlatformHealth{
		Healthy:   allHealthy,
		Ready:     allReady,
		Backends:  statuses,
		CheckedAt: time.Now().Unix(),
	}
	h.mu.Unlock()
}
