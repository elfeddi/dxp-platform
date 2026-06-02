#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "09 — Kyverno (HA lite — 2 replicas)"

kubectl create namespace kyverno --dry-run=client -o yaml | kubectl apply -f -

helm upgrade --install kyverno kyverno/kyverno \
  --namespace kyverno \
  --set admissionController.replicas=2 \
  --set admissionController.resources.requests.memory=128Mi \
  --set admissionController.resources.limits.memory=256Mi \
  --wait --timeout 5m

kubectl apply -f - << POLICYEOF
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: dxp-require-labels
spec:
  validationFailureAction: Audit
  rules:
  - name: check-app-label
    match:
      resources:
        kinds:
        - Pod
    validate:
      message: "Le label 'app' est requis sur tous les pods DxP"
      pattern:
        metadata:
          labels:
            app: "?*"
POLICYEOF

wait_pods kyverno
ok "Kyverno installé (2 replicas) — Master + Worker1"
