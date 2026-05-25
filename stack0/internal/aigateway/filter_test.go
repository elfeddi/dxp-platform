package aigateway

import (
	"testing"
)

func TestSanitizeConfig_Sensitive(t *testing.T) {
	f := NewFilter()
	input := map[string]string{
		"url":      "http://localhost:9090",
		"token":    "eyJhbGciOiJIUzI1NiJ9.secret",
		"password": "dxp-secret-2026",
		"username": "admin",
		"api_key":  "sk-1234",
	}
	result := f.SanitizeConfig(input)

	if result["url"] != "http://localhost:9090" {
		t.Errorf("url ne devrait pas être redacted, obtenu %q", result["url"])
	}
	if result["token"] != "[REDACTED]" {
		t.Errorf("token devrait être redacted, obtenu %q", result["token"])
	}
	if result["password"] != "[REDACTED]" {
		t.Errorf("password devrait être redacted, obtenu %q", result["password"])
	}
	if result["username"] != "admin" {
		t.Errorf("username ne devrait pas être redacted, obtenu %q", result["username"])
	}
	if result["api_key"] != "[REDACTED]" {
		t.Errorf("api_key devrait être redacted, obtenu %q", result["api_key"])
	}
}

func TestSanitizeConfig_Empty(t *testing.T) {
	f := NewFilter()
	result := f.SanitizeConfig(map[string]string{})
	if len(result) != 0 {
		t.Errorf("Map vide attendue, obtenu %d entrées", len(result))
	}
}

func TestSanitizeText_JWT(t *testing.T) {
	f := NewFilter()
	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0In0.SflKxwRJSMeKKF2QT4fwpMeJf36"
	result := f.SanitizeText("prompt avec token " + jwt)
	if result == "prompt avec token "+jwt {
		t.Error("JWT devrait être redacted")
	}
	if result != "[LINE REDACTED — sensitive data detected]" {
		t.Errorf("Ligne JWT attendue redacted, obtenu %q", result)
	}
}

func TestSanitizeText_BearerToken(t *testing.T) {
	f := NewFilter()
	input := "Authorization: Bearer my-secret-token"
	result := f.SanitizeText(input)
	if result == input {
		t.Error("Bearer token devrait être redacted")
	}
}

func TestSanitizeText_SafeText(t *testing.T) {
	f := NewFilter()
	input := "deploie une stack devops avec argocd et harbor sur azure"
	result := f.SanitizeText(input)
	if result != input {
		t.Errorf("Texte safe ne devrait pas être modifié, obtenu %q", result)
	}
}

func TestSanitizeText_MultiLine(t *testing.T) {
	f := NewFilter()
	input := "ligne normale\nAuthorization: Bearer secret\nautre ligne normale"
	result := f.SanitizeText(input)
	lines := splitLines(result)
	if len(lines) != 3 {
		t.Errorf("3 lignes attendues, obtenu %d", len(lines))
	}
	if lines[0] != "ligne normale" {
		t.Errorf("Ligne 1 ne devrait pas être redacted: %q", lines[0])
	}
	if lines[1] != "[LINE REDACTED — sensitive data detected]" {
		t.Errorf("Ligne 2 devrait être redacted: %q", lines[1])
	}
	if lines[2] != "autre ligne normale" {
		t.Errorf("Ligne 3 ne devrait pas être redacted: %q", lines[2])
	}
}

func TestExtractYAML_WithBackticks(t *testing.T) {
	input := "Voici le YAML:\n```yaml\nversion: \"1.0\"\nname: test\n```\nC'est tout."
	result := extractYAML(input)
	if result != "version: \"1.0\"\nname: test" {
		t.Errorf("extractYAML incorrect, obtenu %q", result)
	}
}

func TestExtractYAML_DirectYAML(t *testing.T) {
	input := "version: \"1.0\"\nname: test\nprovider: azure"
	result := extractYAML(input)
	if result != input {
		t.Errorf("extractYAML direct incorrect, obtenu %q", result)
	}
}

func TestExtractYAML_WithoutBackticks(t *testing.T) {
	input := "```\nversion: \"1.0\"\nname: test\n```"
	result := extractYAML(input)
	if result != "version: \"1.0\"\nname: test" {
		t.Errorf("extractYAML sans yaml tag incorrect, obtenu %q", result)
	}
}

// splitLines helper
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	lines = append(lines, s[start:])
	return lines
}
