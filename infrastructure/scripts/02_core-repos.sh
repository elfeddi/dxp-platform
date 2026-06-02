#!/bin/bash
source "$(dirname "$0")/lib/common.sh"

title "02 — Repos Helm"

helm repo add argo             https://argoproj.github.io/argo-helm
helm repo add harbor           https://helm.goharbor.io
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add grafana          https://grafana.github.io/helm-charts
helm repo add open-telemetry   https://open-telemetry.github.io/opentelemetry-helm-charts
helm repo add kyverno          https://kyverno.github.io/kyverno/
helm repo add jetstack         https://charts.jetstack.io
helm repo add falcosecurity    https://falcosecurity.github.io/charts
helm repo add hashicorp        https://helm.releases.hashicorp.com
helm repo add apache-airflow   https://airflow.apache.org
helm repo add community-charts https://community-charts.github.io/helm-charts
helm repo add bitnami          https://charts.bitnami.com/bitnami
helm repo add dex              https://charts.dexidp.io
helm repo update

ok "Tous les repos Helm ajoutés et mis à jour"
