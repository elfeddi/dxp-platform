#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "11 — Prometheus + Grafana + Alertmanager"

kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -

sudo sysctl fs.inotify.max_user_instances=512
sudo sysctl fs.inotify.max_user_watches=524288
grep -q "fs.inotify.max_user_instances" /etc/sysctl.conf || \
  echo "fs.inotify.max_user_instances=512" | sudo tee -a /etc/sysctl.conf
grep -q "fs.inotify.max_user_watches" /etc/sysctl.conf || \
  echo "fs.inotify.max_user_watches=524288" | sudo tee -a /etc/sysctl.conf

helm upgrade --install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --set grafana.adminPassword="${GRAFANA_ADMIN_PASSWORD}" \
  --set grafana.service.type=NodePort \
  --set grafana.service.nodePort=30031 \
  --set grafana.grafana\.ini.server.domain="${DXP_IP}" \
  --set grafana.grafana\.ini.server.root_url="http://${DXP_IP}:3001" \
  --set prometheus-node-exporter.hostRootFsMount.enabled=false \
  --set alertmanager.service.type=NodePort \
  --set alertmanager.service.nodePort=30093 \
  --wait --timeout 8m

wait_pods monitoring
ok "Prometheus + Grafana + Alertmanager installés — http://${DXP_IP}:3001"
