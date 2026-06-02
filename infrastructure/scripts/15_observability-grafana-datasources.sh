#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "15 — Grafana datasources (Loki + Tempo)"

helm upgrade kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --reuse-values \
  --set grafana.additionalDataSources[0].name=Loki \
  --set grafana.additionalDataSources[0].type=loki \
  --set grafana.additionalDataSources[0].url="http://loki-gateway.monitoring.svc.cluster.local" \
  --set grafana.additionalDataSources[0].isDefault=false \
  --set grafana.additionalDataSources[0].jsonData.httpHeaderName1="X-Scope-OrgID" \
  --set grafana.additionalDataSources[0].secureJsonData.httpHeaderValue1="dxp" \
  --set grafana.additionalDataSources[1].name=Tempo \
  --set grafana.additionalDataSources[1].type=tempo \
  --set grafana.additionalDataSources[1].url="http://tempo.monitoring.svc.cluster.local:3100" \
  --set grafana.additionalDataSources[1].isDefault=false \
  --wait --timeout 3m

kubectl rollout restart deployment/kube-prometheus-stack-grafana -n monitoring
kubectl rollout status deployment/kube-prometheus-stack-grafana -n monitoring --timeout=2m

ok "Datasources Loki + Tempo configurées dans Grafana"
