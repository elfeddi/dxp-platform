package aigateway

import (
	"strings"
)

// sensitivePrefixes liste les clés de config qui ne doivent jamais
// atteindre le LLM — tokens, passwords, secrets, credentials.
var sensitivePrefixes = []string{
	"token", "password", "passwd", "secret", "key",
	"credential", "auth", "apikey", "api_key",
	"private", "cert", "tls", "ssl",
}

// Filter filtre les données sensibles avant envoi au LLM.
type Filter struct{}

// NewFilter crée un Filter.
func NewFilter() *Filter { return &Filter{} }

// SanitizeConfig retourne une copie de la map sans les valeurs sensibles.
// Les clés sensibles sont remplacées par "[REDACTED]".
func (f *Filter) SanitizeConfig(config map[string]string) map[string]string {
	out := make(map[string]string, len(config))
	for k, v := range config {
		if f.isSensitive(k) {
			out[k] = "[REDACTED]"
		} else {
			out[k] = v
		}
	}
	return out
}

// SanitizeText remplace les patterns connus de secrets dans un texte libre.
func (f *Filter) SanitizeText(text string) string {
	// Ne jamais envoyer de tokens JWT (commencent par eyJ)
	lines := strings.Split(text, "\n")
	var out []string
	for _, line := range lines {
		if containsJWT(line) || containsBearerToken(line) {
			out = append(out, "[LINE REDACTED — sensitive data detected]")
		} else {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

func (f *Filter) isSensitive(key string) bool {
	lower := strings.ToLower(key)
	for _, prefix := range sensitivePrefixes {
		if strings.Contains(lower, prefix) {
			return true
		}
	}
	return false
}

func containsJWT(s string) bool {
	return strings.Contains(s, "eyJ") && strings.Count(s, ".") >= 2
}

func containsBearerToken(s string) bool {
	lower := strings.ToLower(s)
	return strings.Contains(lower, "bearer ") ||
		strings.Contains(lower, "authorization:") ||
		strings.Contains(lower, "api_key=") ||
		strings.Contains(lower, "token=")
}
