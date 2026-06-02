#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "21 — LiteLLM (API externe)"

kubectl create namespace llmops --dry-run=client -o yaml | kubectl apply -f -

kubectl apply -f - << LITELLMEOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: litellm
  namespace: llmops
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
      nodeSelector:
        role: infra
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
  namespace: llmops
spec:
  type: NodePort
  selector:
    app: litellm
  ports:
  - port: 4000
    targetPort: 4000
    nodePort: 30096
LITELLMEOF

wait_pods llmops
ok "LiteLLM installé — NodePort 30096 — Master"
