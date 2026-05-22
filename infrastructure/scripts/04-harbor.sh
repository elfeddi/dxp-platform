#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "04 — Harbor + CoreDNS"

# Harbor
kubectl create namespace $DXP_NAMESPACE_HARBOR --dry-run=client -o yaml | kubectl apply -f -

helm upgrade --install harbor harbor/harbor \
  --namespace $DXP_NAMESPACE_HARBOR \
  --set expose.type=nodePort \
  --set expose.nodePort.ports.http.nodePort=30091 \
  --set expose.tls.enabled=false \
  --set externalURL=http://harbor.dxp \
  --set harborAdminPassword="${HARBOR_ADMIN_PASSWORD}" \
  --set persistence.enabled=true \
  --wait --timeout 8m

wait_pods $DXP_NAMESPACE_HARBOR

# CoreDNS patch harbor.dxp
log "Patch CoreDNS harbor.dxp..."
kubectl patch configmap coredns -n kube-system --type merge -p \
  '{"data":{"Corefile":".:53 {\n    errors\n    health\n    ready\n    kubernetes cluster.local in-addr.arpa ip6.arpa {\n      pods insecure\n      fallthrough in-addr.arpa ip6.arpa\n    }\n    rewrite name harbor.dxp harbor.harbor.svc.cluster.local\n    prometheus :9153\n    forward . /etc/resolv.conf\n    cache 30\n    loop\n    reload\n    loadbalance\n  }"}}'
kubectl rollout restart deployment coredns -n kube-system
kubectl rollout status deployment coredns -n kube-system --timeout=60s

# Registries k3d pour Harbor insecure
log "Patch registries k3d..."
HARBOR_IP=$(kubectl get svc harbor -n $DXP_NAMESPACE_HARBOR -o jsonpath='{.spec.clusterIP}')
for node in $(k3d node list | grep dxp-poc | grep -v tools | grep -v serverlb | awk '{print $1}'); do
  docker exec "${node}" sh -c "mkdir -p /etc/rancher/k3s && cat > /etc/rancher/k3s/registries.yaml << REGEOF
mirrors:
  harbor.dxp:
    endpoint:
      - http://${HARBOR_IP}
configs:
  harbor.dxp:
    tls:
      insecure_skip_verify: true
REGEOF"
done

ok "Harbor installé — http://${DXP_IP}:9091"
ok "CoreDNS harbor.dxp configuré"
