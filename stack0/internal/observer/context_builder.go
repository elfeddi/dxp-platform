package observer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/elfeddi/dxp/internal/gateway"
)

// PlatformContext est le snapshot JSON envoyé à C6 comme system prompt.
type PlatformContext struct {
	GeneratedAt  int64                    `json:"generated_at"`
	Platform     string                   `json:"platform"`
	Version      string                   `json:"version"`
	Backends     map[string]*BackendCtx   `json:"backends"`
	HealthSummary string                  `json:"health_summary"`
	ActiveStacks []string                 `json:"active_stacks"`
}

// BackendCtx est le contexte dun backend pour C6.
type BackendCtx struct {
	Name    string   `json:"name"`
	Healthy bool     `json:"healthy"`
	Version string   `json:"version,omitempty"`
	Message string   `json:"message,omitempty"`
	Actions []string `json:"available_actions,omitempty"`
}

// ContextBuilder construit le contexte plateforme pour C6.
type ContextBuilder struct {
	registry *gateway.Registry
	checker  *HealthChecker
	version  string
}

// NewContextBuilder crée un ContextBuilder.
func NewContextBuilder(reg *gateway.Registry, checker *HealthChecker, version string) *ContextBuilder {
	return &ContextBuilder{
		registry: reg,
		checker:  checker,
		version:  version,
	}
}

// Build construit le PlatformContext courant.
func (c *ContextBuilder) Build(ctx context.Context) (*PlatformContext, error) {
	health := c.checker.Get()
	backends := make(map[string]*BackendCtx, len(health.Backends))

	for name, hs := range health.Backends {
		bctx := &BackendCtx{
			Name:    name,
			Healthy: hs.Healthy,
			Message: hs.Message,
		}

		// Récupérer les actions disponibles (lecture seule — pas de secrets)
		if b, err := c.registry.Get(name); err == nil {
			if actions, err := b.ListActions(ctx); err == nil {
				for _, a := range actions {
					bctx.Actions = append(bctx.Actions, a.Name)
				}
			}
		}

		backends[name] = bctx
	}

	healthyCount := 0
	for _, b := range backends {
		if b.Healthy {
			healthyCount++
		}
	}

	summary := fmt.Sprintf("%d/%d backends healthy", healthyCount, len(backends))
	activeStacks := []string{"devops", "observability", "security"}

	return &PlatformContext{
		GeneratedAt:   time.Now().Unix(),
		Platform:      "DxP",
		Version:       c.version,
		Backends:      backends,
		HealthSummary: summary,
		ActiveStacks:  activeStacks,
	}, nil
}

// BuildSystemPrompt retourne le contexte formaté comme system prompt pour C6.
func (c *ContextBuilder) BuildSystemPrompt(ctx context.Context) (string, error) {
	pc, err := c.Build(ctx)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(pc, "", "  ")
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("Tu es l'assistant de la plateforme DxP. Voici l'état actuel de la plateforme :\n\n")
	sb.WriteString("```json\n")
	sb.Write(data)
	sb.WriteString("\n```\n\n")
	sb.WriteString("Règles absolues :\n")
	sb.WriteString("- Tu ne transmets jamais de secrets, tokens ou credentials.\n")
	sb.WriteString("- Toute action que tu proposes doit être validée par C1 Resolver.\n")
	sb.WriteString("- Tu ne prends aucune action automatique en production.\n")
	sb.WriteString("- Tu réponds en français sauf demande explicite.\n")

	return sb.String(), nil
}
