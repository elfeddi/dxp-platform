#!/bin/bash
source "$(dirname "$0")/lib/common.sh"

# ══════════════════════════════════════════════════════
# DxP Install — Orchestrateur principal
# Usage : ./dxp-install.sh [script_number]
# Exemple : ./dxp-install.sh 03  ← rejoue uniquement 03-argocd.sh
# Sans argument : installe tout depuis le début
# ══════════════════════════════════════════════════════

SCRIPTS_DIR="$(dirname "$0")"

run_step() {
  local num=$1
  local script="${SCRIPTS_DIR}/${num}-*.sh"
  local file=$(ls $script 2>/dev/null | head -1)
  if [ -z "$file" ]; then
    warn "Script $num non trouvé — ignoré"
    return
  fi
  log "Exécution : $(basename $file)"
  bash "$file" || fail "Erreur dans $(basename $file) — arrêt"
}

title "DxP Platform — Installation complète"
check_prereqs

if [ -n "$1" ]; then
  # Rejouer un script spécifique
  run_step "$1"
else
  # Installation complète
  run_step "01"   # Cluster k3d
  run_step "02"   # Repos Helm
  run_step "03"   # ArgoCD
  run_step "04"   # Harbor + CoreDNS
  run_step "05"   # Tekton
  run_step "06"   # LGTM
  run_step "07"   # Security
  run_step "08"   # DataOps
  run_step "09"   # MLOps
  run_step "10"   # LLMOps
  run_step "11"   # Dex SSO
fi

title "DxP Platform — Installation terminée"
echo ""
log "URLs d'accès :"
echo "  ArgoCD     : http://${DXP_IP}:9090"
echo "  Harbor     : http://${DXP_IP}:9091"
echo "  Grafana    : http://${DXP_IP}:3001"
echo "  Tekton     : http://${DXP_IP}:9295"
echo "  Dex SSO    : http://${DXP_IP}:32000"
echo "  Airflow    : http://${DXP_IP}:9294"
