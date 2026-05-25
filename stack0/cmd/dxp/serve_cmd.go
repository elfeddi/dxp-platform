package main

import "github.com/spf13/cobra"

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Démarre le serveur HTTP C4 Gateway",
	Long: `Démarre le serveur HTTP C4 Gateway.

C4 expose une API REST que Backstage proxifie via /api/dxp.
Auth : Authorization: Bearer <role>
Rôles : admin | operator | viewer | auditor

Exemples :
  dxp serve
  dxp serve --config dxp.yaml --addr :8090
  curl -H "Authorization: Bearer viewer" http://localhost:8090/healthz
  curl -H "Authorization: Bearer viewer" http://localhost:8090/api/dxp/backends`,
	RunE: runServe,
}

func init() {
	serveCmd.Flags().String("config", "dxp.yaml", "Chemin vers dxp.yaml")
	serveCmd.Flags().String("addr", ":8090", "Adresse d'écoute du serveur")
}
