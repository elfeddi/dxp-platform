#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "06 — Observabilité LGTM"

kubectl create namespace $DXP_NAMESPACE_MONITORING --dry-run=client -o yaml | kubectl apply -f -

# inotify pour Promtail
log "Configuration inotify..."
sudo sysctl fs.inotify.max_user_instances=512
sudo sysctl fs.inotify.max_user_watches=524288
grep -q "fs.inotify.max_user_instances" /etc/sysctl.conf || \
  echo "fs.inotify.max_user_instances=512" | sudo tee -a /etc/sysctl.conf
grep -q "fs.inotify.max_user_watches" /etc/sysctl.conf || \
  echo "fs.inotify.max_user_watches=524288" | sudo tee -a /etc/sysctl.conf

# Prometheus + Grafana + Alertmanager
log "Installation kube-prometheus-stack..."
helm upgrade --install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace $DXP_NAMESPACE_MONITORING \
  --set grafana.adminPassword="${GRAFANA_ADMIN_PASSWORD}" \
  --set grafana.service.type=NodePort \
  --set grafana.service.nodePort=30031 \
  --set grafana.grafana\.ini.server.domain="${DXP_IP}" \
  --set grafana.grafana\.ini.server.root_url="http://${DXP_IP}:3001" \
  --set prometheus-node-exporter.hostRootFsMount.enabled=false \
  --set alertmanager.service.type=NodePort \
  --set alertmanager.service.nodePort=30093 \
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
  --wait --timeout 8m

# Loki
log "Installation Loki..."
helm upgrade --install loki grafana/loki \
  --namespace $DXP_NAMESPACE_MONITORING \
  --set deploymentMode=SingleBinary \
  --set loki.commonConfig.replication_factor=1 \
  --set loki.storage.type=filesystem \
  --set loki.useTestSchema=true \
  --set singleBinary.replicas=1 \
  --set read.replicas=0 \
  --set write.replicas=0 \
  --set backend.replicas=0 \
  --set monitoring.selfMonitoring.enabled=false \
  --set monitoring.serviceMonitor.enabled=false \
  --set test.enabled=false \
  --wait --timeout 5m

# Tempo
log "Installation Tempo..."
helm upgrade --install tempo grafana/tempo \
  --namespace $DXP_NAMESPACE_MONITORING \
  --set tempo.storage.trace.backend=local \
  --wait --timeout 3m

# OTel Collector
log "Installation OTel Collector..."
helm upgrade --install otel-collector open-telemetry/opentelemetry-collector \
  --namespace $DXP_NAMESPACE_MONITORING \
  --set mode=daemonset \
  --set image.repository=otel/opentelemetry-collector-contrib \
  --set config.receivers.otlp.protocols.grpc.endpoint="0.0.0.0:4317" \
  --set config.receivers.otlp.protocols.http.endpoint="0.0.0.0:4318" \
  --set config.exporters.otlp/tempo.endpoint="tempo.monitoring.svc.cluster.local:4317" \
  --set config.exporters.otlp/tempo.tls.insecure=true \
  --wait --timeout 5m

wait_pods $DXP_NAMESPACE_MONITORING 300
ok "LGTM installé — Grafana http://${DXP_IP}:3001"
