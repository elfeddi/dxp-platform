#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "22 — pgvector"

kubectl create namespace llmops --dry-run=client -o yaml | kubectl apply -f -

helm upgrade --install pgvector bitnami/postgresql \
  --namespace llmops \
  --set auth.password="${PGVECTOR_PASSWORD}" \
  --set auth.database=dxp_vectors \
  --wait --timeout 3m

wait_pods llmops
ok "pgvector installé"
