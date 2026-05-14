-- name: GetTPObjectDetail :one
SELECT id, model, voltage_class
FROM tp_objects
WHERE id = $1;

-- name: ListTPObjectsAll :many
SELECT
    objects.id                                   AS object_id,
    objects.object_detailed_id                   AS object_detailed_id,
    COALESCE(objects.name, '')::text             AS name,
    COALESCE(objects.status, '')::text           AS status,
    COALESCE(tp_objects.model, '')::text         AS model,
    COALESCE(tp_objects.voltage_class, '')::text AS voltage_class
FROM objects
INNER JOIN tp_objects ON objects.object_detailed_id = tp_objects.id
WHERE
    objects.type = 'tp_objects'
    AND objects.project_id = $1
ORDER BY tp_objects.id DESC;

-- name: ListTPObjectsPaginated :many
SELECT DISTINCT
    objects.id                                   AS object_id,
    tp_objects.id                                AS object_detailed_id,
    COALESCE(objects.name, '')::text             AS name,
    COALESCE(objects.status, '')::text           AS status,
    COALESCE(tp_objects.model, '')::text         AS model,
    COALESCE(tp_objects.voltage_class, '')::text AS voltage_class
FROM objects
INNER JOIN tp_objects ON objects.object_detailed_id = tp_objects.id
FULL JOIN object_teams ON object_teams.object_id = objects.id
FULL JOIN object_supervisors ON object_supervisors.object_id = objects.id
FULL JOIN tp_nourashes_objects ON tp_nourashes_objects.target_id = objects.id
WHERE
    objects.type = 'tp_objects'
    AND objects.project_id = $1
    AND (NULLIF($2::text, '') IS NULL OR objects.name = $2)
    AND (NULLIF($3::bigint, 0) IS NULL OR object_teams.team_id = $3)
    AND (NULLIF($4::bigint, 0) IS NULL OR object_supervisors.supervisor_worker_id = $4)
ORDER BY tp_objects.id DESC
LIMIT $5 OFFSET $6;

-- name: CountTPObjectsFiltered :one
SELECT COUNT(*)::bigint
FROM (
    SELECT DISTINCT objects.id
    FROM objects
    FULL JOIN object_teams ON object_teams.object_id = objects.id
    FULL JOIN object_supervisors ON object_supervisors.object_id = objects.id
    FULL JOIN tp_nourashes_objects ON tp_nourashes_objects.target_id = objects.id
    WHERE
        objects.type = 'tp_objects'
        AND objects.project_id = $1
        AND (NULLIF($2::text, '') IS NULL OR objects.name = $2)
        AND (NULLIF($3::bigint, 0) IS NULL OR object_teams.team_id = $3)
        AND (NULLIF($4::bigint, 0) IS NULL OR object_supervisors.supervisor_worker_id = $4)
) sub;

-- name: CreateTPObjectDetail :one
INSERT INTO tp_objects (model, voltage_class)
VALUES ($1, $2)
RETURNING id, model, voltage_class;

-- name: UpdateTPObjectDetail :exec
UPDATE tp_objects
SET model = $2, voltage_class = $3
WHERE id = $1;

-- name: DeleteTPObjectSupervisorsCascade :exec
DELETE FROM object_supervisors
WHERE object_supervisors.object_id = (
    SELECT DISTINCT objects.id
    FROM objects
    INNER JOIN tp_objects ON tp_objects.id = objects.object_detailed_id
    WHERE
        objects.project_id = $1
        AND tp_objects.id = $2
        AND objects.type = 'tp_objects'
);

-- name: DeleteTPObjectTeamsCascade :exec
DELETE FROM object_teams
WHERE object_teams.object_id = (
    SELECT DISTINCT objects.id
    FROM objects
    INNER JOIN tp_objects ON tp_objects.id = objects.object_detailed_id
    WHERE
        objects.project_id = $1
        AND tp_objects.id = $2
        AND objects.type = 'tp_objects'
);

-- name: DeleteTPObjectDetail :exec
DELETE FROM tp_objects WHERE id = $1;

-- name: DeleteObjectByTPObjectDetailedID :exec
DELETE FROM objects
WHERE object_detailed_id = $1 AND type = 'tp_objects';

-- name: ListTPObjectNamesByProject :many
SELECT COALESCE(objects.name, '')::text AS name
FROM objects
WHERE objects.project_id = $1 AND objects.type = 'tp_objects';

-- name: ListTPObjectNamesForSearch :many
SELECT
    COALESCE(objects.name, '')::text AS label,
    COALESCE(objects.name, '')::text AS value
FROM objects
INNER JOIN tp_objects ON tp_objects.id = objects.object_detailed_id
WHERE
    objects.project_id = $1
    AND objects.type = 'tp_objects';
