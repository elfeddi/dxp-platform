#!/bin/bash
# DxP — Script de démarrage
# Session 10 — 25 mai 2026
# Usage : ./dxp-start.sh
set -euo pipefail

GREEN='\033[0;32m'; CYAN='\033[0;36m'; YELLOW='\033[1;33m'; RESET='\033[0m'
ok()   { echo -e "${GREEN}  ✓${RESET} $1"; }
info() { echo -e "${CYAN}  →${RESET} $1"; }
warn() { echo -e "${YELLOW}  !${RESET} $1"; }
header() { echo -e "\n${CYAN}══ $1 ══${RESET}"; }

header "DxP — Démarrage"

# ── 1. Cluster ────────────────────────────────────────────────────
header "1. Cluster k3d"
k3d cluster start dxp-poc
kubectl wait --for=condition=Ready nodes --all --timeout=90s 2>/dev/null || true
ok "Cluster démarré — $(kubectl get nodes --no-headers | wc -l) nœuds Ready"

# ── 2. CoreDNS — rewrite harbor.dxp ──────────────────────────────
header "2. CoreDNS"
kubectl patch configmap coredns -n kube-system --type merge -p \
  '{"data":{"Corefile":".:53 {\n    errors\n    health\n    ready\n    kubernetes cluster.local in-addr.arpa ip6.arpa {\n      pods insecure\n      fallthrough in-addr.arpa ip6.arpa\n    }\n    rewrite name harbor.dxp harbor.harbor.svc.cluster.local\n    prometheus :9153\n    forward . /etc/resolv.conf\n    cache 30\n    loop\n    reload\n    loadbalance\n  }"}}'
kubectl rollout restart deployment coredns -n kube-system > /dev/null 2>&1
ok "CoreDNS patché — harbor.dxp résolu"

# ── 3. Registries k3d (Harbor IP) ────────────────────────────────
header "3. Harbor registries"
HARBOR_IP=$(kubectl get svc harbor -n harbor -o jsonpath='{.spec.clusterIP}')
info "Harbor ClusterIP : $HARBOR_IP"

for node in $(k3d node list | grep dxp-poc | grep -v tools | grep -v serverlb | awk '{print $1}'); do
  docker exec "${node}" sh -c "mkdir -p /etc/rancher/k3s && cat > /etc/rancher/k3s/registries.yaml << YAML
mirrors:
  harbor.dxp:
    endpoint:
      - http://${HARBOR_IP}
configs:
  harbor.dxp:
    tls:
      insecure_skip_verify: true
YAML" 2>/dev/null
done
ok "registries.yaml mis à jour sur tous les nœuds (dont agent-2)"

# ── 4. Dex SSO — patch port ───────────────────────────────────────
header "4. Dex SSO"
kubectl patch svc dex -n dex --type=json -p='[
  {"op": "replace", "path": "/spec/ports/0/port", "value": 5557},
  {"op": "replace", "path": "/spec/ports/0/targetPort", "value": 5557}
]' 2>/dev/null || true
ok "Dex port patché (5557)"

# ── 5. Ollama ─────────────────────────────────────────────────────
header "5. Ollama"
if systemctl is-active --quiet ollama; then
  ok "Ollama déjà actif"
else
  sudo systemctl start ollama
  sleep 2
  ok "Ollama démarré"
fi
info "Modèle : $(ollama list 2>/dev/null | grep -v NAME | awk '{print $1}' | head -1 || echo 'aucun')"

# ── 6. Harbor jobservice ──────────────────────────────────────────
header "6. Harbor"
kubectl rollout restart deployment harbor-jobservice -n harbor > /dev/null 2>&1 || true
ok "Harbor jobservice redémarré"

# ── 7. dxp-serve pod K8s ─────────────────────────────────────────
header "7. dxp-serve"
DXP_POD=$(kubectl get pods -n dxp-system -l app=dxp-serve --no-headers 2>/dev/null | grep Running | awk '{print $1}' | head -1)
if [ -n "$DXP_POD" ]; then
  ok "dxp-serve pod Running : $DXP_POD"
  # Port-forward dxp-serve → accessible depuis la VM sur :30890
  pkill -f "port-forward.*dxp-serve" 2>/dev/null || true
  sleep 1
  kubectl port-forward -n dxp-system svc/dxp-serve 30890:8090 \
    < /dev/null > /tmp/dxp-serve-pf.log 2>&1 & disown $!
  sleep 2
  ok "dxp-serve port-forward actif → localhost:30890"
else
  warn "dxp-serve pod non Running — vérifier : kubectl get pods -n dxp-system"
  # Tentative de redémarrage
  kubectl rollout restart deployment dxp-serve -n dxp-system > /dev/null 2>&1 || true
  info "Redémarrage déclenché — attendre 15s puis vérifier"
fi

# ── 8. Taint agent-2 (idempotent) ────────────────────────────────
header "8. agent-2"
if kubectl get node k3d-dxp-poc-agent-2-0 > /dev/null 2>&1; then
  kubectl taint nodes k3d-dxp-poc-agent-2-0 node-role=workload:NoSchedule --overwrite 2>/dev/null || true
  kubectl label node k3d-dxp-poc-agent-2-0 node-role=workload environment=demo --overwrite 2>/dev/null || true
  ok "agent-2 : taint + label workload appliqués"
else
  warn "agent-2 non trouvé — ajouter avec : k3d node create dxp-poc-agent-2 --cluster dxp-poc --role agent"
fi

# ── 10. Résumé ────────────────────────────────────────────────────
echo ""
echo -e "${GREEN}╔══════════════════════════════════════════════╗${RESET}"
echo -e "${GREEN}║   DxP est prêt                               ║${RESET}"
echo -e "${GREEN}╚══════════════════════════════════════════════╝${RESET}"
echo ""
echo "  Nœuds    : $(kubectl get nodes --no-headers | wc -l) (dont 1 agent-2 workload)"
echo "  ArgoCD   : https://158.158.8.131:9443/argocd"
echo "  Harbor   : http://158.158.8.131:9091"
echo "  Grafana  : https://158.158.8.131:9443/grafana"
echo "  Backstage: http://158.158.8.131:7007"
echo "  Airflow  : http://158.158.8.131:9294"
echo "  LiteLLM  : http://158.158.8.131:30096"
echo ""
echo "  dxp serve pod : kubectl get pods -n dxp-system"
echo "  Backstage     : cd ~/dxp-platform/dxp-portal && nvm use 20 && yarn workspace backend start < /dev/null > /tmp/backstage-backend.log 2>&1 & disown \$!"
echo ""
