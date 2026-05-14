package dto

type UserPermission struct {
	ResourceName string `json:"resourceName"`
	ResourceURL  string `json:"resourceUrl"`
	R            bool   `json:"r"`
	W            bool   `json:"w"`
	U            bool   `json:"u"`
	D            bool   `json:"d"`
}

// EffectivePermission is one row of the v2 permissions API. ProjectID is
// nil for global grants. The frontend `can(action, resource, projectId)`
// hook matches on these tuples.
type EffectivePermission struct {
	ProjectID    *uint  `json:"projectId"`
	ResourceType string `json:"resourceType"`
	Action       string `json:"action"`
}
