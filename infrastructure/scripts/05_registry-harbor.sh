#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "05 — Harbor + CoreDNS + registre k3s"

kubectl create namespace harbor --dry-run=client -o yaml | kubectl apply -f -

helm upgrade --install harbor harbor/harbor \
  --namespace harbor \
  --set expose.type=nodePort \
  --set expose.nodePort.ports.http.nodePort=30091 \
  --set expose.tls.enabled=false \
  --set externalURL=http://harbor.dxp \
  --set harborAdminPassword="${HARBOR_ADMIN_PASSWORD}" \
  --set persistence.enabled=true \
  --set nodeSelector.role=infra \
  --set core.resources.requests.memory=256Mi \
  --set core.resources.limits.memory=512Mi \
  --set registry.resources.requests.memory=128Mi \
  --set registry.resources.limits.memory=256Mi \
  --set jobservice.resources.requests.memory=64Mi \
  --set jobservice.resources.limits.memory=128Mi \
  --wait --timeout 8m

wait_pods harbor

# CoreDNS patch harbor.dxp
log "Patch CoreDNS harbor.dxp..."
kubectl patch configmap coredns -n kube-system --type merge -p \
  '{"data":{"Corefile":".:53 {\n    errors\n    health\n    ready\n    kubernetes cluster.local in-addr.arpa ip6.arpa {\n      pods insecure\n      fallthrough in-addr.arpa ip6.arpa\n    }\n    rewrite name harbor.dxp harbor.harbor.svc.cluster.local\n    prometheus :9153\n    forward . /etc/resolv.conf\n    cache 30\n    loop\n    reload\n    loadbalance\n  }"}}'
kubectl rollout restart deployment coredns -n kube-system
kubectl rollout status deployment coredns -n kube-system --timeout=60s

# Registre insecure k3s natif
log "Configuration registre insecure k3s..."
HARBOR_IP=$(kubectl get svc harbor -n harbor -o jsonpath='{.spec.clusterIP}')
sudo mkdir -p /etc/rancher/k3s
sudo tee /etc/rancher/k3s/registries.yaml > /dev/null << REGEOF
mirrors:
  harbor.dxp:
    endpoint:
      - http://${HARBOR_IP}
configs:
  harbor.dxp:
    tls:
      insecure_skip_verify: true
REGEOF

sudo systemctl restart k3s
sleep 20
kubectl wait --for=condition=ready node --all --timeout=60s

ok "Harbor installé — http://${DXP_IP}:9091 — Master"
ok "CoreDNS + registre k3s configurés"
