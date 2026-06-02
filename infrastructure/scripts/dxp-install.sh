#!/bin/bash
source "$(dirname "$0")/lib/common.sh"

# ══════════════════════════════════════════════════════
# DxP Install — Orchestrateur principal (k3s natif)
# Usage : ./dxp-install.sh        ← installation complète
#         ./dxp-install.sh 06     ← rejoue uniquement 06_cicd-argocd.sh
# ══════════════════════════════════════════════════════

SCRIPTS_DIR="$(dirname "$0")"

run_step() {
  local num=$1
  local file=$(ls ${SCRIPTS_DIR}/${num}_*.sh 2>/dev/null | head -1)
  if [ -z "$file" ]; then
    warn "Script ${num} non trouvé — ignoré"
    return
  fi
  log "Exécution : $(basename $file)"
  bash "$file" || fail "Erreur dans $(basename $file) — arrêt"
}

title "DxP Platform — Installation complète (k3s natif)"
check_prereqs

if [ -n "$1" ]; then
  run_step "$1"
else
  run_step "01"   # Core — k3s vérification
  run_step "02"   # Core — repos Helm
  run_step "03"   # Core — cert-manager
  run_step "04"   # Network — ingress
  run_step "05"   # Registry — Harbor + CoreDNS
  run_step "06"   # CI/CD — ArgoCD
  run_step "07"   # CI/CD — Tekton
  run_step "08"   # Security — Vault
  run_step "09"   # Security — Kyverno
  run_step "10"   # Security — Falco (optionnel)
  run_step "11"   # Observability — Prometheus + Grafana
  run_step "12"   # Observability — Loki
  run_step "13"   # Observability — Tempo
  run_step "14"   # Observability — OTel
  run_step "15"   # Observability — Grafana datasources
  run_step "16"   # SSO — Dex
  run_step "17"   # DataOps — Airflow
  run_step "18"   # DataOps — dbt + Feast
  run_step "19"   # MLOps — MLflow
  run_step "20"   # AI — Ollama
  run_step "21"   # AI — LiteLLM
  run_step "22"   # AI — pgvector
fi

title "DxP Platform — Installation terminée"
echo ""
log "URLs d'accès :"
echo "  ArgoCD     : http://${DXP_IP}:9090"
echo "  Harbor     : http://${DXP_IP}:9091"
echo "  Grafana    : http://${DXP_IP}:3001"
echo "  Tekton     : http://${DXP_IP}:9295"
echo "  Airflow    : http://${DXP_IP}:9294"
echo "  Dex SSO    : http://${DXP_IP}:32000"
echo "  LiteLLM    : http://${DXP_IP}:30096"
