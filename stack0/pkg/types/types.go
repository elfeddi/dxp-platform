package types

// DxPConfig représente le contenu de dxp.yaml
type DxPConfig struct {
	Version  string            `yaml:"version"`
	Name     string            `yaml:"name"`
	Provider string            `yaml:"provider"`
	Stacks   []Stack           `yaml:"stacks"`
	Metadata map[string]string `yaml:"metadata,omitempty"`
}

// Stack représente une discipline (DevOps, DataOps, MLOps...)
type Stack struct {
	Name       string      `yaml:"name"`
	Enabled    bool        `yaml:"enabled"`
	Connectors []Connector `yaml:"connectors"`
}

// Connector représente un composant à provisionner
type Connector struct {
	Type        string            `yaml:"type"`
	Name        string            `yaml:"name"`
	Namespace   string            `yaml:"namespace,omitempty"`
	Version     string            `yaml:"version,omitempty"`
	Config      map[string]string `yaml:"config,omitempty"`
	DependsOn   []string          `yaml:"dependsOn,omitempty"`
}

// ProvisionStatus représente l'état d'un provisioning
type ProvisionStatus string

const (
	StatusPending    ProvisionStatus = "pending"
	StatusRunning    ProvisionStatus = "running"
	StatusSucceeded  ProvisionStatus = "succeeded"
	StatusFailed     ProvisionStatus = "failed"
	StatusRollingBack ProvisionStatus = "rolling_back"
)

// ProvisionJob représente un job de provisioning
type ProvisionJob struct {
	ID         string            `yaml:"id"`
	StackName  string            `yaml:"stackName"`
	Connector  Connector         `yaml:"connector"`
	Status     ProvisionStatus   `yaml:"status"`
	Error      string            `yaml:"error,omitempty"`
}
