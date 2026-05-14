-- name: ListTeamsByProject :many
SELECT id, project_id, number, mobile_number, company
FROM teams
WHERE project_id = $1
ORDER BY id DESC;

-- name: ListTeamsPaginated :many
SELECT
    teams.id                                                      AS id,
    COALESCE(teams.number, '')::text                              AS team_number,
    COALESCE(workers.id, 0)::bigint                               AS leader_id,
    COALESCE(workers.name, '')::text                              AS leader_name,
    COALESCE(teams.mobile_number, '')::text                       AS team_mobile_number,
    COALESCE(teams.company, '')::text                             AS team_company
FROM teams
INNER JOIN team_leaders ON team_leaders.team_id = teams.id
INNER JOIN workers ON team_leaders.leader_worker_id = workers.id
WHERE
    teams.project_id = $1
    AND (NULLIF($2::text, '') IS NULL OR teams.number = $2)
    AND (NULLIF($3::text, '') IS NULL OR teams.mobile_number = $3)
    AND (NULLIF($4::text, '') IS NULL OR teams.company = $4)
    AND (NULLIF($5::bigint, 0) IS NULL OR team_leaders.leader_worker_id = $5)
ORDER BY teams.id DESC
LIMIT $6 OFFSET $7;

-- name: GetTeam :one
SELECT id, project_id, number, mobile_number, company
FROM teams
WHERE id = $1;

-- name: GetTeamByNumber :one
SELECT id, project_id, number, mobile_number, company
FROM teams
WHERE number = $1
LIMIT 1;

-- name: ListTeamsByIDs :many
SELECT id, project_id, number, mobile_number, company
FROM teams
WHERE id = ANY(sqlc.arg(ids)::bigint[]);

-- name: CreateTeam :one
INSERT INTO teams (project_id, number, mobile_number, company)
VALUES ($1, $2, $3, $4)
RETURNING id, project_id, number, mobile_number, company;

-- name: UpdateTeam :exec
UPDATE teams
SET number = $2, mobile_number = $3, company = $4
WHERE id = $1;

-- name: DeleteTeam :exec
DELETE FROM teams WHERE id = $1;

-- name: CountTeamsFiltered :one
SELECT COUNT(*)::bigint
FROM teams
INNER JOIN team_leaders ON team_leaders.team_id = teams.id
INNER JOIN workers ON team_leaders.leader_worker_id = workers.id
WHERE
    teams.project_id = $1
    AND (NULLIF($2::text, '') IS NULL OR teams.number = $2)
    AND (NULLIF($3::text, '') IS NULL OR teams.mobile_number = $3)
    AND (NULLIF($4::text, '') IS NULL OR teams.company = $4)
    AND (NULLIF($5::bigint, 0) IS NULL OR team_leaders.leader_worker_id = $5);

-- name: ListTeamNumberAndLeadersByID :many
SELECT DISTINCT ON (teams.id)
    COALESCE(teams.number, '')::text   AS team_number,
    COALESCE(workers.name, '')::text   AS team_leader_name
FROM teams
INNER JOIN team_leaders ON team_leaders.team_id = teams.id
INNER JOIN workers ON workers.id = team_leaders.leader_worker_id
WHERE teams.project_id = $1 AND teams.id = $2
ORDER BY teams.id, team_leaders.id;

-- name: TeamNumberExistsForCreate :one
SELECT EXISTS(
    SELECT 1 FROM teams WHERE teams.number = $1 AND teams.project_id = $2
)::boolean;

-- name: TeamNumberExistsForUpdate :one
SELECT EXISTS(
    SELECT 1 FROM teams WHERE teams.number = $1 AND teams.id <> $2 AND teams.project_id = $3
)::boolean;

-- name: ListTeamsForSelect :many
SELECT DISTINCT ON (teams.id)
    teams.id                            AS id,
    COALESCE(teams.number, '')::text    AS team_number,
    COALESCE(workers.name, '')::text    AS team_leader_name
FROM teams
INNER JOIN team_leaders ON team_leaders.team_id = teams.id
INNER JOIN workers ON workers.id = team_leaders.leader_worker_id
WHERE teams.project_id = $1
ORDER BY teams.id, team_leaders.id;

-- name: ListDistinctTeamNumbers :many
SELECT DISTINCT(COALESCE(number, '')::text) FROM teams WHERE project_id = $1;

-- name: ListDistinctTeamMobileNumbers :many
SELECT DISTINCT(COALESCE(mobile_number, '')::text) FROM teams WHERE project_id = $1;

-- name: ListDistinctTeamCompanies :many
SELECT DISTINCT(COALESCE(company, '')::text) FROM teams WHERE project_id = $1;
