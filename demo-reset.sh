#!/bin/bash
# ══════════════════════════════════════════════════════
# DxP Demo Reset — Nettoyage avant démo
# Usage : bash demo-reset.sh [--keep-repos]
# --keep-repos : supprime K8s/ArgoCD mais garde les repos GitHub
# ══════════════════════════════════════════════════════

KEEP_REPOS=false
[[ "$1" == "--keep-repos" ]] && KEEP_REPOS=true

GITHUB_USER="elfeddi"
GITHUB_TOKEN="${GITHUB_TOKEN:-$(grep GITHUB_TOKEN ~/dxp-platform/infrastructure/scripts/.env | cut -d= -f2)}"
DXP_SERVE="http://localhost:30890"
BACKSTAGE="http://localhost:7007"

# Namespaces et apps à préserver (jamais supprimés)
SYSTEM_NS="default kube-system kube-public kube-node-lease argocd harbor tekton-pipelines tekton-pipelines-resolvers monitoring vault kyverno cert-manager dxp-system dataops mlops llmops ingress-nginx dex"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; RED='\033[0;31m'; NC='\033[0m'
ok()   { echo -e "${GREEN}[✓]${NC} $1"; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }
fail() { echo -e "${RED}[✗]${NC} $1"; }
step() { echo -e "\n${CYAN}══ $1 ══${NC}"; }
info() { echo -e "   $1"; }

# ── VÉRIFICATIONS ──────────────────────────────────
step "VÉRIFICATIONS PRÉALABLES"

kubectl get nodes &>/dev/null || { fail "Cluster K8s non disponible"; exit 1; }
ok "Cluster k3s opérationnel"

STATUS=$(curl -s -H "Authorization: Bearer viewer" $DXP_SERVE/api/dxp/status | \
  python3 -c "import sys,json; d=json.load(sys.stdin); print('ready' if d.get('ready') else 'not-ready')" 2>/dev/null)
[ "$STATUS" = "ready" ] && ok "dxp-serve ready" || warn "dxp-serve non disponible — continuer quand même"

# ── DÉTECTION NAMESPACES DE TEST ───────────────────
step "DÉTECTION NAMESPACES DE TEST"

TEST_NS=()
for ns in $(kubectl get namespaces --no-headers -o custom-columns=NAME:.metadata.name 2>/dev/null); do
  is_system=false
  for sys in $SYSTEM_NS; do
    [[ "$ns" == "$sys" ]] && is_system=true && break
  done
  if ! $is_system; then
    TEST_NS+=("$ns")
    info "Namespace de test détecté : $ns"
  fi
done

