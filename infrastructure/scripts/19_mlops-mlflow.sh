#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "19 — MLflow"

kubectl create namespace mlops --dry-run=client -o yaml | kubectl apply -f -

helm upgrade --install mlflow community-charts/mlflow \
  --namespace mlops \
  --set image.tag=2.19.0 \
  --set nodeSelector.role=infra \
  --set securityContext.runAsUser=0 \
  --set securityContext.runAsGroup=0 \
  --set securityContext.runAsNonRoot=false \
  --set resources.requests.memory=512Mi \
  --set resources.limits.memory=2Gi \
  --wait --timeout 10m

wait_pods mlops
ok "MLflow installe — Master"
