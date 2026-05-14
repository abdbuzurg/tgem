-- name: GetSubstationObjectDetail :one
SELECT id, voltage_class, number_of_transformers
FROM substation_objects
WHERE id = $1;

-- name: ListSubstationObjectsAll :many
SELECT id, object_detailed_id, type, name, status, project_id
FROM objects
WHERE type = 'substation_objects' AND project_id = $1;

-- name: ListSubstationObjectNamesByProject :many
SELECT COALESCE(objects.name, '')::text AS name
FROM objects
WHERE objects.type = 'substation_objects' AND objects.project_id = $1;

-- name: GetSubstationObjectByName :one
SELECT id, object_detailed_id, type, name, status, project_id
FROM objects
WHERE name = $1
LIMIT 1;

-- name: ListSubstationObjectsPaginated :many
SELECT DISTINCT
    objects.id                                              AS object_id,
    substation_objects.id                                   AS object_detailed_id,
    COALESCE(objects.name, '')::text                        AS name,
    COALESCE(objects.status, '')::text                      AS status,
    COALESCE(substation_objects.voltage_class, '')::text    AS voltage_class,
    COALESCE(substation_objects.number_of_transformers, 0)::bigint AS number_of_transformers
FROM objects
INNER JOIN substation_objects ON objects.object_detailed_id = substation_objects.id
FULL JOIN object_teams ON object_teams.object_id = objects.id
FULL JOIN object_supervisors ON object_supervisors.object_id = objects.id
WHERE
    objects.type = 'substation_objects'
    AND objects.project_id = $1
    AND (NULLIF($2::text, '') IS NULL OR objects.name = $2)
    AND (NULLIF($3::bigint, 0) IS NULL OR object_teams.team_id = $3)
    AND (NULLIF($4::bigint, 0) IS NULL OR object_supervisors.supervisor_worker_id = $4)
ORDER BY substation_objects.id DESC
LIMIT $5 OFFSET $6;

-- name: CountSubstationObjectsFiltered :one
SELECT COUNT(*)::bigint
FROM (
    SELECT DISTINCT objects.id
    FROM objects
    FULL JOIN object_teams ON object_teams.object_id = objects.id
    FULL JOIN object_supervisors ON object_supervisors.object_id = objects.id
    WHERE
        objects.type = 'substation_objects'
        AND objects.project_id = $1
        AND (NULLIF($2::text, '') IS NULL OR objects.name = $2)
        AND (NULLIF($3::bigint, 0) IS NULL OR object_teams.team_id = $3)
        AND (NULLIF($4::bigint, 0) IS NULL OR object_supervisors.supervisor_worker_id = $4)
) sub;

-- name: CreateSubstationObjectDetail :one
INSERT INTO substation_objects (voltage_class, number_of_transformers)
VALUES ($1, $2)
RETURNING id, voltage_class, number_of_transformers;

-- name: UpdateSubstationObjectDetail :exec
UPDATE substation_objects
SET voltage_class = $2, number_of_transformers = $3
WHERE id = $1;

-- name: DeleteSubstationObjectSupervisorsCascade :exec
DELETE FROM object_supervisors
WHERE object_supervisors.object_id = (
    SELECT DISTINCT objects.id
    FROM objects
    INNER JOIN substation_objects ON substation_objects.id = objects.object_detailed_id
    WHERE
        objects.project_id = $1
        AND substation_objects.id = $2
        AND objects.type = 'substation_objects'
);

-- name: DeleteSubstationObjectTeamsCascade :exec
DELETE FROM object_teams
WHERE object_teams.object_id = (
    SELECT DISTINCT objects.id
    FROM objects
    INNER JOIN substation_objects ON substation_objects.id = objects.object_detailed_id
    WHERE
        objects.project_id = $1
        AND substation_objects.id = $2
        AND objects.type = 'substation_objects'
);

-- name: DeleteSubstationObjectDetail :exec
DELETE FROM substation_objects WHERE id = $1;

-- name: DeleteObjectBySubstationDetailedID :exec
DELETE FROM objects
WHERE object_detailed_id = $1 AND type = 'substation_objects';

-- name: ListSubstationObjectNamesForSearch :many
SELECT
    COALESCE(objects.name, '')::text AS label,
    COALESCE(objects.name, '')::text AS value
FROM objects
INNER JOIN substation_objects ON substation_objects.id = objects.object_detailed_id
WHERE
    objects.project_id = $1
    AND objects.type = 'substation_objects';
