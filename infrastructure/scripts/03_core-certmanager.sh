#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "03 — cert-manager + ClusterIssuer"

kubectl create namespace cert-manager --dry-run=client -o yaml | kubectl apply -f -

helm upgrade --install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --set crds.enabled=true \
  --set nodeSelector.role=infra \
  --set webhook.nodeSelector.role=infra \
  --set cainjector.nodeSelector.role=infra \
  --set resources.requests.memory=64Mi \
  --set resources.limits.memory=256Mi \
  --wait --timeout 5m

kubectl apply -f - << ISSEOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: dxp-selfsigned
spec:
  selfSigned: {}
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: dxp-ca
  namespace: cert-manager
spec:
  isCA: true
  commonName: dxp-ca
  secretName: dxp-ca-secret
  issuerRef:
    name: dxp-selfsigned
    kind: ClusterIssuer
---
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: dxp-ca-issuer
spec:
  ca:
    secretName: dxp-ca-secret
ISSEOF

sleep 5
kubectl get clusterissuer
ok "cert-manager + ClusterIssuer dxp-ca-issuer prêts — Master"
