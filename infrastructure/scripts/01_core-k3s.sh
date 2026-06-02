#!/bin/bash
source "$(dirname "$0")/lib/common.sh"

title "01 — Cluster k3s natif multi-VM"

# ══════════════════════════════════════════════════════
# Ce script est documentaire — k3s s'installe manuellement
# avant de lancer dxp-install.sh
#
# Prérequis :
#   3 VMs Azure dans le même VNet (B4ms + 2×B2s)
#   Ubuntu 22.04 LTS
#   Clé SSH commune sur les 3 VMs
#
# Étape 1 — Sur le Master :
#   curl -sfL https://get.k3s.io | sh -s - server \
#     --tls-san <IP_PUBLIQUE_MASTER> \
#     --disable traefik \
#     --node-label "node-role=server"
#
#   mkdir -p ~/.kube
#   sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
#   sudo chown azureuser:azureuser ~/.kube/config
#
#   TOKEN=$(sudo cat /var/lib/rancher/k3s/server/node-token)
#
# Étape 2 — Sur Worker1 (depuis le Master via SSH) :
#   ssh -i ~/.ssh/dxp-key.pem azureuser@<IP_PRIVEE_WORKER1>
#   curl -sfL https://get.k3s.io | K3S_URL=https://<IP_PRIVEE_MASTER>:6443 \
#     K3S_TOKEN=<TOKEN> \
#     sh -s - agent --node-label "node-role=observability"
#
# Étape 3 — Sur Worker2 (depuis le Master via SSH) :
#   ssh -i ~/.ssh/dxp-key.pem azureuser@<IP_PRIVEE_WORKER2>
#   curl -sfL https://get.k3s.io | K3S_URL=https://<IP_PRIVEE_MASTER>:6443 \
#     K3S_TOKEN=<TOKEN> \
#     sh -s - agent --node-label "node-role=workload"
#
# Étape 4 — Sur le Master :
#   kubectl taint nodes <NOM_WORKER2> node-role=workload:NoSchedule
#   kubectl label nodes <NOM_WORKER1> role=observability
#   kubectl label nodes <NOM_WORKER2> role=workload
# ══════════════════════════════════════════════════════

log "Vérification cluster k3s..."
kubectl get nodes || fail "Cluster k3s non disponible — installer k3s manuellement (voir commentaires ci-dessus)"

NODE_COUNT=$(kubectl get nodes --no-headers | grep -c Ready)
if [ "$NODE_COUNT" -lt 3 ]; then
  warn "Seulement $NODE_COUNT nœud(s) Ready — 3 attendus (Master + Worker1 + Worker2)"
else
  ok "Cluster k3s prêt — $NODE_COUNT nœuds Ready"
fi

kubectl get nodes -o wide
