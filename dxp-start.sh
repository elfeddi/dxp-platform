#!/bin/bash
# DxP Start — Demarrage automatique apres reboot VM

ok()   { echo "[v] $1"; }
warn() { echo "[!] $1"; }
fail() { echo "[x] $1"; exit 1; }
log()  { echo "[->] $1"; }

log "Attente cluster k3s..."
until kubectl get nodes &>/dev/null; do sleep 3; done
kubectl wait --for=condition=ready node --all --timeout=120s
ok "Cluster k3s pret"

log "Verification CoreDNS harbor.dxp..."
if ! kubectl get configmap coredns -n kube-system -o yaml | grep -q "harbor.dxp"; then
  warn "CoreDNS patch manquant - application..."
  kubectl patch configmap coredns -n kube-system --type merge -p '{"data":{"Corefile":".:53 {\n    errors\n    health\n    ready\n    kubernetes cluster.local in-addr.arpa ip6.arpa {\n      pods insecure\n      fallthrough in-addr.arpa ip6.arpa\n    }\n    rewrite name harbor.dxp harbor.harbor.svc.cluster.local\n    prometheus :9153\n    forward . /etc/resolv.conf\n    cache 30\n    loop\n    reload\n    loadbalance\n  }"}}'
  kubectl rollout restart deployment coredns -n kube-system
  kubectl rollout status deployment coredns -n kube-system --timeout=60s
  sleep 10
  kubectl rollout restart deployment -n tekton-pipelines 2>/dev/null || true
  sleep 5
fi
ok "CoreDNS harbor.dxp configure"

# Config registre insecure harbor.dxp sur les Workers
log "Configuration registre Harbor sur les Workers..."
for WORKER_IP in 10.0.0.5 10.0.0.6; do
  ssh -i ~/.ssh/dxp-key.pem -o StrictHostKeyChecking=no azureuser@$WORKER_IP \
    "sudo mkdir -p /etc/rancher/k3s
     grep -q 'harbor.dxp' /etc/hosts || echo '10.0.0.4 harbor.dxp' | sudo tee -a /etc/hosts
     sudo tee /etc/rancher/k3s/registries.yaml << 'YAML'
mirrors:
  harbor.dxp:
    endpoint:
      - \"http://10.0.0.4:30091\"
configs:
  harbor.dxp:
    auth:
      username: admin
      password: dxp-Harbor2026
    tls:
      insecure_skip_verify: true
YAML
     sudo systemctl restart k3s-agent" \
    &>/dev/null && echo "[v] Registre configuré sur $WORKER_IP" || echo "[!] Erreur registre sur $WORKER_IP"
done
log "Attente ArgoCD..."
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=argocd-server -n argocd --timeout=120s
ok "ArgoCD pret"

log "Regeneration token ArgoCD..."
kubectl port-forward -n argocd svc/argocd-server 19090:80 > /tmp/pf-argocd.log 2>&1 &
PF_PID=$!
sleep 20

SESSION_TOKEN=$(curl -sk -X POST http://localhost:19090/api/v1/session \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"dxp-Argocd2026"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])" 2>/dev/null)

if [ -z "$SESSION_TOKEN" ]; then
  kill $PF_PID 2>/dev/null
# Enregistrer les entites catalog Backstage
log "Enregistrement entites Backstage catalog..."
sleep 30
BST_TOKEN=$(curl -s -X POST http://localhost:7007/api/auth/guest/refresh   -H "Content-Type: application/json" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('backstageIdentity',{}).get('token',''))" 2>/dev/null)

for target in   "https://github.com/elfeddi/dxp-platform/blob/main/templates/golden-path-devops/template.yaml"   "https://github.com/elfeddi/dxp-platform/blob/main/templates/golden-path-go/template.yaml"   "https://github.com/elfeddi/dxp-platform/blob/main/templates/golden-path-java/template.yaml"   "https://github.com/elfeddi/dxp-platform/blob/main/templates/golden-path-react/template.yaml"   "https://github.com/elfeddi/dxp-platform/blob/main/templates/golden-path-php/template.yaml"   "https://github.com/elfeddi/dxp-platform/blob/main/stack0/catalog-info.yaml"; do
  curl -s -X POST http://localhost:7007/api/catalog/locations     -H "Content-Type: application/json"     -H "Authorization: Bearer $BST_TOKEN"     -d "{\"type\":\"url\",\"target\":\"$target\"}" &>/dev/null
done
ok "Entites Backstage enregistrees"

echo ""
echo "DxP Platform - Demarrage OK"
