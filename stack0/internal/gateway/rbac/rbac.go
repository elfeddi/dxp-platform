package rbac

import "fmt"

// Role représente un rôle DxP.
type Role string

const (
	RoleAdmin    Role = "admin"    // accès complet + actions dangereuses
	RoleOperator Role = "operator" // lecture + actions non-dangereuses
	RoleViewer   Role = "viewer"   // lecture seule
	RoleAuditor  Role = "auditor"  // lecture seule + logs
)

// Permission décrit ce qu'un rôle peut faire.
type Permission struct {
	Read           bool
	ReadLogs       bool
	ExecuteActions bool
	ExecuteDanger  bool
	ListActions    bool
}

var permissions = map[Role]Permission{
	RoleAdmin:    {Read: true, ReadLogs: true, ExecuteActions: true, ExecuteDanger: true, ListActions: true},
	RoleOperator: {Read: true, ReadLogs: true, ExecuteActions: true, ExecuteDanger: false, ListActions: true},
	RoleViewer:   {Read: true, ReadLogs: false, ListActions: true},
	RoleAuditor:  {Read: true, ReadLogs: true, ListActions: true},
}

// Op représente une opération demandée sur un backend.
type Op string

const (
	OpGetStatus     Op = "get_status"
	OpGetMetrics    Op = "get_metrics"
	OpGetLogs       Op = "get_logs"
	OpExecuteAction Op = "execute_action"
	OpListActions   Op = "list_actions"
)

// Checker vérifie les autorisations.
type Checker struct{}

func NewChecker() *Checker { return &Checker{} }

// Check vérifie si le rôle est autorisé à effectuer op.
func (c *Checker) Check(role Role, op Op, dangerous bool) error {
	perm, ok := permissions[role]
	if !ok {
		return fmt.Errorf("rbac: unknown role %q", role)
	}
	switch op {
	case OpGetStatus, OpGetMetrics:
		if !perm.Read {
			return fmt.Errorf("rbac: role %q cannot read status/metrics", role)
		}
	case OpGetLogs:
		if !perm.ReadLogs {
			return fmt.Errorf("rbac: role %q cannot read logs", role)
		}
	case OpListActions:
		if !perm.ListActions {
			return fmt.Errorf("rbac: role %q cannot list actions", role)
		}
	case OpExecuteAction:
		if dangerous {
			if !perm.ExecuteDanger {
				return fmt.Errorf("rbac: role %q cannot execute dangerous actions", role)
			}
		} else {
			if !perm.ExecuteActions {
				return fmt.Errorf("rbac: role %q cannot execute actions", role)
			}
		}
	default:
		return fmt.Errorf("rbac: unknown operation %q", op)
	}
	return nil
}

// RoleFromString convertit une string en Role.
func RoleFromString(s string) (Role, error) {
	r := Role(s)
	if _, ok := permissions[r]; !ok {
		return "", fmt.Errorf("rbac: unknown role %q", s)
	}
	return r, nil
}
