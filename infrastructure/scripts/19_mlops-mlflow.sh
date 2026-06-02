#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "09 — MLOps (MLflow + Feast)"

kubectl create namespace $DXP_NAMESPACE_MLOPS --dry-run=client -o yaml | kubectl apply -f -

# MLflow
log "Installation MLflow..."
helm upgrade --install mlflow community-charts/mlflow \
  --namespace $DXP_NAMESPACE_MLOPS \
  --set service.type=NodePort \
  --set backendStore.postgres.enabled=false \
  --set backendStore.databaseMigration=false \
  --wait --timeout 5m

wait_pods $DXP_NAMESPACE_MLOPS
ok "MLflow installé"

# Feast (déjà dans dbt-env installé en 08)
log "Vérification Feast..."
~/dbt-env/bin/python -c "import feast; print('Feast version:', feast.__version__)" \
  && ok "Feast disponible" \
  || warn "Feast non trouvé — relancer 08-dataops.sh"

ok "Stack MLOps complète"
