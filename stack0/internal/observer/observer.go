package observer

import (
	"context"
	"time"

	"github.com/elfeddi/dxp/internal/gateway"
)

const defaultVersion = "v0.2.0"

// Observer est le point dentrée de C5.
// Il agrège HealthChecker, StatusAggregator, MetricsCollector et ContextBuilder.
type Observer struct {
	Health   *HealthChecker
	Status   *StatusAggregator
	Metrics  *MetricsCollector
	Context  *ContextBuilder
	Tracer   *OTelTracer
}

// New crée un Observer complet depuis le registry C4.
func New(reg *gateway.Registry) *Observer {
	checker := NewHealthChecker(reg, 30*time.Second)
	return &Observer{
		Health:  checker,
		Status:  NewStatusAggregator(checker, defaultVersion),
		Metrics: NewMetricsCollector(checker),
		Context: NewContextBuilder(reg, checker, defaultVersion),
		Tracer:  NewOTelTracer("dxp-c4-gateway"),
	}
}

// Start lance les processus background de C5 (health checks périodiques).
func (o *Observer) Start(ctx context.Context) {
	o.Health.Start(ctx)
}
