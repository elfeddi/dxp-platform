#!/bin/bash

# ══ Couleurs ══
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
BLUE='\033[0;34m'; CYAN='\033[0;36m'; NC='\033[0m'

# ══ Fonctions de log ══
log()   { echo -e "${BLUE}[DxP]${NC} $1"; }
ok()    { echo -e "${GREEN}[✓]${NC} $1"; }
warn()  { echo -e "${YELLOW}[!]${NC} $1"; }
fail()  { echo -e "${RED}[✗]${NC} $1"; exit 1; }
title() { echo -e "\n${CYAN}══════════════════════════════════${NC}"; \
          echo -e "${CYAN}  $1${NC}"; \
          echo -e "${CYAN}══════════════════════════════════${NC}\n"; }

# ══ Attente pods ══
wait_pods() {
  local ns=$1
  local timeout=${2:-180}
  log "Attente pods namespace '$ns'..."
  kubectl wait --for=condition=ready pod --all -n $ns --timeout=${timeout}s 2>/dev/null \
    && ok "Pods '$ns' prêts" \
    || warn "Timeout sur '$ns' — vérifier manuellement"
}

# ══ Vérification prérequis ══
check_prereqs() {
  title "Vérification prérequis"
  for cmd in k3d helm kubectl docker curl; do
    command -v $cmd >/dev/null 2>&1 \
      && ok "$cmd trouvé" \
      || fail "$cmd non installé — arrêt"
  done
}

# ══ Chargement .env ══
load_env() {
  local env_file="$(dirname "$0")/../.env"
  if [ -f "$env_file" ]; then
    source "$env_file"
    ok ".env chargé"
  else
    fail ".env non trouvé — copier .env.example et remplir les valeurs"
  fi
}

# ══ Variables globales namespaces ══
export DXP_NAMESPACE_ARGOCD="argocd"
export DXP_NAMESPACE_HARBOR="harbor"
export DXP_NAMESPACE_MONITORING="monitoring"
export DXP_NAMESPACE_KYVERNO="kyverno"
export DXP_NAMESPACE_FALCO="falco"
export DXP_NAMESPACE_CERT_MANAGER="cert-manager"
export DXP_NAMESPACE_VAULT="vault"
export DXP_NAMESPACE_DATAOPS="dataops"
export DXP_NAMESPACE_MLOPS="mlops"
export DXP_NAMESPACE_LLMOPS="llmops"
export DXP_NAMESPACE_DEX="dex"
export DXP_NAMESPACE_TEKTON="tekton-pipelines"

export DXP_IP=$(curl -s ifconfig.me 2>/dev/null)

