-- name: GetMJDObjectDetail :one
SELECT id, model, amount_stores, amount_entrances, has_basement
FROM mjd_objects
WHERE id = $1;

-- name: ListMJDObjectsPaginated :many
SELECT DISTINCT
    objects.id                                    AS object_id,
    mjd_objects.id                                AS object_detailed_id,
    COALESCE(objects.name, '')::text              AS name,
    COALESCE(objects.status, '')::text            AS status,
    COALESCE(mjd_objects.model, '')::text         AS model,
    COALESCE(mjd_objects.amount_stores, 0)::bigint    AS amount_stores,
    COALESCE(mjd_objects.amount_entrances, 0)::bigint AS amount_entrances,
    COALESCE(mjd_objects.has_basement, false)::boolean AS has_basement
FROM objects
INNER JOIN mjd_objects ON mjd_objects.id = objects.object_detailed_id
FULL JOIN object_teams ON object_teams.object_id = objects.id
FULL JOIN object_supervisors ON object_supervisors.object_id = objects.id
FULL JOIN tp_nourashes_objects ON tp_nourashes_objects.target_id = objects.id
WHERE
    objects.type = 'mjd_objects'
    AND objects.project_id = $1
    AND (NULLIF($2::text, '') IS NULL OR objects.name = $2)
    AND (NULLIF($3::bigint, 0) IS NULL OR object_teams.team_id = $3)
    AND (NULLIF($4::bigint, 0) IS NULL OR object_supervisors.supervisor_worker_id = $4)
    AND (NULLIF($5::bigint, 0) IS NULL OR tp_nourashes_objects.tp_object_id = $5)
ORDER BY mjd_objects.id DESC
LIMIT $6 OFFSET $7;

-- name: CountMJDObjectsFiltered :one
SELECT COUNT(*)::bigint
FROM (
    SELECT DISTINCT objects.id
    FROM objects
    INNER JOIN mjd_objects ON mjd_objects.id = objects.object_detailed_id
    FULL JOIN object_teams ON object_teams.object_id = objects.id
    FULL JOIN object_supervisors ON object_supervisors.object_id = objects.id
    FULL JOIN tp_nourashes_objects ON tp_nourashes_objects.target_id = objects.id
    WHERE
        objects.type = 'mjd_objects'
        AND objects.project_id = $1
        AND (NULLIF($2::text, '') IS NULL OR objects.name = $2)
        AND (NULLIF($3::bigint, 0) IS NULL OR object_teams.team_id = $3)
        AND (NULLIF($4::bigint, 0) IS NULL OR object_supervisors.supervisor_worker_id = $4)
        AND (NULLIF($5::bigint, 0) IS NULL OR tp_nourashes_objects.tp_object_id = $5)
) sub;

-- name: CreateMJDObjectDetail :one
INSERT INTO mjd_objects (model, amount_stores, amount_entrances, has_basement)
VALUES ($1, $2, $3, $4)
RETURNING id, model, amount_stores, amount_entrances, has_basement;

-- name: UpdateMJDObjectDetail :exec
UPDATE mjd_objects
SET model = $2, amount_stores = $3, amount_entrances = $4, has_basement = $5
WHERE id = $1;

-- name: DeleteMJDObjectSupervisorsCascade :exec
DELETE FROM object_supervisors
WHERE object_supervisors.object_id = (
    SELECT DISTINCT objects.id
    FROM objects
    INNER JOIN mjd_objects ON mjd_objects.id = objects.object_detailed_id
    WHERE
        objects.project_id = $1
        AND mjd_objects.id = $2
        AND objects.type = 'mjd_objects'
);

-- name: DeleteMJDObjectTeamsCascade :exec
DELETE FROM object_teams
WHERE object_teams.object_id = (
    SELECT DISTINCT objects.id
    FROM objects
    INNER JOIN mjd_objects ON mjd_objects.id = objects.object_detailed_id
    WHERE
        objects.project_id = $1
        AND mjd_objects.id = $2
        AND objects.type = 'mjd_objects'
);

-- name: DeleteMJDObjectTPNourashesCascade :exec
DELETE FROM tp_nourashes_objects
WHERE tp_nourashes_objects.target_id = (
    SELECT DISTINCT objects.id
    FROM objects
    INNER JOIN mjd_objects ON mjd_objects.id = objects.object_detailed_id
    WHERE
        objects.project_id = $1
        AND mjd_objects.id = $2
        AND objects.type = 'mjd_objects'
);

-- name: DeleteMJDObjectDetail :exec
DELETE FROM mjd_objects WHERE id = $1;

-- name: DeleteObjectByMJDDetailedID :exec
DELETE FROM objects
WHERE object_detailed_id = $1 AND type = 'mjd_objects';

-- name: ListMJDObjectNamesForSearch :many
SELECT
    COALESCE(objects.name, '')::text AS label,
    COALESCE(objects.name, '')::text AS value
FROM objects
INNER JOIN mjd_objects ON mjd_objects.id = objects.object_detailed_id
WHERE
    objects.project_id = $1
    AND objects.type = 'mjd_objects';
