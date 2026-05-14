-- name: ListUsers :many
SELECT id, worker_id, username, password, role_id
FROM users
ORDER BY id DESC;

-- name: ListUsersPaginated :many
SELECT id, worker_id, username, password, role_id
FROM users
ORDER BY id DESC
LIMIT $1 OFFSET $2;

-- name: ListUsersPaginatedFiltered :many
SELECT id, worker_id, username, password, role_id
FROM users
WHERE
    (NULLIF($1::bigint, 0) IS NULL OR worker_id = $1)
    AND (NULLIF($2::text, '') IS NULL OR username = $2)
ORDER BY id DESC
LIMIT $3 OFFSET $4;

-- name: GetUser :one
SELECT id, worker_id, username, password, role_id
FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT id, worker_id, username, password, role_id
FROM users
WHERE username = $1
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (worker_id, username, password, role_id)
VALUES ($1, $2, $3, $4)
RETURNING id, worker_id, username, password, role_id;

-- name: UpdateUser :exec
UPDATE users
SET worker_id = $2,
    username = $3,
    password = COALESCE(NULLIF($4::text, ''), password),
    role_id = $5
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: CountUsers :one
SELECT COUNT(*)::bigint FROM users;
