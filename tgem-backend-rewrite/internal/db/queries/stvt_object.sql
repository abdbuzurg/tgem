-- name: GetSTVTObjectDetail :one
SELECT id, voltage_class, tt_coefficient
FROM stvt_objects
WHERE id = $1;

-- name: ListSTVTObjectsPaginated :many
SELECT DISTINCT
    objects.id                                    AS object_id,
    stvt_objects.id                               AS object_detailed_id,
    COALESCE(objects.name, '')::text              AS name,
    COALESCE(objects.status, '')::text            AS status,
    COALESCE(stvt_objects.voltage_class, '')::text  AS voltage_class,
    COALESCE(stvt_objects.tt_coefficient, '')::text AS tt_coefficient
FROM objects
INNER JOIN stvt_objects ON objects.object_detailed_id = stvt_objects.id
FULL JOIN object_teams ON object_teams.object_id = objects.id
FULL JOIN object_supervisors ON object_supervisors.object_id = objects.id
FULL JOIN tp_nourashes_objects ON tp_nourashes_objects.target_id = objects.id
WHERE
    objects.type = 'stvt_objects'
    AND objects.project_id = $1
    AND (NULLIF($2::text, '') IS NULL OR objects.name = $2)
    AND (NULLIF($3::bigint, 0) IS NULL OR object_teams.team_id = $3)
    AND (NULLIF($4::bigint, 0) IS NULL OR object_supervisors.supervisor_worker_id = $4)
ORDER BY stvt_objects.id DESC
LIMIT $5 OFFSET $6;

-- name: CountSTVTObjectsFiltered :one
SELECT COUNT(*)::bigint
FROM (
    SELECT DISTINCT objects.id
    FROM objects
    INNER JOIN stvt_objects ON objects.object_detailed_id = stvt_objects.id
    FULL JOIN object_teams ON object_teams.object_id = objects.id
    FULL JOIN object_supervisors ON object_supervisors.object_id = objects.id
    FULL JOIN tp_nourashes_objects ON tp_nourashes_objects.target_id = objects.id
    WHERE
        objects.type = 'stvt_objects'
        AND objects.project_id = $1
        AND (NULLIF($2::text, '') IS NULL OR objects.name = $2)
        AND (NULLIF($3::bigint, 0) IS NULL OR object_teams.team_id = $3)
        AND (NULLIF($4::bigint, 0) IS NULL OR object_supervisors.supervisor_worker_id = $4)
) sub;

-- name: CreateSTVTObjectDetail :one
INSERT INTO stvt_objects (voltage_class, tt_coefficient)
VALUES ($1, $2)
RETURNING id, voltage_class, tt_coefficient;

-- name: UpdateSTVTObjectDetail :exec
UPDATE stvt_objects
SET voltage_class = $2, tt_coefficient = $3
WHERE id = $1;

-- name: DeleteSTVTObjectSupervisorsCascade :exec
DELETE FROM object_supervisors
WHERE object_supervisors.object_id = (
    SELECT DISTINCT objects.id
    FROM objects
    INNER JOIN stvt_objects ON stvt_objects.id = objects.object_detailed_id
    WHERE
        objects.project_id = $1
        AND stvt_objects.id = $2
        AND objects.type = 'stvt_objects'
);

-- name: DeleteSTVTObjectTeamsCascade :exec
DELETE FROM object_teams
WHERE object_teams.object_id = (
    SELECT DISTINCT objects.id
    FROM objects
    INNER JOIN stvt_objects ON stvt_objects.id = objects.object_detailed_id
    WHERE
        objects.project_id = $1
        AND stvt_objects.id = $2
        AND objects.type = 'stvt_objects'
);

-- name: DeleteSTVTObjectDetail :exec
DELETE FROM stvt_objects WHERE id = $1;

-- name: DeleteObjectBySTVTDetailedID :exec
DELETE FROM objects
WHERE object_detailed_id = $1 AND type = 'stvt_objects';

-- name: ListSTVTObjectNamesForSearch :many
SELECT
    COALESCE(objects.name, '')::text AS label,
    COALESCE(objects.name, '')::text AS value
FROM objects
INNER JOIN stvt_objects ON stvt_objects.id = objects.object_detailed_id
WHERE
    objects.project_id = $1
    AND objects.type = 'stvt_objects';
