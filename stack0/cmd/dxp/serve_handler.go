package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/elfeddi/dxp/internal/gateway"
	argocdbackend "github.com/elfeddi/dxp/internal/gateway/backends/argocd"
	grafanabackend "github.com/elfeddi/dxp/internal/gateway/backends/grafana"
	harborbackend "github.com/elfeddi/dxp/internal/gateway/backends/harbor"
	"github.com/elfeddi/dxp/internal/gateway/router"
	"github.com/elfeddi/dxp/internal/aigateway"
	"github.com/elfeddi/dxp/internal/observer"
	"github.com/elfeddi/dxp/internal/resolver"
	"github.com/elfeddi/dxp/pkg/types"
	"github.com/spf13/cobra"
)

func runServe(cmd *cobra.Command, args []string) error {
	configFile, _ := cmd.Flags().GetString("config")
	addr, _ := cmd.Flags().GetString("addr")

	cfg, err := resolver.NewParser().ParseFile(configFile)
	if err != nil {
		return fmt.Errorf("C1 resolver: %w", err)
	}
	slog.Info("config loaded", "stacks", len(cfg.Stacks))

	reg := gateway.NewRegistry()
	if err := buildRegistry(reg, cfg); err != nil {
		return fmt.Errorf("C4 registry: %w", err)
	}
	slog.Info("backends registered", "count", len(reg.List()), "backends", reg.List())

	srv := &http.Server{
		Addr:  addr,
		Handler: func() http.Handler {
			obs := observer.New(reg)
			obs.Start(context.Background())
			ai := aigateway.New(obs.Context)
			return router.New(reg, obs, ai).Handler()
		}(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 300 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("C4 Gateway listening", "addr", addr)
		fmt.Printf("\n  ✓ C4 Gateway running on http://localhost%s\n", addr)
		fmt.Printf("  ✓ Auth: Authorization: Bearer <role>\n")
		fmt.Printf("  ✓ Roles: admin | operator | viewer | auditor\n\n")
		fmt.Printf("  Routes:\n")
		fmt.Printf("    GET  /healthz\n")
		fmt.Printf("    GET  /api/dxp/backends\n")
		fmt.Printf("    GET  /api/dxp/backends/{name}/status\n")
		fmt.Printf("    GET  /api/dxp/backends/{name}/metrics\n")
		fmt.Printf("    GET  /api/dxp/backends/{name}/logs\n")
		fmt.Printf("    GET  /api/dxp/backends/{name}/actions\n")
		fmt.Printf("    POST /api/dxp/backends/{name}/actions\n\n")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		slog.Info("shutdown signal received", "signal", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}
	slog.Info("C4 Gateway stopped cleanly")
	return nil
}

func buildRegistry(reg *gateway.Registry, cfg *types.DxPConfig) error {
	for _, stack := range cfg.Stacks {
		if !stack.Enabled {
			continue
		}
		for _, conn := range stack.Connectors {
			var b gateway.BackendConnector
			switch conn.Type {
			case "argocd":
				url := getConfigStr(conn.Config, "url", "https://localhost:9443/argocd")
				token := getConfigStr(conn.Config, "token", "")
				if token == "" {
					token = os.Getenv("ARGOCD_TOKEN")
				}
				b = argocdbackend.New(url, token)
			case "harbor":
				url := getConfigStr(conn.Config, "url", "http://localhost:9091")
				token := getConfigStr(conn.Config, "token", "")
				if token == "" {
					token = os.Getenv("HARBOR_API_TOKEN")
				}
				b = harborbackend.New(url, token)
			case "grafana":
				url := getConfigStr(conn.Config, "url", "http://localhost:3001")
				token := getConfigStr(conn.Config, "token", "")
				if token == "" {
					token = os.Getenv("GRAFANA_API_TOKEN")
				}
				b = grafanabackend.New(url, token)
			// case "tekton":  b = tektonbackend.New(...)
			default:
				slog.Warn("no C4 backend for connector type — skipping",
					"type", conn.Type, "name", conn.Name)
				continue
			}
			if err := reg.Register(b); err != nil {
				slog.Warn("backend already registered — skipping", "name", b.Name())
			}
		}
	}
	return nil
}

// Config est map[string]string — pas de cast nécessaire
func getConfigStr(config map[string]string, key, defaultVal string) string {
	if v, ok := config[key]; ok && v != "" {
		return v
	}
	return defaultVal
}
