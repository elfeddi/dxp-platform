#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "12 — Loki (single binary)"

helm upgrade --install loki grafana/loki \
  --namespace monitoring \
  --set deploymentMode=SingleBinary \
  --set loki.commonConfig.replication_factor=1 \
  --set loki.storage.type=filesystem \
  --set loki.useTestSchema=true \
  --set singleBinary.replicas=1 \
  --set singleBinary.nodeSelector.role=observability \
  --set singleBinary.resources.requests.memory=128Mi \
  --set singleBinary.resources.limits.memory=256Mi \
  --set read.replicas=0 \
  --set write.replicas=0 \
  --set backend.replicas=0 \
  --set monitoring.selfMonitoring.enabled=false \
  --set monitoring.serviceMonitor.enabled=false \
  --set test.enabled=false \
  --wait --timeout 5m

wait_pods monitoring
ok "Loki installé — Worker1"
