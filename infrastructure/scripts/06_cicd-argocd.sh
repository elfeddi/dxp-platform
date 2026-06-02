#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "06 — ArgoCD (HA lite — 2 replicas)"

kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -

helm upgrade --install argocd argo/argo-cd \
  --namespace argocd \
  --set server.replicas=2 \
  --set server.nodeSelector.role=infra \
  --set repoServer.nodeSelector.role=infra \
  --set applicationSet.nodeSelector.role=infra \
  --set controller.nodeSelector.role=infra \
  --set redis.nodeSelector.role=infra \
  --set server.service.type=NodePort \
  --set server.service.nodePortHttp=30090 \
  --set server.extraArgs[0]="--insecure" \
  --set configs.secret.argocdServerAdminPassword="${ARGOCD_ADMIN_PASSWORD_HASH}" \
  --set server.resources.requests.memory=128Mi \
  --set server.resources.limits.memory=512Mi \
  --set repoServer.resources.requests.memory=128Mi \
  --set repoServer.resources.limits.memory=256Mi \
  --wait --timeout 5m

kubectl patch configmap argocd-cm -n argocd \
  --type merge -p '{"data":{"accounts.admin":"apiKey,login"}}'

wait_pods argocd
ok "ArgoCD installé (2 replicas) — http://${DXP_IP}:9090 — Master"
