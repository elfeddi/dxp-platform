#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "20 — Ollama (systemd)"

# Installation
curl -fsSL https://ollama.com/install.sh | sh

# Configurer écoute sur 0.0.0.0 (pas uniquement localhost)
sudo mkdir -p /etc/systemd/system/ollama.service.d
sudo tee /etc/systemd/system/ollama.service.d/override.conf > /dev/null << OLLAMAEOF
[Service]
Environment="OLLAMA_HOST=0.0.0.0"
OLLAMAEOF

sudo systemctl daemon-reload
sudo systemctl enable ollama
sudo systemctl restart ollama
sleep 5

# Vérification
curl -s http://localhost:11434 | grep -q "Ollama" && ok "Ollama installé — localhost:11434" || warn "Ollama démarrage en cours..."

# Pull modèle par défaut
log "Pull modèle ${OLLAMA_MODEL:-mistral}..."
ollama pull ${OLLAMA_MODEL:-mistral}
ok "Modèle ${OLLAMA_MODEL:-mistral} disponible"
