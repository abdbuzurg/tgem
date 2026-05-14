-- name: ListObjectsByProject :many
SELECT id, object_detailed_id, type, name, status, project_id
FROM objects
WHERE project_id = $1
ORDER BY id DESC;

-- name: ListObjectsByProjectAndType :many
SELECT id, object_detailed_id, type, name, status, project_id
FROM objects
WHERE project_id = $1 AND type = $2
ORDER BY id DESC;

-- name: ListObjectsPaginated :many
SELECT id, object_detailed_id, type, name, status, project_id
FROM objects
ORDER BY id DESC
LIMIT $1 OFFSET $2;

-- name: GetObject :one
SELECT id, object_detailed_id, type, name, status, project_id
FROM objects
WHERE id = $1;

-- name: GetObjectByName :one
SELECT id, object_detailed_id, type, name, status, project_id
FROM objects
WHERE name = $1
LIMIT 1;

-- name: ListObjectsByIDs :many
SELECT id, object_detailed_id, type, name, status, project_id
FROM objects
WHERE id = ANY(sqlc.arg(ids)::bigint[]);

-- name: ListObjectsPaginatedFiltered :many
SELECT
    objects.id                                   AS id,
    COALESCE(objects.type, '')::text             AS object_type,
    COALESCE(objects.name, '')::text             AS object_name,
    COALESCE(objects.status, '')::text           AS object_status,
    COALESCE(workers.name, '')::text             AS supervisor_name
FROM object_supervisors
INNER JOIN objects ON objects.id = object_supervisors.object_id
INNER JOIN workers ON workers.id = object_supervisors.supervisor_worker_id
WHERE
    (NULLIF($1::bigint, 0) IS NULL OR objects.object_detailed_id = $1)
    AND (NULLIF($2::text, '') IS NULL OR objects.type = $2)
    AND (NULLIF($3::text, '') IS NULL OR objects.name = $3)
    AND (NULLIF($4::text, '') IS NULL OR objects.status = $4)
ORDER BY objects.id DESC
LIMIT $5 OFFSET $6;

-- name: CreateObject :one
INSERT INTO objects (object_detailed_id, type, name, status, project_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, object_detailed_id, type, name, status, project_id;

-- name: UpdateObject :one
UPDATE objects
SET object_detailed_id = $2, type = $3, name = $4, status = $5, project_id = $6
WHERE id = $1
RETURNING id, object_detailed_id, type, name, status, project_id;

-- name: DeleteObject :exec
DELETE FROM objects WHERE id = $1;

-- name: CountObjects :one
SELECT COUNT(*)::bigint FROM objects;
