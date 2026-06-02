#!/bin/bash
source "$(dirname "$0")/lib/common.sh"

title "01 — Vérification cluster k3s + labels"

# Vérification cluster
kubectl get nodes || fail "Cluster k3s non disponible"

NODE_COUNT=$(kubectl get nodes --no-headers | grep -c Ready)
if [ "$NODE_COUNT" -lt 3 ]; then
  fail "Seulement $NODE_COUNT nœud(s) Ready — 3 attendus (Master + Worker1 + Worker2)"
fi
ok "Cluster k3s prêt — $NODE_COUNT nœuds Ready"

# Labels par rôle
log "Application des labels..."
MASTER=$(kubectl get nodes --no-headers | grep control-plane | awk '{print $1}')
WORKER1=$(kubectl get nodes --no-headers | grep -v control-plane | awk 'NR==1{print $1}')
WORKER2=$(kubectl get nodes --no-headers | grep -v control-plane | awk 'NR==2{print $1}')

kubectl label node $MASTER  role=infra         --overwrite
kubectl label node $WORKER1 role=observability --overwrite
kubectl label node $WORKER2 role=workload      --overwrite

# Taint Worker2 — apps clients uniquement
kubectl taint node $WORKER2 role=workload:NoSchedule --overwrite 2>/dev/null || true

kubectl get nodes --show-labels
ok "Labels et taints appliqués"
echo ""
log "Architecture POC :"
echo "  Master  (role=infra)          → $MASTER"
echo "  Worker1 (role=observability)  → $WORKER1"
echo "  Worker2 (role=workload)       → $WORKER2"
