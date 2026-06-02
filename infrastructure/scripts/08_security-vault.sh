#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "08 — Vault (mode dev)"

kubectl create namespace vault --dry-run=client -o yaml | kubectl apply -f -

helm upgrade --install vault hashicorp/vault \
  --namespace vault \
  --set server.dev.enabled=true \
  --set server.dev.devRootToken="${VAULT_ROOT_TOKEN}" \
  --set server.nodeSelector.role=infra \
  --set server.resources.requests.memory=128Mi \
  --set server.resources.limits.memory=256Mi \
  --wait --timeout 5m

wait_pods vault
ok "Vault installé (mode dev) — Master"
