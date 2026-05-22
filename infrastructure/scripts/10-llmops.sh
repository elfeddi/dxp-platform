#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "10 — LLMOps (LiteLLM + pgvector)"

kubectl create namespace $DXP_NAMESPACE_LLMOPS --dry-run=client -o yaml | kubectl apply -f -

# LiteLLM
log "Installation LiteLLM..."
kubectl apply -f - << LITELLMEOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: litellm
  namespace: ${DXP_NAMESPACE_LLMOPS}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: litellm
  template:
    metadata:
      labels:
        app: litellm
    spec:
      containers:
      - name: litellm
        image: ghcr.io/berriai/litellm:main-latest
        ports:
        - containerPort: 4000
        env:
        - name: LITELLM_MASTER_KEY
          value: "${LITELLM_MASTER_KEY}"
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: litellm
  namespace: ${DXP_NAMESPACE_LLMOPS}
spec:
  type: NodePort
  selector:
    app: litellm
  ports:
  - port: 4000
    targetPort: 4000
    nodePort: 30096
LITELLMEOF

# pgvector
log "Installation pgvector..."
helm upgrade --install pgvector bitnami/postgresql \
  --namespace $DXP_NAMESPACE_LLMOPS \
  --set auth.password="${PGVECTOR_PASSWORD}" \
  --set auth.database=dxp_vectors \
  --wait --timeout 3m

wait_pods $DXP_NAMESPACE_LLMOPS
ok "LiteLLM + pgvector installés"
