-- name: ListDistricts :many
SELECT id, name, project_id
FROM districts
WHERE project_id = $1
ORDER BY id DESC;

-- name: ListDistrictsPaginated :many
SELECT id, name, project_id
FROM districts
WHERE project_id = $1
ORDER BY id DESC
LIMIT $2 OFFSET $3;

-- name: GetDistrict :one
SELECT id, name, project_id
FROM districts
WHERE id = $1;

-- name: GetDistrictByName :one
SELECT id, name, project_id
FROM districts
WHERE name = $1
LIMIT 1;

-- name: CreateDistrict :one
INSERT INTO districts (name, project_id)
VALUES ($1, $2)
RETURNING id, name, project_id;

-- name: UpdateDistrict :one
UPDATE districts
SET name = $2, project_id = $3
WHERE id = $1
RETURNING id, name, project_id;

-- name: DeleteDistrict :exec
DELETE FROM districts WHERE id = $1;

-- name: CountDistricts :one
SELECT COUNT(*) FROM districts WHERE project_id = $1;
