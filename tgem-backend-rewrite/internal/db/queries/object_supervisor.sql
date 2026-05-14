-- name: ListObjectSupervisorsByObjectID :many
SELECT id, supervisor_worker_id, object_id
FROM object_supervisors
WHERE object_id = $1;

-- name: ListObjectSupervisorsByWorkerID :many
SELECT id, supervisor_worker_id, object_id
FROM object_supervisors
WHERE supervisor_worker_id = $1;

-- name: CreateObjectSupervisorsBatch :copyfrom
INSERT INTO object_supervisors (supervisor_worker_id, object_id) VALUES ($1, $2);

-- name: ListSupervisorAndObjectNamesByObjectID :many
SELECT
    COALESCE(objects.name, '')::text     AS object_name,
    COALESCE(objects.type, '')::text     AS object_type,
    COALESCE(workers.name, '')::text     AS supervisor_name
FROM object_supervisors
RIGHT JOIN objects ON objects.id = object_supervisors.object_id
LEFT JOIN workers ON workers.id = object_supervisors.supervisor_worker_id
WHERE objects.project_id = $1 AND objects.id = $2;

-- name: ListSupervisorNamesByObjectID :many
SELECT COALESCE(workers.name, '')::text AS supervisor_name
FROM object_supervisors
INNER JOIN objects ON objects.id = object_supervisors.object_id
INNER JOIN workers ON workers.id = object_supervisors.supervisor_worker_id
WHERE objects.id = $1;

-- name: DeleteObjectSupervisorsByObjectID :exec
DELETE FROM object_supervisors WHERE object_id = $1;
