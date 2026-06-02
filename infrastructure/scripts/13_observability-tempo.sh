#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "13 — Tempo (all-in-one)"

helm upgrade --install tempo grafana/tempo \
  --namespace monitoring \
  --set tempo.storage.trace.backend=local \
  --set nodeSelector.role=observability \
  --set resources.requests.memory=64Mi \
  --set resources.limits.memory=256Mi \
  --wait --timeout 3m

wait_pods monitoring
ok "Tempo installé — Worker1"
