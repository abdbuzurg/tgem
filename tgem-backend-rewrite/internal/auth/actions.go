// Package auth holds the constants and resolver for the v2 permission model.
// See docs/permissions-spec.md for the full taxonomy. The string values here
// are the source of truth — they must match the rows seeded by migration
// 00005_permissions_v2_foundation.sql.
package auth

// Action is a verb performed on a ResourceType.
type Action string

const (
	ActionView    Action = "view"
	ActionCreate  Action = "create"
	ActionEdit    Action = "edit"
	ActionDelete  Action = "delete"
	ActionConfirm Action = "confirm"
	ActionCorrect Action = "correct"
	ActionImport  Action = "import"
	ActionExport  Action = "export"
	ActionReport  Action = "report"
)

// AllActions enumerates every action. Update when an action is added.
var AllActions = []Action{
	ActionView, ActionCreate, ActionEdit, ActionDelete,
	ActionConfirm, ActionCorrect, ActionImport, ActionExport, ActionReport,
}
