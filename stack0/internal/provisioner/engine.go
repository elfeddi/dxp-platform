package provisioner

import (
	"context"
	"fmt"
	"time"

	"github.com/elfeddi/dxp/internal/connector"
	"github.com/elfeddi/dxp/internal/resolver"
	"github.com/elfeddi/dxp/pkg/types"
)

type Engine struct {
	registry  *connector.Registry
	parser    *resolver.Parser
	validator *resolver.Validator
	dag       *resolver.DAGResolver
}

func NewEngine(registry *connector.Registry) *Engine {
	return &Engine{
		registry:  registry,
		parser:    resolver.NewParser(),
		validator: resolver.NewValidator(),
		dag:       resolver.NewDAGResolver(),
	}
}

func (e *Engine) Apply(ctx context.Context, path string) error {
	config, err := e.parser.ParseFile(path)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	if errs := e.validator.Validate(config); len(errs) > 0 {
		return fmt.Errorf("validation: %d erreur(s)", len(errs))
	}

	ordered, err := e.dag.Resolve(config)
	if err != nil {
		return fmt.Errorf("dag: %w", err)
	}

	jobs := make([]*ProvisionJob, 0, len(ordered))
	for _, c := range ordered {
		jobs = append(jobs, &ProvisionJob{
			ID:        fmt.Sprintf("%s-%d", c.Name, time.Now().UnixNano()),
			Connector: c,
			Status:    types.StatusPending,
		})
	}

	completed := []string{}
	for _, job := range jobs {
		if err := e.runJob(ctx, job); err != nil {
			job.Status = types.StatusFailed
			job.Error = err.Error()
			fmt.Printf("  ✗ [FAILED] %s: %s\n", job.Connector.Name, err)
			e.rollback(ctx, completed)
			return fmt.Errorf("provisioning échoué sur %s: %w", job.Connector.Name, err)
		}
		completed = append(completed, job.Connector.Name)
	}

	return nil
}

func (e *Engine) runJob(ctx context.Context, job *ProvisionJob) error {
	job.Status = types.StatusRunning

	conn, err := e.registry.Create(job.Connector.Type, job.Connector.Config)
	if err != nil {
		return fmt.Errorf("création connecteur: %w", err)
	}

	// Vérifier si déjà installé — idempotence
	ok, err := conn.HealthCheck(ctx)
	if err == nil && ok {
		fmt.Printf("  ✓ [ALREADY RUNNING] %s (%s)\n", conn.Name(), conn.Type())
		job.Status = types.StatusSucceeded
		return nil
	}

	// Pas encore installé — on installe
	fmt.Printf("  → [INSTALLING] %s (%s)...\n", conn.Name(), conn.Type())
	if err := conn.Install(ctx); err != nil {
		return fmt.Errorf("install: %w", err)
	}

	if err := conn.Configure(ctx); err != nil {
		return fmt.Errorf("configure: %w", err)
	}

	// Health check post-installation
	ok, err = conn.HealthCheck(ctx)
	if err != nil || !ok {
		return fmt.Errorf("health check échoué après installation: %v", err)
	}

	job.Status = types.StatusSucceeded
	fmt.Printf("  ✓ [SUCCEEDED] %s\n", conn.Name())
	return nil
}

func (e *Engine) rollback(ctx context.Context, completed []string) {
	for i := len(completed) - 1; i >= 0; i-- {
		fmt.Printf("  ↩ rollback %s...\n", completed[i])
	}
}
