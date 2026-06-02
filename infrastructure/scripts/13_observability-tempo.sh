#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "13 — Tempo"

helm upgrade --install tempo grafana/tempo \
  --namespace monitoring \
  --set tempo.storage.trace.backend=local \
  --wait --timeout 3m

wait_pods monitoring
ok "Tempo installé"
