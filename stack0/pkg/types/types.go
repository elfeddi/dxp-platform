package types

type DxPConfig struct {
	Version  string            `yaml:"version"`
	Name     string            `yaml:"name"`
	Provider string            `yaml:"provider"`
	Stacks   []Stack           `yaml:"stacks"`
	Metadata map[string]string `yaml:"metadata,omitempty"`
}

type Stack struct {
	Name       string      `yaml:"name"`
	Enabled    bool        `yaml:"enabled"`
	Connectors []Connector `yaml:"connectors"`
}

type Connector struct {
	Type      string            `yaml:"type"`
	Name      string            `yaml:"name"`
	Namespace string            `yaml:"namespace,omitempty"`
	Version   string            `yaml:"version,omitempty"`
	Config    map[string]string `yaml:"config,omitempty"`
	DependsOn []string          `yaml:"dependsOn,omitempty"`
}

type ProvisionStatus string

const (
	StatusPending     ProvisionStatus = "pending"
	StatusRunning     ProvisionStatus = "running"
	StatusSucceeded   ProvisionStatus = "succeeded"
	StatusFailed      ProvisionStatus = "failed"
	StatusRollingBack ProvisionStatus = "rolling_back"
)

type ProvisionJob struct {
	ID        string          `yaml:"id"`
	StackName string          `yaml:"stackName"`
	Connector Connector       `yaml:"connector"`
	Status    ProvisionStatus `yaml:"status"`
	Error     string          `yaml:"error,omitempty"`
}
