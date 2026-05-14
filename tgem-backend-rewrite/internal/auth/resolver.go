package auth

import (
	"context"
	"fmt"

	"backend-v2/internal/db"
)

// Resolver answers "is user U allowed to do action A on resource R within
// project P?" against the v2 permission model.
//
// Project semantics:
//   - A grant with project_id = NULL applies in every project (global, used
//     for system roles like superadmin).
//   - A grant with a specific project_id applies only when the request
//     targets that project.
//   - Routes that are not project-scoped (e.g. auctions, admin) pass
//     projectID = 0; only global grants will match.
type Resolver interface {
	Allowed(ctx context.Context, userID uint, projectID uint, resource ResourceType, action Action) (bool, error)
}

// dbResolver is the production resolver. No caching for now — phase 2 ships
// with a per-request query so we can observe load before adding TTL caching
// in phase 4 (when enforcement turns on).
type dbResolver struct {
	q *db.Queries
}

// NewResolver constructs a Resolver backed by the given sqlc queries handle.
func NewResolver(q *db.Queries) Resolver {
	return &dbResolver{q: q}
}

func (r *dbResolver) Allowed(ctx context.Context, userID uint, projectID uint, resource ResourceType, action Action) (bool, error) {
	if userID == 0 {
		return false, nil
	}
	rows, err := r.q.ListEffectivePermissionsForUser(ctx, int64(userID))
	if err != nil {
		return false, fmt.Errorf("auth.Allowed: list permissions for user %d: %w", userID, err)
	}
	for _, row := range rows {
		if row.ResourceTypeCode != string(resource) || row.ActionCode != string(action) {
			continue
		}
		// project_id IS NULL → global grant, allow regardless of request project.
		if !row.ProjectID.Valid {
			return true, nil
		}
		// project_id matches request → allow.
		if uint(row.ProjectID.Int64) == projectID {
			return true, nil
		}
	}
	return false, nil
}
