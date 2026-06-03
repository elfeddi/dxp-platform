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
HARBOR_IP=$(kubectl get svc harbor -n harbor -o jsonpath='{.spec.clusterIP}' 2>/dev/null)
if [ -n "$HARBOR_IP" ]; then
  for WORKER_IP in 10.0.0.5 10.0.0.6; do
    ssh -i ~/.ssh/dxp-key.pem -o StrictHostKeyChecking=no azureuser@$WORKER_IP       "sudo mkdir -p /etc/rancher/k3s && echo \"mirrors:\n  harbor.dxp:\n    endpoint:\n      - http://${HARBOR_IP}\nconfigs:\n  harbor.dxp:\n    tls:\n      insecure_skip_verify: true\" | sudo tee /etc/rancher/k3s/registries.yaml > /dev/null && sudo systemctl restart k3s-agent"       &>/dev/null && echo "[v] Registre configuré sur $WORKER_IP" || echo "[!] Erreur registre sur $WORKER_IP"
  done
fi

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
  cat /tmp/pf-argocd.log
  fail "Impossible de se connecter a ArgoCD"
fi

ARGOCD_API_TOKEN=$(curl -sk -X POST "http://localhost:19090/api/v1/account/admin/token" \
  -H "Authorization: Bearer $SESSION_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"dxp-start-$(date +%s)\",\"expiresIn\":0}" | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])" 2>/dev/null)

kill $PF_PID 2>/dev/null

if [ -z "$ARGOCD_API_TOKEN" ]; then
  fail "Impossible de generer le token ArgoCD API"
fi

source ~/dxp-platform/infrastructure/scripts/.env
kubectl create secret generic dxp-serve-env -n dxp-system \
  --from-literal=ARGOCD_TOKEN="$ARGOCD_API_TOKEN" \
  --from-literal=HARBOR_API_TOKEN="$HARBOR_API_TOKEN" \
  --from-literal=GRAFANA_API_TOKEN="$GRAFANA_API_TOKEN" \
  --from-literal=LITELLM_API_KEY="$LITELLM_API_KEY" \
  --from-literal=LITELLM_API_BASE="$LITELLM_API_BASE" \
  --from-literal=LITELLM_MODEL="$LITELLM_MODEL" \
  --from-literal=GITHUB_TOKEN="$GITHUB_TOKEN" \
  --dry-run=client -o yaml | kubectl apply -f - &>/dev/null

sed -i "s|ARGOCD_API_TOKEN=.*|ARGOCD_API_TOKEN=$ARGOCD_API_TOKEN|" \
  ~/dxp-platform/infrastructure/scripts/.env

ok "Token ArgoCD regenere"

kubectl apply -f ~/dxp-platform/infrastructure/configs/dxp-rbac.yaml &>/dev/null && ok "RBAC dxp-provisioner appliqué"
log "Redemarrage dxp-serve..."
kubectl rollout restart deployment dxp-serve -n dxp-system &>/dev/null
kubectl rollout status deployment dxp-serve -n dxp-system --timeout=60s
ok "dxp-serve redemarre"

log "Verification dxp-serve..."
kubectl port-forward -n dxp-system svc/dxp-serve 18090:8090 > /tmp/pf-dxp.log 2>&1 &
PF_PID=$!
sleep 5

STATUS=$(curl -s -H "Authorization: Bearer viewer" http://localhost:18090/api/dxp/status | \
  python3 -c "import sys,json; d=json.load(sys.stdin); print('ready' if d.get('ready') else 'not ready')" 2>/dev/null)

kill $PF_PID 2>/dev/null

if [ "$STATUS" = "ready" ]; then
  ok "dxp-serve ready — tous les backends operationnels"
else
  warn "dxp-serve not ready — verifier manuellement"
fi

# Enregistrer les entites catalog Backstage
log "Enregistrement entites Backstage catalog..."
sleep 30
curl -s -X POST http://localhost:7007/api/catalog/locations   -H "Content-Type: application/json"   -d '{"type":"url","target":"https://github.com/elfeddi/dxp-platform/blob/main/templates/golden-path-devops/template.yaml"}' &>/dev/null
curl -s -X POST http://localhost:7007/api/catalog/locations   -H "Content-Type: application/json"   -d '{"type":"url","target":"https://github.com/elfeddi/dxp-platform/blob/main/stack0/catalog-info.yaml"}' &>/dev/null
ok "Entites Backstage enregistrees"


# Enregistrer les entites catalog Backstage
log "Enregistrement entites Backstage catalog..."
sleep 30
BST_TOKEN=$(curl -s -X POST http://localhost:7007/api/auth/guest/refresh \
  -H "Content-Type: application/json" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get(\'backstageIdentity\',{}).get(\'token\',[\'\'[0]))" 2>/dev/null)
curl -s -X POST http://localhost:7007/api/catalog/locations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $BST_TOKEN" \
  -d '{"type":"url","target":"https://github.com/elfeddi/dxp-platform/blob/main/templates/golden-path-devops/template.yaml"}' &>/dev/null
curl -s -X POST http://localhost:7007/api/catalog/locations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $BST_TOKEN" \
  -d '{"type":"url","target":"https://github.com/elfeddi/dxp-platform/blob/main/stack0/catalog-info.yaml"}' &>/dev/null
ok "Entites Backstage enregistrees"

echo ""
echo "DxP Platform - Demarrage OK"
