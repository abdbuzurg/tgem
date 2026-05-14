-- name: ListSubstationCellObjectsPaginated :many
SELECT DISTINCT
    objects.id                          AS object_id,
    substation_cell_objects.id          AS object_detailed_id,
    COALESCE(objects.name, '')::text    AS name,
    COALESCE(objects.status, '')::text  AS status
FROM objects
INNER JOIN substation_cell_objects ON objects.object_detailed_id = substation_cell_objects.id
FULL JOIN object_teams ON object_teams.object_id = objects.id
FULL JOIN object_supervisors ON object_supervisors.object_id = objects.id
FULL JOIN substation_cell_nourashes_substation_objects
    ON substation_cell_nourashes_substation_objects.substation_cell_object_id = objects.id
WHERE
    objects.type = 'substation_cell_objects'
    AND objects.project_id = $1
    AND (NULLIF($2::text, '') IS NULL OR objects.name = $2)
    AND (NULLIF($3::bigint, 0) IS NULL OR object_teams.team_id = $3)
    AND (NULLIF($4::bigint, 0) IS NULL OR object_supervisors.supervisor_worker_id = $4)
    AND (NULLIF($5::bigint, 0) IS NULL OR substation_cell_nourashes_substation_objects.substation_object_id = $5)
ORDER BY substation_cell_objects.id DESC
LIMIT $6 OFFSET $7;

-- name: CountSubstationCellObjectsFiltered :one
SELECT COUNT(*)::bigint
FROM (
    SELECT DISTINCT objects.id
    FROM objects
    INNER JOIN substation_cell_objects ON objects.object_detailed_id = substation_cell_objects.id
    FULL JOIN object_teams ON object_teams.object_id = objects.id
    FULL JOIN object_supervisors ON object_supervisors.object_id = objects.id
    FULL JOIN substation_cell_nourashes_substation_objects
        ON substation_cell_nourashes_substation_objects.substation_cell_object_id = objects.id
    WHERE
        objects.type = 'substation_cell_objects'
        AND objects.project_id = $1
        AND (NULLIF($2::text, '') IS NULL OR objects.name = $2)
        AND (NULLIF($3::bigint, 0) IS NULL OR object_teams.team_id = $3)
        AND (NULLIF($4::bigint, 0) IS NULL OR object_supervisors.supervisor_worker_id = $4)
        AND (NULLIF($5::bigint, 0) IS NULL OR substation_cell_nourashes_substation_objects.substation_object_id = $5)
) sub;

-- name: CreateSubstationCellObjectDetail :one
-- substation_cell_objects has only an id column.
INSERT INTO substation_cell_objects DEFAULT VALUES
RETURNING id;

-- name: DeleteSubstationCellObjectSupervisorsCascade :exec
-- The GORM-era SQL had a copy-paste typo (`stvt_objects.id` instead of
-- `substation_cell_objects.id`) that would have raised "missing
-- FROM-clause entry" at execution time. Fixed during phase 6.
DELETE FROM object_supervisors
WHERE object_supervisors.object_id = (
    SELECT DISTINCT objects.id
    FROM objects
    INNER JOIN substation_cell_objects ON substation_cell_objects.id = objects.object_detailed_id
    WHERE
        objects.project_id = $1
        AND substation_cell_objects.id = $2
        AND objects.type = 'substation_cell_objects'
);

-- name: DeleteSubstationCellObjectTeamsCascade :exec
DELETE FROM object_teams
WHERE object_teams.object_id = (
    SELECT DISTINCT objects.id
    FROM objects
    INNER JOIN substation_cell_objects ON substation_cell_objects.id = objects.object_detailed_id
    WHERE
        objects.project_id = $1
        AND substation_cell_objects.id = $2
        AND objects.type = 'substation_cell_objects'
);

-- name: DeleteSubstationCellObjectNourashesCascade :exec
DELETE FROM substation_cell_nourashes_substation_objects
WHERE substation_cell_nourashes_substation_objects.substation_cell_object_id = (
    SELECT DISTINCT objects.id
    FROM objects
    INNER JOIN substation_cell_objects ON substation_cell_objects.id = objects.object_detailed_id
    WHERE
        objects.project_id = $1
        AND substation_cell_objects.id = $2
        AND objects.type = 'substation_cell_objects'
);

-- name: DeleteSubstationCellObjectDetail :exec
DELETE FROM substation_cell_objects WHERE id = $1;

-- name: DeleteObjectBySubstationCellDetailedID :exec
DELETE FROM objects
WHERE object_detailed_id = $1 AND type = 'substation_cell_objects';

-- name: ListSubstationCellObjectNamesForSearch :many
SELECT
    COALESCE(objects.name, '')::text AS label,
    COALESCE(objects.name, '')::text AS value
FROM objects
WHERE
    objects.project_id = $1
    AND objects.type = 'substation_cell_objects';

-- name: GetSubstationNameByCellObjectID :one
SELECT COALESCE(substation_objects.name, '')::text AS name
FROM substation_cell_nourashes_substation_objects
INNER JOIN objects AS substation_cell_objects
    ON substation_cell_objects.id = substation_cell_nourashes_substation_objects.substation_cell_object_id
INNER JOIN objects AS substation_objects
    ON substation_objects.id = substation_cell_nourashes_substation_objects.substation_object_id
WHERE
    substation_cell_nourashes_substation_objects.substation_cell_object_id = $1
LIMIT 1;

-- name: DeleteSubstationCellNourashesByCellObjectID :exec
DELETE FROM substation_cell_nourashes_substation_objects
WHERE substation_cell_object_id = $1;

-- name: CreateSubstationCellNourash :exec
INSERT INTO substation_cell_nourashes_substation_objects (substation_object_id, substation_cell_object_id)
VALUES ($1, $2);
