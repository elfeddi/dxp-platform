#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "20 — Ollama (désactivé — API externe utilisée en POC)"

# ADR S0-020 : API LLM externe utilisée pour la démo (OpenAI/Anthropic/Gemini)
# Ollama non installé en POC — décommenter pour activation locale

warn "Ollama désactivé (ADR S0-020) — LiteLLM pointe sur API externe"
warn "Pour activer : décommenter le bloc ci-dessous"

# curl -fsSL https://ollama.com/install.sh | sh
# sudo mkdir -p /etc/systemd/system/ollama.service.d
# sudo tee /etc/systemd/system/ollama.service.d/override.conf > /dev/null << OLLAMAEOF
# [Service]
# Environment="OLLAMA_HOST=0.0.0.0"
# OLLAMAEOF
# sudo systemctl daemon-reload
# sudo systemctl enable ollama
# sudo systemctl restart ollama
# ollama pull ${OLLAMA_MODEL:-mistral}

ok "Ollama — skippé (POC)"
