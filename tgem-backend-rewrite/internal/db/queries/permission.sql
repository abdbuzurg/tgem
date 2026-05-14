-- name: ListPermissions :many
SELECT id, role_id, resource_id, r, w, u, d
FROM permissions;

-- name: ListPermissionsByRoleID :many
SELECT id, role_id, resource_id, r, w, u, d
FROM permissions
WHERE role_id = $1 AND (r OR w OR u OR d);

-- name: GetPermissionByResourceURL :one
SELECT permissions.id, permissions.role_id, permissions.resource_id,
       permissions.r, permissions.w, permissions.u, permissions.d
FROM permissions
INNER JOIN roles ON roles.id = permissions.role_id
INNER JOIN resources ON resources.id = permissions.resource_id
WHERE permissions.role_id = $1 AND resources.url = $2;

-- name: ListUserPermissionsByRoleName :many
SELECT
    resources.name AS resource_name,
    resources.url  AS resource_url,
    permissions.r,
    permissions.w,
    permissions.u,
    permissions.d
FROM permissions
INNER JOIN roles ON roles.id = permissions.role_id
INNER JOIN resources ON resources.id = permissions.resource_id
WHERE roles.name = $1;

-- name: CreatePermission :one
INSERT INTO permissions (role_id, resource_id, r, w, u, d)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, role_id, resource_id, r, w, u, d;

-- name: CreatePermissionsBatch :copyfrom
INSERT INTO permissions (role_id, resource_id, r, w, u, d)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: UpdatePermission :one
UPDATE permissions
SET role_id = $2, resource_id = $3, r = $4, w = $5, u = $6, d = $7
WHERE id = $1
RETURNING id, role_id, resource_id, r, w, u, d;

-- name: DeletePermission :exec
DELETE FROM permissions WHERE id = $1;
