#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "07 — Tekton Pipelines + Triggers + Dashboard"

kubectl apply -f https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml
kubectl wait --for=condition=ready pod \
  -l app.kubernetes.io/part-of=tekton-pipelines \
  -n tekton-pipelines --timeout=3m

kubectl apply -f https://storage.googleapis.com/tekton-releases/triggers/latest/release.yaml
kubectl apply -f https://storage.googleapis.com/tekton-releases/triggers/latest/interceptors.yaml
kubectl wait --for=condition=ready pod \
  -l app.kubernetes.io/part-of=tekton-triggers \
  -n tekton-pipelines --timeout=3m

kubectl apply -f https://storage.googleapis.com/tekton-releases/dashboard/latest/release.yaml
kubectl wait --for=condition=ready pod \
  -l app=tekton-dashboard \
  -n tekton-pipelines --timeout=3m

# NodePort Dashboard
kubectl apply -f - << SVCEOF
apiVersion: v1
kind: Service
metadata:
  name: tekton-dashboard-nodeport
  namespace: tekton-pipelines
spec:
  type: NodePort
  selector:
    app: tekton-dashboard
  ports:
  - port: 9097
    targetPort: 9097
    nodePort: 30095
SVCEOF

# Patch nodeSelector sur les deployments Tekton
for deploy in tekton-pipelines-controller tekton-pipelines-webhook; do
  kubectl patch deployment $deploy -n tekton-pipelines \
    --type merge -p '{"spec":{"template":{"spec":{"nodeSelector":{"role":"infra"}}}}}'
done

ok "Tekton installé — http://${DXP_IP}:9295 — Master"
