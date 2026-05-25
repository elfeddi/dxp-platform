package gateway

import "context"

// BackendConnector est le contrat que chaque backend DxP doit implémenter.
// Ajouter un composant = écrire un connector qui implémente cette interface
// + 3 lignes dans dxp.yaml. C4 n'a pas besoin d'être modifié.
type BackendConnector interface {
	// Name retourne l'identifiant unique du backend (ex: "argocd", "harbor").
	Name() string

	// GetStatus retourne l'état courant du composant.
	GetStatus(ctx context.Context) (*ComponentStatus, error)

	// GetMetrics retourne les métriques Prometheus du composant (format texte).
	GetMetrics(ctx context.Context) (string, error)

	// GetLogs retourne les N dernières lignes de logs du composant.
	GetLogs(ctx context.Context, req *LogRequest) ([]*LogEntry, error)

	// ExecuteAction exécute une action sur le composant (sync, restart, etc.).
	// Les actions disponibles sont déclarées par le backend via ListActions().
	ExecuteAction(ctx context.Context, req *ActionRequest) (*ActionResult, error)

	// ListActions retourne la liste des actions supportées par ce backend.
	ListActions(ctx context.Context) ([]*ActionDescriptor, error)
}

// ComponentStatus représente l'état d'un composant backend.
type ComponentStatus struct {
	Name      string            `json:"name"`
	Healthy   bool              `json:"healthy"`
	Ready     bool              `json:"ready"`
	Version   string            `json:"version,omitempty"`
	Message   string            `json:"message,omitempty"`
	Details   map[string]string `json:"details,omitempty"`
	UpdatedAt int64             `json:"updated_at"` // Unix timestamp
}

// LogRequest paramètre une requête de logs.
type LogRequest struct {
	Lines     int    `json:"lines"`     // Nombre de lignes (défaut : 100)
	Since     string `json:"since"`     // Durée relative ex: "5m", "1h"
	Filter    string `json:"filter"`    // Filtre texte optionnel
	Namespace string `json:"namespace"` // Namespace K8s optionnel
}

// LogEntry représente une ligne de log.
type LogEntry struct {
	Timestamp int64  `json:"timestamp"` // Unix ms
	Level     string `json:"level"`     // info, warn, error
	Message   string `json:"message"`
	Source    string `json:"source,omitempty"` // pod/container source
}

// ActionRequest décrit une action à exécuter sur un backend.
type ActionRequest struct {
	Action string            `json:"action"`            // ex: "sync", "restart", "rotate-secret"
	Target string            `json:"target,omitempty"`  // ressource cible optionnelle
	Params map[string]string `json:"params,omitempty"`  // paramètres additionnels
	DryRun bool              `json:"dry_run,omitempty"` // simulation sans effet réel
}

// ActionResult retourne le résultat d'une action exécutée.
type ActionResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output,omitempty"` // sortie détaillée si disponible
}

// ActionDescriptor décrit une action disponible sur un backend.
type ActionDescriptor struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Params      map[string]string `json:"params,omitempty"` // nom → description du param
	Dangerous   bool              `json:"dangerous"`        // action destructive ?
}
