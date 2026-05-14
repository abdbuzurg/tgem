-- name: GetSIPObjectDetail :one
SELECT id, amount_feeders
FROM s_ip_objects
WHERE id = $1;

-- name: ListSIPObjectsPaginated :many
SELECT DISTINCT
    objects.id                                       AS object_id,
    s_ip_objects.id                                  AS object_detailed_id,
    COALESCE(objects.name, '')::text                 AS name,
    COALESCE(objects.status, '')::text               AS status,
    COALESCE(s_ip_objects.amount_feeders, 0)::bigint AS amount_feeders
FROM objects
INNER JOIN s_ip_objects ON objects.object_detailed_id = s_ip_objects.id
FULL JOIN object_teams ON object_teams.object_id = objects.id
FULL JOIN object_supervisors ON object_supervisors.object_id = objects.id
FULL JOIN tp_nourashes_objects ON tp_nourashes_objects.target_id = objects.id
WHERE
    objects.type = 'sip_objects'
    AND objects.project_id = $1
    AND (NULLIF($2::text, '') IS NULL OR objects.name = $2)
    AND (NULLIF($3::bigint, 0) IS NULL OR object_teams.team_id = $3)
    AND (NULLIF($4::bigint, 0) IS NULL OR object_supervisors.supervisor_worker_id = $4)
ORDER BY s_ip_objects.id DESC
LIMIT $5 OFFSET $6;

-- name: CountSIPObjectsFiltered :one
SELECT COUNT(*)::bigint
FROM (
    SELECT DISTINCT objects.id
    FROM objects
    INNER JOIN s_ip_objects ON objects.object_detailed_id = s_ip_objects.id
    FULL JOIN object_teams ON object_teams.object_id = objects.id
    FULL JOIN object_supervisors ON object_supervisors.object_id = objects.id
    FULL JOIN tp_nourashes_objects ON tp_nourashes_objects.target_id = objects.id
    WHERE
        objects.type = 'sip_objects'
        AND objects.project_id = $1
        AND (NULLIF($2::text, '') IS NULL OR objects.name = $2)
        AND (NULLIF($3::bigint, 0) IS NULL OR object_teams.team_id = $3)
        AND (NULLIF($4::bigint, 0) IS NULL OR object_supervisors.supervisor_worker_id = $4)
) sub;

-- name: CreateSIPObjectDetail :one
INSERT INTO s_ip_objects (amount_feeders)
VALUES ($1)
RETURNING id, amount_feeders;

-- name: UpdateSIPObjectDetail :exec
UPDATE s_ip_objects
SET amount_feeders = $2
WHERE id = $1;

-- name: DeleteSIPObjectSupervisorsCascade :exec
DELETE FROM object_supervisors
WHERE object_supervisors.object_id = (
    SELECT DISTINCT objects.id
    FROM objects
    INNER JOIN s_ip_objects ON s_ip_objects.id = objects.object_detailed_id
    WHERE
        objects.project_id = $1
        AND s_ip_objects.id = $2
        AND objects.type = 'sip_objects'
);

-- name: DeleteSIPObjectTeamsCascade :exec
DELETE FROM object_teams
WHERE object_teams.object_id = (
    SELECT DISTINCT objects.id
    FROM objects
    INNER JOIN s_ip_objects ON s_ip_objects.id = objects.object_detailed_id
    WHERE
        objects.project_id = $1
        AND s_ip_objects.id = $2
        AND objects.type = 'sip_objects'
);

-- name: DeleteSIPObjectDetail :exec
DELETE FROM s_ip_objects WHERE id = $1;

-- name: DeleteObjectBySIPDetailedID :exec
DELETE FROM objects
WHERE object_detailed_id = $1 AND type = 'sip_objects';

-- name: ListSIPObjectNamesForSearch :many
SELECT
    COALESCE(objects.name, '')::text AS label,
    COALESCE(objects.name, '')::text AS value
FROM objects
INNER JOIN s_ip_objects ON s_ip_objects.id = objects.object_detailed_id
WHERE
    objects.project_id = $1
    AND objects.type = 'sip_objects';
