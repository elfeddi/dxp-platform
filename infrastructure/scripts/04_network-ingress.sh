#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "04 — Ingress Controller (nginx)"

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
