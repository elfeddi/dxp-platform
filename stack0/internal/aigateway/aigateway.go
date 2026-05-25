package aigateway

import (
	"os"

	"github.com/elfeddi/dxp/internal/observer"
)

const (
	defaultBaseURL = "http://localhost:30096"
	defaultModel   = "dxp-default"
)

// New crée un Gateway C6 depuis les variables d'environnement.
// Variables lues :
//   LITELLM_API_BASE  — URL de LiteLLM (défaut: http://localhost:30096)
//   LITELLM_API_KEY   — clé API LiteLLM
//   LITELLM_MODEL     — modèle à utiliser (défaut: dxp-default)
func New(ctxBuilder *observer.ContextBuilder) *Gateway {
	baseURL := getEnv("LITELLM_API_BASE", defaultBaseURL)
	apiKey := getEnv("LITELLM_API_KEY", "")
	model := getEnv("LITELLM_MODEL", defaultModel)

	return NewGateway(ctxBuilder, baseURL, apiKey, model)
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
