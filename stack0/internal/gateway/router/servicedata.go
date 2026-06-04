package router

// ServiceOverview agrège toutes les données d'un service pour Backstage.
type ServiceOverview struct {
	Name      string        `json:"name"`
	Namespace string        `json:"namespace"`
	Pipeline  PipelineInfo  `json:"pipeline"`
	Deploy    DeployInfo    `json:"deploy"`
	Pods      []PodInfo     `json:"pods"`
}

// PipelineInfo contient les derniers PipelineRuns Tekton.
type PipelineInfo struct {
	Runs  []PipelineRun `json:"runs"`
	Total int           `json:"total"`
}

// PipelineRun représente un PipelineRun Tekton.
type PipelineRun struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	StartTime string `json:"start_time"`
	Duration  string `json:"duration"`
}

// DeployInfo contient le statut ArgoCD.
type DeployInfo struct {
	AppName    string `json:"app_name"`
	SyncStatus string `json:"sync_status"`
	Health     string `json:"health"`
	Revision   string `json:"revision"`
	Message    string `json:"message,omitempty"`
}

// PodInfo représente un pod K8s.
type PodInfo struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Ready    string `json:"ready"`
	Restarts int32  `json:"restarts"`
	Age      string `json:"age"`
}
