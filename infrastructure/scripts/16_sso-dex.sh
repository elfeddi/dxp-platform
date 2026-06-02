#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "11 — Dex SSO"

kubectl create namespace $DXP_NAMESPACE_DEX --dry-run=client -o yaml | kubectl apply -f -

cat > /tmp/dex-values.yaml << DEXEOF
config:
  issuer: http://${DXP_IP}:32000
  storage:
    type: kubernetes
    config:
      inCluster: true
  web:
    http: 0.0.0.0:5556
  connectors:
  - type: github
    id: github
    name: GitHub
    config:
      clientID: ${DEX_GITHUB_CLIENT_ID}
      clientSecret: ${DEX_GITHUB_CLIENT_SECRET}
      redirectURI: http://${DXP_IP}:32000/callback
  staticClients:
  - id: dxp-argocd
    name: ArgoCD
    secret: dxp-argocd-secret
    redirectURIs:
    - http://${DXP_IP}:9090/auth/callback
  - id: dxp-grafana
    name: Grafana
    secret: dxp-grafana-secret
    redirectURIs:
    - http://${DXP_IP}:3001/login/generic_oauth
  - id: dxp-harbor
    name: Harbor
    secret: dxp-harbor-secret
    redirectURIs:
    - http://${DXP_IP}:9091/c/oidc/callback
service:
  type: NodePort
  ports:
    http:
      nodePort: 32000
DEXEOF

helm upgrade --install dex dex/dex \
  --namespace $DXP_NAMESPACE_DEX \
  -f /tmp/dex-values.yaml \
  --wait --timeout 3m

wait_pods $DXP_NAMESPACE_DEX

# Configurer ArgoCD pour Dex
log "Configuration ArgoCD + Dex..."
kubectl patch configmap argocd-cm -n $DXP_NAMESPACE_ARGOCD --type merge -p "{
  \"data\": {
    \"url\": \"http://${DXP_IP}:9090\",
    \"oidc.config\": \"name: GitHub via Dex\nissuer: http://${DXP_IP}:32000\nclientID: dxp-argocd\nclientSecret: dxp-argocd-secret\nrequestedScopes: [\\\"openid\\\", \\\"profile\\\", \\\"email\\\", \\\"groups\\\"]\n\"
  }
}"
kubectl rollout restart deployment/argocd-server -n $DXP_NAMESPACE_ARGOCD
kubectl rollout status deployment/argocd-server -n $DXP_NAMESPACE_ARGOCD --timeout=2m

# Configurer Grafana pour Dex
log "Configuration Grafana + Dex..."
kubectl patch configmap kube-prometheus-stack-grafana -n $DXP_NAMESPACE_MONITORING --type merge -p "{
  \"data\": {
    \"grafana.ini\": \"[analytics]\ncheck_for_updates = true\n[log]\nmode = console\n[paths]\ndata = /var/lib/grafana/\nlogs = /var/log/grafana\nplugins = /var/lib/grafana/plugins\nprovisioning = /etc/grafana/provisioning\n[server]\ndomain = ${DXP_IP}\nroot_url = http://${DXP_IP}:3001\n[unified_storage]\nindex_path = /var/lib/grafana-search/bleve\n[auth.generic_oauth]\nenabled = true\nname = GitHub via Dex\nclient_id = dxp-grafana\nclient_secret = dxp-grafana-secret\nscopes = openid profile email groups\nauth_url = http://${DXP_IP}:32000/auth\ntoken_url = http://${DXP_IP}:32000/token\napi_url = http://${DXP_IP}:32000/userinfo\nallow_sign_up = true\n\"
  }
}"
kubectl rollout restart deployment/kube-prometheus-stack-grafana -n $DXP_NAMESPACE_MONITORING
kubectl rollout status deployment/kube-prometheus-stack-grafana -n $DXP_NAMESPACE_MONITORING --timeout=2m

ok "Dex SSO installé — http://${DXP_IP}:32000"
ok "ArgoCD + Grafana configurés avec SSO GitHub via Dex"
