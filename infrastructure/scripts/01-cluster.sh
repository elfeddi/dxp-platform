#!/bin/bash
source "$(dirname "$0")/lib/common.sh"

title "01 — Cluster k3d"

# Supprimer le cluster existant si présent
if k3d cluster list | grep -q "dxp-poc"; then
  warn "Cluster dxp-poc existant — suppression..."
  
  # Stopper et supprimer les containers k3d manuellement
  docker ps -a --filter "name=k3d-dxp-poc" --format "{{.ID}}" | \
    xargs -r docker stop 2>/dev/null || true
  docker ps -a --filter "name=k3d-dxp-poc" --format "{{.ID}}" | \
    xargs -r docker rm -f 2>/dev/null || true
  
  # Supprimer le réseau Docker
  docker network rm k3d-dxp-poc 2>/dev/null || true
  
  # Supprimer le cluster k3d
  k3d cluster delete dxp-poc 2>/dev/null || true
  sleep 3
fi

log "Création cluster k3d dxp-poc..."
k3d cluster create dxp-poc \
  --servers 1 --agents 2 \
  --k3s-arg "--disable=traefik@server:0" \
  --port "9090:30090@loadbalancer" \
  --port "9091:30091@loadbalancer" \
  --port "8080:30080@loadbalancer" \
  --port "3001:30031@loadbalancer" \
  --port "9292:30092@loadbalancer" \
  --port "9293:30093@loadbalancer" \
  --port "9294:30094@loadbalancer" \
  --port "9295:30095@loadbalancer" \
  --port "9443:30443@loadbalancer" \
  --port "32000:32000@loadbalancer" \
  --port "30096:30096@loadbalancer" \
  --wait

kubectl wait --for=condition=ready node --all --timeout=60s
ok "Cluster k3d prêt — $(kubectl get nodes --no-headers | wc -l) nœuds"