if [ ${#TEST_NS[@]} -eq 0 ]; then
  ok "Aucun namespace de test — environnement déjà propre"
else
  echo ""
  warn "${#TEST_NS[@]} namespace(s) à supprimer : ${TEST_NS[*]}"
fi

# ── DÉTECTION APPS ARGOCD DE TEST ─────────────────
step "DÉTECTION APPS ARGOCD DE TEST"

# Apps système à préserver
SYSTEM_APPS="dxp-serve harbor argocd tekton vault kyverno cert-manager prometheus grafana loki tempo otel airflow mlflow litellm pgvector"

TEST_APPS=()
for app in $(kubectl get applications -n argocd --no-headers -o custom-columns=NAME:.metadata.name 2>/dev/null); do
  is_system=false
  for sys in $SYSTEM_APPS; do
    [[ "$app" == "$sys" ]] && is_system=true && break
  done
  if ! $is_system; then
    TEST_APPS+=("$app")
    info "App ArgoCD de test détectée : $app"
  fi
done

[ ${#TEST_APPS[@]} -eq 0 ] && ok "Aucune app ArgoCD de test"

# ── REPOS GITHUB DE TEST ───────────────────────────
step "DÉTECTION REPOS GITHUB DE TEST"

# Repos système à préserver
SYSTEM_REPOS="dxp-platform dxp-demo task-api svc-demo-2"

TEST_REPOS=()
if [ "$KEEP_REPOS" = false ]; then
  REPOS_JSON=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
    "https://api.github.com/users/$GITHUB_USER/repos?per_page=100" 2>/dev/null)
  for repo in $(echo "$REPOS_JSON" | python3 -c "import sys,json; [print(r['name']) for r in json.load(sys.stdin)]" 2>/dev/null); do
    is_system=false
    for sys in $SYSTEM_REPOS; do
      [[ "$repo" == "$sys" ]] && is_system=true && break
    done
    # Seulement les repos créés via Golden Path DxP (svc-* ou dxp-*)
    is_dxp=false
    [[ "$repo" == svc-* || "$repo" == dxp-* ]] && is_dxp=true
    if ! $is_system && $is_dxp; then
      TEST_REPOS+=("$repo")
      info "Repo GitHub DxP de test détecté : $repo"
    fi
  done
  [ ${#TEST_REPOS[@]} -eq 0 ] && ok "Aucun repo GitHub de test"
else
  info "Mode --keep-repos : repos GitHub non touchés"
fi

# ── CONFIRMATION ───────────────────────────────────
echo ""
echo "════════════════════════════════════════════"
echo "  DxP Demo Reset — Résumé des suppressions"
echo "════════════════════════════════════════════"
echo "  Namespaces K8s  : ${TEST_NS[*]:-aucun}"
echo "  Apps ArgoCD     : ${TEST_APPS[*]:-aucune}"
echo "  Repos GitHub    : ${TEST_REPOS[*]:-aucun}"
echo "  keep-repos      : $KEEP_REPOS"
echo "════════════════════════════════════════════"
echo ""

if [ ${#TEST_NS[@]} -eq 0 ] && [ ${#TEST_APPS[@]} -eq 0 ] && [ ${#TEST_REPOS[@]} -eq 0 ]; then
  ok "Environnement déjà propre — rien à supprimer"
else
  read -p "Confirmer la suppression ? [y/N] " confirm
  [[ "$confirm" != "y" && "$confirm" != "Y" ]] && { warn "Annulé."; exit 0; }
fi

# ── SUPPRESSION APPS ARGOCD ────────────────────────
if [ ${#TEST_APPS[@]} -gt 0 ]; then
  step "SUPPRESSION APPS ARGOCD"
  for app in "${TEST_APPS[@]}"; do
    kubectl delete application "$app" -n argocd --ignore-not-found &>/dev/null && \
      ok "App ArgoCD supprimée : $app" || warn "Impossible de supprimer l'app : $app"
  done
fi

# ── SUPPRESSION NAMESPACES K8S ─────────────────────
if [ ${#TEST_NS[@]} -gt 0 ]; then
  step "SUPPRESSION NAMESPACES K8S"
  for ns in "${TEST_NS[@]}"; do
    kubectl delete namespace "$ns" --ignore-not-found &>/dev/null &
    ok "Namespace en cours de suppression : $ns"
  done
  # Attendre que tous les namespaces soient supprimés (max 60s)
  echo "Attente finalisation des suppressions..."
  for i in $(seq 1 20); do
    PENDING=$(kubectl get namespaces --no-headers 2>/dev/null | grep Terminating | wc -l)
    [ "$PENDING" -eq 0 ] && break
    sleep 3
  done
  ok "Namespaces supprimés"
fi

# ── SUPPRESSION REPOS GITHUB ───────────────────────
if [ ${#TEST_REPOS[@]} -gt 0 ]; then
  step "SUPPRESSION REPOS GITHUB"
  for repo in "${TEST_REPOS[@]}"; do
    HTTP=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE \
      -H "Authorization: token $GITHUB_TOKEN" \
      "https://api.github.com/repos/$GITHUB_USER/$repo")
    [ "$HTTP" = "204" ] && ok "Repo supprimé : $repo" || warn "Impossible de supprimer : $repo (HTTP $HTTP)"
  done
fi

# ── NETTOYAGE PIPELINERUNS ─────────────────────────
step "NETTOYAGE PIPELINERUNS TEKTON"

OLD_RUNS=$(kubectl get pipelineruns -n tekton-pipelines --no-headers 2>/dev/null | \
  grep -E "Succeeded|Failed" | awk '{print $1}')

if [ -n "$OLD_RUNS" ]; then
  echo "$OLD_RUNS" | while read run; do
    kubectl delete pipelinerun "$run" -n tekton-pipelines --ignore-not-found &>/dev/null
  done
  ok "Anciens PipelineRuns supprimés"
else
  ok "Aucun PipelineRun à nettoyer"
fi

# ── RÉENREGISTREMENT CATALOG BACKSTAGE ─────────────
step "RÉENREGISTREMENT CATALOG BACKSTAGE"

# Vérifier que Backstage est accessible (tunnel SSH requis)
HTTP=$(curl -s -o /dev/null -w "%{http_code}" "$BACKSTAGE/" 2>/dev/null)
if [ "$HTTP" = "200" ]; then
  # Réenregistrer les entités catalog via l'API Backstage
  LOCATIONS=(
    "https://github.com/elfeddi/dxp-platform/blob/main/templates/golden-path-devops/template.yaml"
    "https://github.com/elfeddi/dxp-platform/blob/main/stack0/catalog-info.yaml"
  )
  for loc in "${LOCATIONS[@]}"; do
    RESULT=$(curl -s -X POST "$BACKSTAGE/api/catalog/locations" \
      -H "Content-Type: application/json" \
      -d "{\"type\":\"url\",\"target\":\"$loc\"}" 2>/dev/null | \
      python3 -c "import sys,json; d=json.load(sys.stdin); print('ok')" 2>/dev/null)
    [ "$RESULT" = "ok" ] && ok "Catalog enregistré : $(basename $loc)" || \
      warn "Catalog déjà présent ou erreur : $(basename $loc)"
  done
else
  warn "Backstage non accessible (HTTP $HTTP) — SSH tunnel requis pour le catalog"
  info "Lancer : ssh -L 7007:127.0.0.1:7007 azureuser@158.158.8.131 -N"
fi

# ── VÉRIFICATION FINALE ────────────────────────────
step "VÉRIFICATION FINALE"

# dxp-serve
STATUS=$(curl -s -H "Authorization: Bearer viewer" $DXP_SERVE/api/dxp/status | \
  python3 -c "import sys,json; d=json.load(sys.stdin); print('ready' if d.get('ready') else 'not-ready')" 2>/dev/null)
[ "$STATUS" = "ready" ] && ok "dxp-serve ready" || warn "dxp-serve non ready — vérifier dxp-system"

# Pods non Running
NOT_RUNNING=$(kubectl get pods -A --no-headers 2>/dev/null | \
  grep -v "Running\|Completed\|Evicted\|Terminating" | wc -l)
[ "$NOT_RUNNING" -eq 0 ] && ok "Tous les pods système en Running" || \
  warn "$NOT_RUNNING pod(s) non Running — vérifier avec : kubectl get pods -A | grep -v Running"

# Namespaces restants
REMAINING=$(kubectl get namespaces --no-headers 2>/dev/null | \
  grep -vE "^(default|kube-system|kube-public|kube-node-lease|argocd|harbor|tekton-pipelines|monitoring|vault|kyverno|cert-manager|dxp-system|dataops|mlops|llmops|ingress-nginx)" | wc -l)
[ "$REMAINING" -eq 0 ] && ok "Aucun namespace de test résiduel" || \
  warn "$REMAINING namespace(s) résiduel(s)"

# ── RÉSUMÉ ─────────────────────────────────────────
echo ""
echo "════════════════════════════════════════════"
echo -e "${GREEN}  DxP — Environnement prêt pour la démo${NC}"
echo "════════════════════════════════════════════"
echo "  Backstage   : http://localhost:7007 (SSH tunnel)"
echo "  ArgoCD      : http://158.158.8.131:9090"
echo "  Harbor      : http://158.158.8.131:9091"
echo "  Grafana     : http://158.158.8.131:3001"
echo "  Tekton      : http://158.158.8.131:9295"
echo ""
echo "  Scénario démo :"
echo "  1. Backstage → Create → DxP Golden Path"
echo "  2. git push → Tekton (35s) → pod Running"
echo "  3. Backstage Catalog → onglet Kubernetes"
echo "════════════════════════════════════════════"
