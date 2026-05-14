-- name: GetKL04KVObjectDetail :one
SELECT id, length, nourashes
FROM kl04_kv_objects
WHERE id = $1;

-- name: ListKL04KVObjectsPaginated :many
SELECT DISTINCT
    objects.id                                   AS object_id,
    kl04_kv_objects.id                           AS object_detailed_id,
    COALESCE(objects.name, '')::text             AS name,
    COALESCE(objects.status, '')::text           AS status,
    kl04_kv_objects.length                       AS length,
    COALESCE(kl04_kv_objects.nourashes, '')::text AS nourashes
FROM objects
INNER JOIN kl04_kv_objects ON kl04_kv_objects.id = objects.object_detailed_id
FULL JOIN object_teams ON object_teams.object_id = objects.id
FULL JOIN object_supervisors ON object_supervisors.object_id = objects.id
FULL JOIN tp_nourashes_objects ON tp_nourashes_objects.target_id = objects.id
WHERE
    objects.type = 'kl04kv_objects'
    AND objects.project_id = $1
    AND (NULLIF($2::text, '') IS NULL OR objects.name = $2)
    AND (NULLIF($3::bigint, 0) IS NULL OR object_teams.team_id = $3)
    AND (NULLIF($4::bigint, 0) IS NULL OR object_supervisors.supervisor_worker_id = $4)
    AND (NULLIF($5::bigint, 0) IS NULL OR tp_nourashes_objects.tp_object_id = $5)
ORDER BY kl04_kv_objects.id DESC
LIMIT $6 OFFSET $7;

-- name: CountKL04KVObjectsFiltered :one
SELECT COUNT(*)::bigint
FROM (
    SELECT DISTINCT objects.id
    FROM objects
    INNER JOIN kl04_kv_objects ON kl04_kv_objects.id = objects.object_detailed_id
    FULL JOIN object_teams ON object_teams.object_id = objects.id
    FULL JOIN object_supervisors ON object_supervisors.object_id = objects.id
    FULL JOIN tp_nourashes_objects ON tp_nourashes_objects.target_id = objects.id
    WHERE
        objects.type = 'kl04kv_objects'
        AND objects.project_id = $1
        AND (NULLIF($2::text, '') IS NULL OR objects.name = $2)
        AND (NULLIF($3::bigint, 0) IS NULL OR object_teams.team_id = $3)
        AND (NULLIF($4::bigint, 0) IS NULL OR object_supervisors.supervisor_worker_id = $4)
        AND (NULLIF($5::bigint, 0) IS NULL OR tp_nourashes_objects.tp_object_id = $5)
) sub;

-- name: CreateKL04KVObjectDetail :one
INSERT INTO kl04_kv_objects (length, nourashes)
VALUES ($1, $2)
RETURNING id, length, nourashes;

-- name: UpdateKL04KVObjectDetail :exec
UPDATE kl04_kv_objects
SET length = $2, nourashes = $3
WHERE id = $1;

-- name: DeleteKL04KVObjectSupervisorsCascade :exec
DELETE FROM object_supervisors
WHERE object_supervisors.object_id = (
    SELECT DISTINCT objects.id
    FROM objects
    INNER JOIN kl04_kv_objects ON kl04_kv_objects.id = objects.object_detailed_id
    WHERE
        objects.project_id = $1
        AND kl04_kv_objects.id = $2
        AND objects.type = 'kl04kv_objects'
);

-- name: DeleteKL04KVObjectTeamsCascade :exec
DELETE FROM object_teams
WHERE object_teams.object_id = (
    SELECT DISTINCT objects.id
    FROM objects
    INNER JOIN kl04_kv_objects ON kl04_kv_objects.id = objects.object_detailed_id
    WHERE
        objects.project_id = $1
        AND kl04_kv_objects.id = $2
        AND objects.type = 'kl04kv_objects'
);

-- name: DeleteKL04KVObjectTPNourashesCascade :exec
DELETE FROM tp_nourashes_objects
WHERE tp_nourashes_objects.target_id = (
    SELECT DISTINCT objects.id
    FROM objects
    INNER JOIN kl04_kv_objects ON kl04_kv_objects.id = objects.object_detailed_id
    WHERE
        objects.project_id = $1
        AND kl04_kv_objects.id = $2
        AND objects.type = 'kl04kv_objects'
);

-- name: DeleteKL04KVObjectDetail :exec
DELETE FROM kl04_kv_objects WHERE id = $1;

-- name: DeleteObjectByKL04KVDetailedID :exec
DELETE FROM objects
WHERE object_detailed_id = $1 AND type = 'kl04kv_objects';

-- name: ListKL04KVObjectNamesForSearch :many
SELECT
    COALESCE(objects.name, '')::text AS label,
    COALESCE(objects.name, '')::text AS value
FROM objects
INNER JOIN kl04_kv_objects ON kl04_kv_objects.id = objects.object_detailed_id
WHERE
    objects.project_id = $1
    AND objects.type = 'kl04kv_objects';
