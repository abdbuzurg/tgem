package characterization_test

import (
	"context"
	"testing"

	"backend-v2/internal/auth"
	"backend-v2/internal/db"
	"backend-v2/test/characterization/helpers"
)

// TestPermissionsV2_Resolver_Superadmin verifies the phase-1 seed data feeds
// the phase-2 resolver correctly. Superadmin has 9 actions × 42 resources
// granted globally, so every (action, resource) tuple is allowed in every
// project (including project 0, which represents non-project-scoped routes
// like admin / auctions).
func TestPermissionsV2_Resolver_Superadmin(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}

	q := db.New(helpers.Pool())
	resolver := auth.NewResolver(q)

	// SeedTesterUser inserts user "tester" with role Суперадмин — id is
	// stable across runs but discover it via the queries handle.
	type userRow struct {
		ID int64
	}
	rows, err := helpers.Pool().Query(context.Background(),
		`SELECT id FROM users WHERE username = $1`, helpers.TesterUsername)
	if err != nil {
		t.Fatalf("locate tester user: %v", err)
	}
	defer rows.Close()
	var userID uint
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("scan user id: %v", err)
		}
		userID = uint(id)
	}
	if userID == 0 {
		t.Fatalf("tester user not seeded")
	}

	cases := []struct {
		desc     string
		project  uint
		resource auth.ResourceType
		action   auth.Action
		want     bool
	}{
		{"output view in project 1",     1, auth.ResInvoiceOutput,    auth.ActionView,    true},
		{"output confirm in project 1",  1, auth.ResInvoiceOutput,    auth.ActionConfirm, true},
		{"correction correct in project 2", 2, auth.ResInvoiceCorrection, auth.ActionCorrect, true},
		{"admin user view (no project)", 0, auth.ResAdminUser,        auth.ActionView,    true},
		{"auction bid create (no project)", 0, auth.ResAuctionBidPublic, auth.ActionCreate, true},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := resolver.Allowed(context.Background(), userID, tc.project, tc.resource, tc.action)
			if err != nil {
				t.Fatalf("resolver: %v", err)
			}
			if got != tc.want {
				t.Errorf("Allowed(%s, project=%d, %s, %s) = %v, want %v",
					tc.desc, tc.project, tc.resource, tc.action, got, tc.want)
			}
		})
	}
}

// TestPermissionsV2_Resolver_NoPermissions confirms that a user without any
// user_roles entry is denied everything — phase-2 baseline.
func TestPermissionsV2_Resolver_NoPermissions(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}

	q := db.New(helpers.Pool())
	resolver := auth.NewResolver(q)

	allowed, err := resolver.Allowed(context.Background(),
		/* userID = */ 999999, /* projectID = */ 1,
		auth.ResInvoiceOutput, auth.ActionView)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	if allowed {
		t.Errorf("non-existent user must not be allowed; got allowed=true")
	}
}
