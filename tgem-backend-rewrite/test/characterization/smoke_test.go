package characterization_test

import (
	"backend-v2/pkg/jwt"
	"backend-v2/test/characterization/helpers"
	"testing"
)

// TestSmoke_BootsAndLogsIn validates the harness end-to-end: the test database
// was created, AutoMigrate + seeds + the tester user fixture all succeeded,
// and the Gin router is reachable.
func TestSmoke_BootsAndLogsIn(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}

	token := helpers.LoginAsTester(t)
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	if _, err := jwt.VerifyToken(token); err != nil {
		t.Fatalf("VerifyToken: %v", err)
	}
}
