#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "08 — DataOps (Airflow + dbt + Great Expectations)"

kubectl create namespace $DXP_NAMESPACE_DATAOPS --dry-run=client -o yaml | kubectl apply -f -

# Airflow
log "Installation Airflow..."
helm upgrade --install airflow apache-airflow/airflow \
  --namespace $DXP_NAMESPACE_DATAOPS \
  --set executor=KubernetesExecutor \
  --set createUserJob.defaultUser.password="${AIRFLOW_ADMIN_PASSWORD}" \
  --set webserver.service.type=NodePort \
  --set webserver.service.nodePort=30094 \
  --set workers.celery.replicas=0 \
  --set triggerer.enabled=false \
  --set dagProcessor.enabled=false \
  --set flower.enabled=false \
  --set statsd.enabled=false \
  --set redis.enabled=false \
  --wait --timeout 10m

wait_pods $DXP_NAMESPACE_DATAOPS 300
ok "Airflow installé — http://${DXP_IP}:9294"

# dbt + Great Expectations dans virtualenv
log "Installation dbt + Great Expectations..."
if [ ! -d "$HOME/dbt-env" ]; then
  python3 -m venv ~/dbt-env
fi
~/dbt-env/bin/pip install dbt-core dbt-postgres great-expectations feast \
  --quiet 2>&1 | tail -3

# Ajouter au PATH si pas déjà fait
grep -q "dbt-env" ~/.bashrc || \
  echo 'export PATH="$HOME/dbt-env/bin:$PATH"' >> ~/.bashrc

~/dbt-env/bin/dbt --version | head -2
ok "dbt + Great Expectations + Feast installés"
