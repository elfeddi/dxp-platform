#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "10 — Falco (optionnel — désactivé sur POC)"

# ADR S0-017 : Falco désactivé sur VM POC (surcharge CPU)
# Décommenter pour activation en production

warn "Falco désactivé sur POC (ADR S0-017 — surcharge CPU)"
warn "Pour activer en production, décommenter le bloc ci-dessous"

# kubectl create namespace falco --dry-run=client -o yaml | kubectl apply -f -
# helm upgrade --install falco falcosecurity/falco \
#   --namespace falco \
#   --set driver.kind=modern_ebpf \
#   --set tty=true \
#   --wait --timeout 5m
# wait_pods falco
# ok "Falco installé"

ok "Falco — skippé (POC)"
