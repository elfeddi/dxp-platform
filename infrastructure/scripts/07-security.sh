#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "07 — Security (Kyverno + cert-manager + Falco + Vault)"

# Kyverno
log "Installation Kyverno..."
kubectl create namespace $DXP_NAMESPACE_KYVERNO --dry-run=client -o yaml | kubectl apply -f -
helm upgrade --install kyverno kyverno/kyverno \
  --namespace $DXP_NAMESPACE_KYVERNO \
  --set admissionController.replicas=1 \
  --wait --timeout 5m

# Policy baseline
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
ok "Kyverno + policy dxp-require-labels installés"

# cert-manager
log "Installation cert-manager..."
kubectl create namespace $DXP_NAMESPACE_CERT_MANAGER --dry-run=client -o yaml | kubectl apply -f -
helm upgrade --install cert-manager jetstack/cert-manager \
  --namespace $DXP_NAMESPACE_CERT_MANAGER \
  --set crds.enabled=true \
  --wait --timeout 5m

# ClusterIssuer self-signed
kubectl apply -f - << ISSUEREOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: dxp-ca-issuer
spec:
  selfSigned: {}
ISSUEREOF
ok "cert-manager + ClusterIssuer dxp-ca-issuer installés"

# Falco
log "Installation Falco..."
kubectl create namespace $DXP_NAMESPACE_FALCO --dry-run=client -o yaml | kubectl apply -f -
helm upgrade --install falco falcosecurity/falco \
  --namespace $DXP_NAMESPACE_FALCO \
  --set driver.kind=modern_ebpf \
  --set tty=true \
  --wait --timeout 5m
ok "Falco installé"

# Vault
log "Installation Vault..."
kubectl create namespace $DXP_NAMESPACE_VAULT --dry-run=client -o yaml | kubectl apply -f -
helm upgrade --install vault hashicorp/vault \
  --namespace $DXP_NAMESPACE_VAULT \
  --set server.dev.enabled=true \
  --set server.dev.devRootToken="${VAULT_ROOT_TOKEN}" \
  --wait --timeout 5m
ok "Vault installé (mode dev)"

wait_pods $DXP_NAMESPACE_KYVERNO
wait_pods $DXP_NAMESPACE_CERT_MANAGER
wait_pods $DXP_NAMESPACE_FALCO
wait_pods $DXP_NAMESPACE_VAULT
ok "Stack Security complète"
