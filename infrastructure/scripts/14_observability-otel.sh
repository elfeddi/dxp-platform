#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "14 — OTel Collector"

helm upgrade --install otel-collector open-telemetry/opentelemetry-collector \
  --namespace monitoring \
  --set mode=daemonset \
  --set image.repository=otel/opentelemetry-collector-contrib \
  --set config.receivers.otlp.protocols.grpc.endpoint="0.0.0.0:4317" \
  --set config.receivers.otlp.protocols.http.endpoint="0.0.0.0:4318" \
  --set config.exporters.otlp/tempo.endpoint="tempo.monitoring.svc.cluster.local:4317" \
  --set config.exporters.otlp/tempo.tls.insecure=true \
  --wait --timeout 5m

wait_pods monitoring
ok "OTel Collector installé"
