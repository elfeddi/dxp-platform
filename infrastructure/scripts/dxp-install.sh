#!/bin/bash
source "$(dirname "$0")/lib/common.sh"

# ══════════════════════════════════════════════════════
# DxP Install — Orchestrateur principal (k3s natif)
# Usage : ./dxp-install.sh        ← installation complète
#         ./dxp-install.sh 06     ← rejoue uniquement 06_cicd-argocd.sh
#
# Architecture POC (ADR S0-020) :
#   Master  B4ms  role=infra          → plateforme + dataops + AI
#   Worker1 B2s   role=observability  → LGTM + OTel
#   Worker2 B2s   role=workload       → apps clients (taint NoSchedule)
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

title "DxP Platform — Installation complète (k3s natif · POC production-like)"
check_prereqs

if [ -n "$1" ]; then
  run_step "$1"
else
  # Core
  run_step "01"   # k3s vérification + labels
  run_step "02"   # Repos Helm
  run_step "03"   # cert-manager + ClusterIssuer

  # Network
  run_step "04"   # Ingress nginx

  # Registry
  run_step "05"   # Harbor + CoreDNS + registre k3s

  # CI/CD
  run_step "06"   # ArgoCD (2 replicas)
  run_step "07"   # Tekton

  # Security
  run_step "08"   # Vault (mode dev)
  run_step "09"   # Kyverno (2 replicas)
  run_step "10"   # Falco (optionnel)

  # Observability
  run_step "11"   # Prometheus + Grafana (Worker1)
  run_step "12"   # Loki (Worker1)
  run_step "13"   # Tempo (Worker1)
  run_step "14"   # OTel Collector (DaemonSet)
  run_step "15"   # Grafana datasources

  # SSO
  run_step "16"   # Dex SSO

  # DataOps
  run_step "17"   # Airflow (Master)
  run_step "18"   # dbt + Feast + Great Expectations

  # MLOps
  run_step "19"   # MLflow (Master)

  # AI/LLMOps
  run_step "20"   # Ollama (désactivé POC)
  run_step "21"   # LiteLLM (Master)
  run_step "22"   # pgvector (Master)
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
