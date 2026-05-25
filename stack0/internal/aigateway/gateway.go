package aigateway

import (
	"context"
	"fmt"
	"strings"

	"github.com/elfeddi/dxp/internal/observer"
	"github.com/elfeddi/dxp/internal/resolver"
)

// GenerateResult est le résultat dune génération YAML.
type GenerateResult struct {
	YAML          string   `json:"yaml"`
	Valid         bool     `json:"valid"`
	Errors        []string `json:"errors,omitempty"`
	PromptTokens  int      `json:"prompt_tokens"`
	OutputTokens  int      `json:"output_tokens"`
}

// DiagnosticResult est le résultat dun diagnostic incident.
type DiagnosticResult struct {
	Analysis     string `json:"analysis"`
	PromptTokens int    `json:"prompt_tokens"`
	OutputTokens int    `json:"output_tokens"`
}

// Gateway est le point dentrée de C6 AIGateway.
// Orchestre : C5 context → filter → prompt → LiteLLM → C1 validate.
type Gateway struct {
	llm     *LiteLLMClient
	builder *PromptBuilder
	filter  *Filter
}

// NewGateway crée un Gateway C6.
func NewGateway(ctxBuilder *observer.ContextBuilder, baseURL, apiKey, model string) *Gateway {
	filter := NewFilter()
	return &Gateway{
		llm:     NewLiteLLMClient(baseURL, apiKey, model),
		builder: NewPromptBuilder(ctxBuilder, filter),
		filter:  filter,
	}
}

// Generate génère un dxp.yaml depuis un prompt naturel.
// Chaîne : C5 context → filter → LiteLLM → C1 validate.
func (g *Gateway) Generate(ctx context.Context, userPrompt string) (*GenerateResult, error) {
	// 1. Construire le prompt avec contexte C5
	systemPrompt, sanitizedPrompt, err := g.builder.BuildGeneratePrompt(ctx, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("C6: prompt build failed: %w", err)
	}

	// 2. Appeler LiteLLM
	resp, err := g.llm.Complete(ctx, systemPrompt, sanitizedPrompt)
	if err != nil {
		return nil, fmt.Errorf("C6: LLM call failed: %w", err)
	}

	// 3. Extraire et nettoyer le YAML
	rawContent := g.llm.ExtractContent(resp)
	yamlContent := extractYAML(rawContent)

	// 4. Valider via C1 Resolver (filet architectural obligatoire)
	result := &GenerateResult{
		YAML:         yamlContent,
		PromptTokens: resp.Usage.PromptTokens,
		OutputTokens: resp.Usage.CompletionTokens,
	}

	parser := resolver.NewParser()
	cfg, err := parser.ParseBytes([]byte(yamlContent))
	if err != nil {
		result.Valid = false
		result.Errors = []string{fmt.Sprintf("C1 parse error: %s", err.Error())}
		return result, nil
	}

	validator := resolver.NewValidator()
	errs := validator.Validate(cfg)
	if len(errs) > 0 {
		result.Valid = false
		for _, e := range errs {
			result.Errors = append(result.Errors, e.Error())
		}
		return result, nil
	}

	result.Valid = true
	return result, nil
}

// Diagnose diagnostique un incident en langage naturel.
func (g *Gateway) Diagnose(ctx context.Context, incident string) (*DiagnosticResult, error) {
	systemPrompt, sanitizedIncident, err := g.builder.BuildDiagnosticPrompt(ctx, incident)
	if err != nil {
		return nil, fmt.Errorf("C6: diagnostic prompt build failed: %w", err)
	}

	resp, err := g.llm.Complete(ctx, systemPrompt, sanitizedIncident)
	if err != nil {
		return nil, fmt.Errorf("C6: LLM call failed: %w", err)
	}

	return &DiagnosticResult{
		Analysis:     g.llm.ExtractContent(resp),
		PromptTokens: resp.Usage.PromptTokens,
		OutputTokens: resp.Usage.CompletionTokens,
	}, nil
}

// extractYAML extrait le bloc YAML depuis la réponse LLM.
// Le LLM peut encapsuler le YAML dans des backticks markdown.
func extractYAML(content string) string {
	content = strings.TrimSpace(content)

	// Cas 1 : bloc ```yaml ... ```
	if idx := strings.Index(content, "```yaml"); idx != -1 {
		start := idx + 7
		end := strings.Index(content[start:], "```")
		if end != -1 {
			return strings.TrimSpace(content[start : start+end])
		}
	}

	// Cas 2 : bloc ``` ... ```
	if idx := strings.Index(content, "```"); idx != -1 {
		start := idx + 3
		end := strings.Index(content[start:], "```")
		if end != -1 {
			return strings.TrimSpace(content[start : start+end])
		}
	}

	// Cas 3 : YAML direct (commence par "version:")
	if strings.HasPrefix(content, "version:") {
		return content
	}

	return content
}
