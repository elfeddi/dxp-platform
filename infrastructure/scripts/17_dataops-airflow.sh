#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "17 — Airflow (KubernetesExecutor)"

kubectl create namespace dataops --dry-run=client -o yaml | kubectl apply -f -

helm upgrade --install airflow apache-airflow/airflow \
  --namespace dataops \
  --set executor=KubernetesExecutor \
  --set createUserJob.defaultUser.password="${AIRFLOW_ADMIN_PASSWORD}" \
  --set webserver.service.type=NodePort \
  --set webserver.service.nodePort=30094 \
  --set webserver.nodeSelector.role=infra \
  --set scheduler.nodeSelector.role=infra \
  --set webserver.resources.requests.memory=256Mi \
  --set webserver.resources.limits.memory=512Mi \
  --set scheduler.resources.requests.memory=256Mi \
  --set scheduler.resources.limits.memory=512Mi \
  --set workers.celery.replicas=0 \
  --set triggerer.enabled=false \
  --set dagProcessor.enabled=false \
  --set flower.enabled=false \
  --set statsd.enabled=false \
  --set redis.enabled=false \
  --wait --timeout 10m

wait_pods dataops 300
ok "Airflow installé — http://${DXP_IP}:9294 — Master"
