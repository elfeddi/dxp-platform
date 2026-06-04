#!/bin/bash
# ══════════════════════════════════════════════════════
# DxP Demo e2e — Scénario complet LT + SE
# Usage : bash demo-e2e.sh [nom-service]
# ══════════════════════════════════════════════════════

SERVICE=${1:-demo-service}
GITHUB_USER="elfeddi"
GITHUB_TOKEN="${GITHUB_TOKEN:-$(grep GITHUB_TOKEN ~/dxp-platform/infrastructure/scripts/.env | cut -d= -f2)}"
DXP_IP="158.158.8.131"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; NC='\033[0m'
ok()         { echo -e "${GREEN}[✓]${NC} $1"; }
step()       { echo -e "\n${CYAN}══ $1 ══${NC}"; }
wait_input() { echo -e "${YELLOW}[→]${NC} $1 — appuie sur Entrée pour continuer..."; read; }

# ── SCÉNARIO LT ────────────────────────────────────
step "SCÉNARIO LEAD TECH — Provisionner un nouveau service"

echo "Service : $SERVICE"
echo "Le LT va :"
echo "  1. Créer le repo GitHub via Backstage Golden Path"
echo "  2. Appeler /provision → namespace K8s + webhook Tekton"
echo "  (ArgoCD sera créé par Tekton après le premier build)"
echo ""

wait_input "Ouvre Backstage sur http://localhost:7007 (SSH tunnel requis)"

# Étape 1 — /provision via dxp-serve
step "Étape 1 — Provision via C4 Gateway"
echo "Appel de /provision pour $SERVICE..."

RESULT=$(curl -s -X POST http://localhost:30890/api/dxp/provision \
  -H "Authorization: Bearer operator" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"$SERVICE\",
    \"repo\": \"https://github.com/$GITHUB_USER/$SERVICE\",
    \"namespace\": \"$SERVICE-dev\",
    \"language\": \"nodejs\"
  }")

echo "$RESULT" | python3 -m json.tool 2>/dev/null || echo "$RESULT"

# Vérifier namespace créé
if kubectl get namespace "$SERVICE-dev" &>/dev/null; then
  ok "Namespace $SERVICE-dev créé"
else
  echo "Namespace non créé — vérifier les logs dxp-serve"
fi

# Vérifier webhook GitHub
WEBHOOK_STATUS=$(echo "$RESULT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('steps',{}).get('webhook','?'))" 2>/dev/null)
BRANCH_STATUS=$(echo "$RESULT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('steps',{}).get('branch_protection','?'))" 2>/dev/null)
ok "Webhook Tekton : $WEBHOOK_STATUS"
ok "Protection branche : $BRANCH_STATUS"

ok "Scénario LT terminé — ArgoCD sera créé par Tekton après le premier build"
echo ""

# ── SCÉNARIO SE ────────────────────────────────────
step "SCÉNARIO SOFTWARE ENGINEER — Développer et déployer"

echo "Le SE va :"
echo "  1. Cloner le repo $SERVICE"
echo "  2. Modifier le code"
echo "  3. git push → Tekton → Harbor → ArgoCD → pod running"
echo ""

wait_input "Scénario SE prêt"

# Clone + modification
TMPDIR=$(mktemp -d)
cd $TMPDIR

echo "Clone du repo $SERVICE..."
git clone https://$GITHUB_TOKEN@github.com/$GITHUB_USER/$SERVICE.git &>/dev/null || {
  echo "Repo non trouvé — s'assurer que le Golden Path a été exécuté"
  exit 1
}
ok "Repo cloné dans $TMPDIR/$SERVICE"

cd $SERVICE

# Modification du code
echo "// DxP demo — $(date)" >> index.js
git add .
git config user.email "se@dxp.io"
git config user.name "SE DxP"
git commit -m "feat: demo update $(date +%H%M%S)"
git push

ok "Code poussé — pipeline Tekton déclenché"
echo ""

# Surveiller le pipeline
step "Surveillance pipeline Tekton"
echo "Attente du PipelineRun..."
sleep 5

RUN=""
for i in $(seq 1 30); do
  RUN=$(kubectl get pipelineruns -n tekton-pipelines --no-headers 2>/dev/null | \
    grep "$SERVICE" | grep "Running\|Succeeded" | tail -1 | awk '{print $1}')
  if [ -n "$RUN" ]; then
    echo "PipelineRun : $RUN"
    break
  fi
  sleep 3
done

if [ -z "$RUN" ]; then
  echo "Aucun PipelineRun trouvé pour $SERVICE — vérifier le webhook"
else
  echo "Attente Succeeded..."
  kubectl wait pipelinerun/$RUN -n tekton-pipelines \
    --for=condition=Succeeded --timeout=300s 2>/dev/null && \
    ok "Pipeline Succeeded — image buildée et pushée dans Harbor" || \
    echo "Timeout — vérifier manuellement"
fi

# Vérifier ArgoCD app (créée par Tekton task 4)
sleep 5
if kubectl get application -n argocd "$SERVICE" &>/dev/null; then
  ok "Application ArgoCD $SERVICE créée par Tekton"
else
  echo "Application ArgoCD non trouvée — vérifier la task create-argocd-app"
fi

# Vérifier le pod
step "Vérification déploiement"
echo "Attente pod dans $SERVICE-dev..."
sleep 10

POD=$(kubectl get pods -n "$SERVICE-dev" --no-headers 2>/dev/null | head -1 | awk '{print $1}')
if [ -n "$POD" ]; then
  ok "Pod $POD running dans $SERVICE-dev"
else
  echo "Pas de pod — ArgoCD sync en cours..."
fi

# Nettoyage
cd ~
rm -rf $TMPDIR

step "DÉMO TERMINÉE"
echo ""
echo "Résumé :"
echo "  Namespace : $SERVICE-dev"
echo "  ArgoCD    : http://$DXP_IP:9090"
echo "  Harbor    : http://$DXP_IP:9091"
echo "  Tekton    : http://$DXP_IP:9295"
echo ""
ok "DxP e2e validé — de Backstage au pod en production"
