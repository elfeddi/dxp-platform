#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "05 — Tekton + Triggers + Dashboard"

# Tekton Pipelines
log "Installation Tekton Pipelines..."
kubectl apply -f https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml
kubectl wait --for=condition=ready pod \
  -l app.kubernetes.io/part-of=tekton-pipelines \
  -n $DXP_NAMESPACE_TEKTON --timeout=3m

# Tekton Triggers
log "Installation Tekton Triggers..."
kubectl apply -f https://storage.googleapis.com/tekton-releases/triggers/latest/release.yaml
kubectl apply -f https://storage.googleapis.com/tekton-releases/triggers/latest/interceptors.yaml
kubectl wait --for=condition=ready pod \
  -l app.kubernetes.io/part-of=tekton-triggers \
  -n $DXP_NAMESPACE_TEKTON --timeout=3m

# Tekton Dashboard
log "Installation Tekton Dashboard..."
kubectl apply -f https://storage.googleapis.com/tekton-releases/dashboard/latest/release.yaml
kubectl wait --for=condition=ready pod \
  -l app=tekton-dashboard \
  -n $DXP_NAMESPACE_TEKTON --timeout=3m

# NodePort Tekton Dashboard
kubectl apply -f - << SVCEOF
apiVersion: v1
kind: Service
metadata:
  name: tekton-dashboard-nodeport
  namespace: $DXP_NAMESPACE_TEKTON
spec:
  type: NodePort
  selector:
    app: tekton-dashboard
  ports:
  - port: 9097
    targetPort: 9097
    nodePort: 30095
SVCEOF

# NodePort Tekton Webhook EventListener (créé plus tard via GitOps)
ok "Tekton installé — Dashboard http://${DXP_IP}:9295"
