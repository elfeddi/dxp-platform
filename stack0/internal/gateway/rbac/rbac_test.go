package rbac

import (
	"testing"
)

func TestCheck_AdminAllOps(t *testing.T) {
	c := NewChecker()
	ops := []Op{OpGetStatus, OpGetMetrics, OpGetLogs, OpListActions}
	for _, op := range ops {
		if err := c.Check(RoleAdmin, op, false); err != nil {
			t.Errorf("admin devrait pouvoir %q: %v", op, err)
		}
	}
	// admin peut exécuter des actions dangereuses
	if err := c.Check(RoleAdmin, OpExecuteAction, true); err != nil {
		t.Errorf("admin devrait pouvoir exécuter des actions dangereuses: %v", err)
	}
}

func TestCheck_OperatorNoDanger(t *testing.T) {
	c := NewChecker()
	// operator peut exécuter des actions non-dangereuses
	if err := c.Check(RoleOperator, OpExecuteAction, false); err != nil {
		t.Errorf("operator devrait pouvoir exécuter des actions: %v", err)
	}
	// operator ne peut pas exécuter des actions dangereuses
	if err := c.Check(RoleOperator, OpExecuteAction, true); err == nil {
		t.Error("operator ne devrait pas pouvoir exécuter des actions dangereuses")
	}
}

func TestCheck_ViewerReadOnly(t *testing.T) {
	c := NewChecker()
	// viewer peut lire status et metrics
	if err := c.Check(RoleViewer, OpGetStatus, false); err != nil {
		t.Errorf("viewer devrait pouvoir lire status: %v", err)
	}
	if err := c.Check(RoleViewer, OpGetMetrics, false); err != nil {
		t.Errorf("viewer devrait pouvoir lire metrics: %v", err)
	}
	// viewer ne peut pas lire les logs
	if err := c.Check(RoleViewer, OpGetLogs, false); err == nil {
		t.Error("viewer ne devrait pas pouvoir lire les logs")
	}
	// viewer ne peut pas exécuter d'actions
	if err := c.Check(RoleViewer, OpExecuteAction, false); err == nil {
		t.Error("viewer ne devrait pas pouvoir exécuter des actions")
	}
}

func TestCheck_AuditorLogs(t *testing.T) {
	c := NewChecker()
	// auditor peut lire les logs
	if err := c.Check(RoleAuditor, OpGetLogs, false); err != nil {
		t.Errorf("auditor devrait pouvoir lire les logs: %v", err)
	}
	// auditor ne peut pas exécuter d'actions
	if err := c.Check(RoleAuditor, OpExecuteAction, false); err == nil {
		t.Error("auditor ne devrait pas pouvoir exécuter des actions")
	}
}

func TestCheck_UnknownRole(t *testing.T) {
	c := NewChecker()
	if err := c.Check(Role("hacker"), OpGetStatus, false); err == nil {
		t.Error("Erreur attendue pour rôle inconnu")
	}
}

func TestCheck_UnknownOp(t *testing.T) {
	c := NewChecker()
	if err := c.Check(RoleAdmin, Op("unknown_op"), false); err == nil {
		t.Error("Erreur attendue pour opération inconnue")
	}
}

func TestRoleFromString_Valid(t *testing.T) {
	cases := []struct {
		input    string
		expected Role
	}{
		{"admin", RoleAdmin},
		{"operator", RoleOperator},
		{"viewer", RoleViewer},
		{"auditor", RoleAuditor},
	}
	for _, tc := range cases {
		r, err := RoleFromString(tc.input)
		if err != nil {
			t.Errorf("RoleFromString(%q) erreur inattendue: %v", tc.input, err)
		}
		if r != tc.expected {
			t.Errorf("RoleFromString(%q) = %q, attendu %q", tc.input, r, tc.expected)
		}
	}
}

func TestRoleFromString_Invalid(t *testing.T) {
	_, err := RoleFromString("superadmin")
	if err == nil {
		t.Error("Erreur attendue pour rôle invalide")
	}
}
