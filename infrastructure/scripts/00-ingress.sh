#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "00 — Ingress Controller + cert-manager + TLS"

# nginx Ingress Controller
log "Installation nginx Ingress Controller..."
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace \
  --set controller.service.type=NodePort \
  --set controller.service.nodePorts.http=30080 \
  --set controller.service.nodePorts.https=30444 \
  --wait --timeout 3m

wait_pods ingress-nginx
ok "nginx Ingress Controller installé"

# cert-manager (si pas déjà installé par 07-security.sh)
if ! kubectl get namespace cert-manager &>/dev/null; then
  log "Installation cert-manager..."
  kubectl create namespace cert-manager
  helm upgrade --install cert-manager jetstack/cert-manager \
    --namespace cert-manager \
    --set crds.enabled=true \
    --wait --timeout 5m
  ok "cert-manager installé"
else
  ok "cert-manager déjà présent"
fi

# ClusterIssuer self-signed pour le POC
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
ok "ClusterIssuer dxp-ca-issuer prêt"

