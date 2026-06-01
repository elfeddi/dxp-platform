# ── Stage 1 : build ──────────────────────────────────────────────
FROM golang:1.26-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN GOWORK=off go mod download
COPY . .
RUN GOWORK=off CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o dxp ./cmd/dxp/
# ── Stage 2 : image finale ───────────────────────────────────────
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/dxp /dxp
EXPOSE 8090
ENTRYPOINT ["/dxp", "serve", "--addr", ":8090"]
