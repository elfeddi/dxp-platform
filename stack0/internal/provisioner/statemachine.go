package provisioner

import (
	"fmt"

	"github.com/elfeddi/dxp/pkg/types"
)

// ProvisionJob représente un job de provisioning avec sa state machine
type ProvisionJob struct {
	ID        string
	Connector types.Connector
	Status    types.ProvisionStatus
	Error     string
}

// Transitions valides entre états
var validTransitions = map[types.ProvisionStatus][]types.ProvisionStatus{
	types.StatusPending:     {types.StatusRunning},
	types.StatusRunning:     {types.StatusSucceeded, types.StatusFailed},
	types.StatusFailed:      {types.StatusRollingBack, types.StatusPending},
	types.StatusRollingBack: {types.StatusPending, types.StatusFailed},
	types.StatusSucceeded:   {}, // état terminal
}

// Transition effectue une transition d'état
func (j *ProvisionJob) Transition(next types.ProvisionStatus) error {
	allowed, ok := validTransitions[j.Status]
	if !ok {
		return fmt.Errorf("état inconnu: %s", j.Status)
	}

	for _, s := range allowed {
		if s == next {
			j.Status = next
			return nil
		}
	}

	return fmt.Errorf("transition invalide: %s → %s", j.Status, next)
}

// IsTerminal retourne true si l'état est terminal
func (j *ProvisionJob) IsTerminal() bool {
	return j.Status == types.StatusSucceeded
}

// IsFailed retourne true si le job est en échec
func (j *ProvisionJob) IsFailed() bool {
	return j.Status == types.StatusFailed
}

// String retourne une représentation lisible du job
func (j *ProvisionJob) String() string {
	return fmt.Sprintf("Job[%s] %s (%s) → %s",
		j.ID[:8], j.Connector.Name, j.Connector.Type, j.Status)
}
