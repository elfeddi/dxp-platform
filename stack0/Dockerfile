# ── Stage 1 : build ──────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Dépendances d'abord (cache Docker)
COPY go.mod go.work* ./
RUN go mod download

# Sources
COPY . .

# Build statique — pas de CGO, binaire autonome
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o dxp ./cmd/dxp/

# ── Stage 2 : image finale ───────────────────────────────────────
FROM scratch

# Certificats TLS pour les appels HTTPS (ArgoCD Ingress)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Binaire uniquement
COPY --from=builder /build/dxp /dxp

EXPOSE 8090

ENTRYPOINT ["/dxp", "serve", "--addr", ":8090"]
