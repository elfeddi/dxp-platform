#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "19 — MLflow"

kubectl create namespace mlops --dry-run=client -o yaml | kubectl apply -f -

helm upgrade --install mlflow community-charts/mlflow \
  --namespace mlops \
  --set service.type=NodePort \
  --set nodeSelector.role=infra \
  --set backendStore.postgres.enabled=false \
  --set backendStore.databaseMigration=false \
  --set resources.requests.memory=128Mi \
  --set resources.limits.memory=256Mi \
  --wait --timeout 5m

wait_pods mlops
ok "MLflow installé — Master"
