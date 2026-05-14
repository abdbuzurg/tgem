-- The `code` column was added in 00005_permissions_v2_foundation.sql. Selecting
-- it here keeps sqlc emitting db.Role uniformly; the legacy usecase ignores it.

-- name: ListRoles :many
SELECT id, name, description, code
FROM roles;

-- name: GetRole :one
SELECT id, name, description, code
FROM roles
WHERE id = $1;

-- name: CreateRole :one
INSERT INTO roles (name, description)
VALUES ($1, $2)
RETURNING id, name, description, code;

-- name: UpdateRole :one
UPDATE roles
SET name = $2, description = $3
WHERE id = $1
RETURNING id, name, description, code;

-- name: DeleteRole :exec
DELETE FROM roles WHERE id = $1;
