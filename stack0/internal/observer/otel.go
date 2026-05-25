package observer

import (
	"context"
	"log/slog"
	"time"
)

// Span représente une trace OpenTelemetry simplifiée pour le POC.
// En production : remplacer par go.opentelemetry.io/otel/trace.
type Span struct {
	Name      string
	StartTime time.Time
	Attrs     map[string]string
}

// OTelTracer est un traceur OpenTelemetry stub pour le POC.
// Phase 2 : remplacer par le SDK OTel complet avec export vers Tempo.
type OTelTracer struct {
	serviceName string
}

// NewOTelTracer crée un OTelTracer.
func NewOTelTracer(serviceName string) *OTelTracer {
	return &OTelTracer{serviceName: serviceName}
}

// Start démarre une span et retourne un contexte enrichi + une fonction end().
func (t *OTelTracer) Start(ctx context.Context, spanName string, attrs map[string]string) (context.Context, func()) {
	span := &Span{
		Name:      spanName,
		StartTime: time.Now(),
		Attrs:     attrs,
	}

	end := func() {
		duration := time.Since(span.StartTime)
		slog.Debug("otel span",
			"service", t.serviceName,
			"span", span.Name,
			"duration_ms", duration.Milliseconds(),
			"attrs", span.Attrs,
		)
	}

	return ctx, end
}
