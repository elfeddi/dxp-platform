package observer

import (
	"fmt"
	"strings"
	"time"
)

// MetricsCollector produit des métriques Prometheus pour DxP.
type MetricsCollector struct {
	checker *HealthChecker
}

// NewMetricsCollector crée un MetricsCollector.
func NewMetricsCollector(checker *HealthChecker) *MetricsCollector {
	return &MetricsCollector{checker: checker}
}

// Collect retourne les métriques DxP au format texte Prometheus.
func (m *MetricsCollector) Collect() string {
	health := m.checker.Get()
	var b strings.Builder
	now := time.Now().UnixMilli()

	// dxp_platform_healthy
	b.WriteString("# HELP dxp_platform_healthy 1 si la plateforme DxP est healthy, 0 sinon.\n")
	b.WriteString("# TYPE dxp_platform_healthy gauge\n")
	if health.Healthy {
		b.WriteString(fmt.Sprintf("dxp_platform_healthy 1 %d\n", now))
	} else {
		b.WriteString(fmt.Sprintf("dxp_platform_healthy 0 %d\n", now))
	}

	// dxp_platform_ready
	b.WriteString("# HELP dxp_platform_ready 1 si la plateforme DxP est prête, 0 sinon.\n")
	b.WriteString("# TYPE dxp_platform_ready gauge\n")
	if health.Ready {
		b.WriteString(fmt.Sprintf("dxp_platform_ready 1 %d\n", now))
	} else {
		b.WriteString(fmt.Sprintf("dxp_platform_ready 0 %d\n", now))
	}

	// dxp_backend_healthy
	b.WriteString("# HELP dxp_backend_healthy 1 si le backend est healthy, 0 sinon.\n")
	b.WriteString("# TYPE dxp_backend_healthy gauge\n")
	for name, hs := range health.Backends {
		val := 0
		if hs.Healthy {
			val = 1
		}
		b.WriteString(fmt.Sprintf("dxp_backend_healthy{backend=%q} %d %d\n", name, val, now))
	}

	// dxp_backend_ready
	b.WriteString("# HELP dxp_backend_ready 1 si le backend est prêt, 0 sinon.\n")
	b.WriteString("# TYPE dxp_backend_ready gauge\n")
	for name, hs := range health.Backends {
		val := 0
		if hs.Ready {
			val = 1
		}
		b.WriteString(fmt.Sprintf("dxp_backend_ready{backend=%q} %d %d\n", name, val, now))
	}

	// dxp_backends_total
	b.WriteString("# HELP dxp_backends_total Nombre total de backends enregistrés.\n")
	b.WriteString("# TYPE dxp_backends_total gauge\n")
	b.WriteString(fmt.Sprintf("dxp_backends_total %d %d\n", len(health.Backends), now))

	// dxp_backends_healthy_total
	b.WriteString("# HELP dxp_backends_healthy_total Nombre de backends healthy.\n")
	b.WriteString("# TYPE dxp_backends_healthy_total gauge\n")
	healthyCount := 0
	for _, hs := range health.Backends {
		if hs.Healthy {
			healthyCount++
		}
	}
	b.WriteString(fmt.Sprintf("dxp_backends_healthy_total %d %d\n", healthyCount, now))

	return b.String()
}
