#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "03 — ArgoCD"

kubectl create namespace $DXP_NAMESPACE_ARGOCD --dry-run=client -o yaml | kubectl apply -f -

helm upgrade --install argocd argo/argo-cd \
  --namespace $DXP_NAMESPACE_ARGOCD \
  --set server.service.type=NodePort \
  --set server.service.nodePortHttp=30090 \
  --set server.extraArgs[0]="--insecure" \
  --set configs.secret.argocdServerAdminPassword="${ARGOCD_ADMIN_PASSWORD_HASH}" \
  --wait --timeout 5m

kubectl patch configmap argocd-cm -n $DXP_NAMESPACE_ARGOCD \
  --type merge -p '{"data":{"accounts.admin":"apiKey,login"}}'

wait_pods $DXP_NAMESPACE_ARGOCD
ok "ArgoCD installé — http://${DXP_IP}:9090"
