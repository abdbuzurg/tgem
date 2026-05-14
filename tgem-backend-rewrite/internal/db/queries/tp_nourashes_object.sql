-- name: ListTPNourashesObjectNamesByTarget :many
SELECT COALESCE(objects.name, '')::text AS name
FROM tp_nourashes_objects
INNER JOIN objects ON objects.id = tp_nourashes_objects.tp_object_id
WHERE
    tp_nourashes_objects.target_id = $1
    AND tp_nourashes_objects.target_type = $2;

-- name: DeleteTPNourashesObjectsByTarget :exec
DELETE FROM tp_nourashes_objects
WHERE target_id = $1 AND target_type = $2;

-- name: CreateTPNourashesObjectsBatch :copyfrom
INSERT INTO tp_nourashes_objects (tp_object_id, target_id, target_type)
VALUES ($1, $2, $3);
