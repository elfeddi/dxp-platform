#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "17 — Airflow 3.x (KubernetesExecutor)"

kubectl create namespace dataops --dry-run=client -o yaml | kubectl apply -f -

helm upgrade --install airflow apache-airflow/airflow \
  --namespace dataops \
  --set executor=KubernetesExecutor \
  --set createUserJob.defaultUser.password="${AIRFLOW_ADMIN_PASSWORD}" \
  --set apiServer.service.type=NodePort \
  --set apiServer.nodeSelector.role=infra \
  --set apiServer.resources.requests.memory=256Mi \
  --set apiServer.resources.limits.memory=512Mi \
  --set scheduler.nodeSelector.role=infra \
  --set scheduler.resources.requests.memory=256Mi \
  --set scheduler.resources.limits.memory=512Mi \
  --set triggerer.enabled=false \
  --set dagProcessor.enabled=false \
  --set flower.enabled=false \
  --set statsd.enabled=false \
  --set redis.enabled=false \
  --timeout 15m

# Migration manuelle — bug chart airflow 1.21.0 avec Airflow 3.x
log "Migration base de donnees Airflow..."
PGPASSWORD=$(kubectl get secret airflow-postgresql -n dataops -o jsonpath="{.data.postgres-password}" | base64 -d)
kubectl run airflow-migrate \
  --image=apache/airflow:3.2.0 \
  --namespace=dataops \
  --restart=Never \
  --env="AIRFLOW__DATABASE__SQL_ALCHEMY_CONN=postgresql+psycopg2://postgres:${PGPASSWORD}@airflow-postgresql:5432/postgres" \
  -- bash -c "airflow db migrate"
kubectl wait --for=condition=ready pod/airflow-migrate -n dataops --timeout=120s
kubectl logs -n dataops airflow-migrate | tail -5
kubectl delete pod airflow-migrate -n dataops

# Attendre que les pods demarrent
kubectl wait --for=condition=ready pod -l component=api-server -n dataops --timeout=300s
kubectl wait --for=condition=ready pod -l component=scheduler -n dataops --timeout=300s

# Patch NodePort
kubectl patch svc airflow-api-server -n dataops \
  --type merge -p '{"spec":{"ports":[{"port":8080,"targetPort":8080,"nodePort":30094}]}}'

ok "Airflow installe — http://${DXP_IP}:9294 — Master"
