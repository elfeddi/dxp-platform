package aigateway

import (
	"context"
	"fmt"
	"strings"

	"github.com/elfeddi/dxp/internal/observer"
)

// PromptBuilder construit les prompts pour C6.
type PromptBuilder struct {
	ctxBuilder *observer.ContextBuilder
	filter     *Filter
}

// NewPromptBuilder crée un PromptBuilder.
func NewPromptBuilder(ctxBuilder *observer.ContextBuilder, filter *Filter) *PromptBuilder {
	return &PromptBuilder{ctxBuilder: ctxBuilder, filter: filter}
}

// BuildGeneratePrompt construit le prompt pour générer un dxp.yaml.
func (p *PromptBuilder) BuildGeneratePrompt(ctx context.Context, userPrompt string) (string, string, error) {
	// System prompt depuis C5 context_builder
	systemPrompt, err := p.ctxBuilder.BuildSystemPrompt(ctx)
	if err != nil {
		return "", "", fmt.Errorf("prompt_builder: failed to build system prompt: %w", err)
	}

	// Ajouter les instructions de génération YAML
	systemPrompt += p.yamlInstructions()

	// Filtrer le prompt utilisateur (au cas où il contiendrait des données sensibles)
	sanitizedUserPrompt := p.filter.SanitizeText(userPrompt)

	return systemPrompt, sanitizedUserPrompt, nil
}

// BuildDiagnosticPrompt construit le prompt pour diagnostiquer un incident.
func (p *PromptBuilder) BuildDiagnosticPrompt(ctx context.Context, incident string) (string, string, error) {
	systemPrompt, err := p.ctxBuilder.BuildSystemPrompt(ctx)
	if err != nil {
		return "", "", fmt.Errorf("prompt_builder: failed to build system prompt: %w", err)
	}

	systemPrompt += p.diagnosticInstructions()
	sanitizedIncident := p.filter.SanitizeText(incident)

	return systemPrompt, sanitizedIncident, nil
}

func (p *PromptBuilder) yamlInstructions() string {
	var sb strings.Builder
	sb.WriteString("\n## Instructions de génération\n\n")
	sb.WriteString("Quand l'utilisateur décrit ce qu'il veut déployer, génère UNIQUEMENT un fichier dxp.yaml valide.\n")
	sb.WriteString("Le format YAML doit respecter exactement cette structure :\n\n")
	sb.WriteString("```yaml\n")
	sb.WriteString("version: \"1.0\"\n")
	sb.WriteString("name: <nom-du-projet>\n")
	sb.WriteString("provider: <azure|aws|gcp|on-premise>\n")
	sb.WriteString("stacks:\n")
	sb.WriteString("  - name: <nom-stack>\n")
	sb.WriteString("    enabled: true\n")
	sb.WriteString("    connectors:\n")
	sb.WriteString("      - type: <argocd|harbor|grafana|tekton|vault>\n")
	sb.WriteString("        name: <nom-connecteur>\n")
	sb.WriteString("        config:\n")
	sb.WriteString("          url: <url>\n")
	sb.WriteString("```\n\n")
	sb.WriteString("Règles strictes :\n")
	sb.WriteString("- Ne génère QUE du YAML — aucun texte avant ou après.\n")
	sb.WriteString("- N'inclus jamais de credentials, tokens ou secrets dans le YAML.\n")
	sb.WriteString("- Utilise uniquement les types de connecteurs disponibles sur la plateforme.\n")
	sb.WriteString("- Si la demande est ambiguë, génère le YAML le plus probable et ajoute un commentaire.\n")
	sb.WriteString("- Utilise UNIQUEMENT ces URLs réelles pour les connecteurs :\n")
	sb.WriteString("  argocd: https://localhost:9443/argocd\n")
	sb.WriteString("  harbor: http://localhost:9091\n")
	sb.WriteString("  grafana: http://localhost:3001\n")
	return sb.String()
}

func (p *PromptBuilder) diagnosticInstructions() string {
	var sb strings.Builder
	sb.WriteString("\n## Instructions de diagnostic\n\n")
	sb.WriteString("Analyse le problème décrit et fournis :\n")
	sb.WriteString("1. Cause probable (1-2 phrases)\n")
	sb.WriteString("2. Actions recommandées (liste ordonnée)\n")
	sb.WriteString("3. Composants DxP potentiellement impactés\n\n")
	sb.WriteString("Sois concis et actionnable. Ne propose jamais d'actions automatiques en production.\n")
	return sb.String()
}
