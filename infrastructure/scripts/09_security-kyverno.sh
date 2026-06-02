#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "09 — Kyverno"

kubectl create namespace kyverno --dry-run=client -o yaml | kubectl apply -f -

helm upgrade --install kyverno kyverno/kyverno \
  --namespace kyverno \
  --set admissionController.replicas=1 \
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
ok "Kyverno + policy dxp-require-labels installés"
