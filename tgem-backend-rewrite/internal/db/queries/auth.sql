-- name: ListEffectivePermissionsForUser :many
-- Permissions v2 resolver query. Returns one row per (project_id, resource,
-- action) the user is granted via any role assignment. project_id IS NULL
-- means a global grant; non-null project_id is a per-project grant. The
-- resolver merges these against the request's project_id.
SELECT
    ur.project_id,
    rg.resource_type_code,
    rg.action_code
FROM user_roles ur
INNER JOIN role_grants rg ON rg.role_id = ur.role_id
WHERE ur.user_id = $1;
