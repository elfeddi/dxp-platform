#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "08 — Vault"

kubectl create namespace vault --dry-run=client -o yaml | kubectl apply -f -

helm upgrade --install vault hashicorp/vault \
  --namespace vault \
  --set server.dev.enabled=true \
  --set server.dev.devRootToken="${VAULT_ROOT_TOKEN}" \
  --wait --timeout 5m

wait_pods vault
ok "Vault installé (mode dev) — token: ${VAULT_ROOT_TOKEN}"
