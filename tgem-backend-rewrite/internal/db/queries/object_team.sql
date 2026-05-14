-- name: ListTeamNumbersByObjectID :many
SELECT COALESCE(teams.number, '')::text AS team_number
FROM object_teams
INNER JOIN objects ON objects.id = object_teams.object_id
INNER JOIN teams ON teams.id = object_teams.team_id
WHERE objects.id = $1;

-- name: ListTeamsForSelectByObjectID :many
SELECT
    teams.id                            AS id,
    COALESCE(teams.number, '')::text    AS team_number,
    COALESCE(workers.name, '')::text    AS team_leader_name
FROM object_teams
INNER JOIN teams ON teams.id = object_teams.team_id
INNER JOIN team_leaders ON team_leaders.team_id = object_teams.team_id
INNER JOIN workers ON workers.id = team_leaders.leader_worker_id
WHERE object_teams.object_id = $1;

-- name: CreateObjectTeam :exec
INSERT INTO object_teams (object_id, team_id) VALUES ($1, $2);

-- name: DeleteObjectTeamsByObjectID :exec
DELETE FROM object_teams WHERE object_id = $1;

-- name: CreateObjectTeamsBatch :copyfrom
INSERT INTO object_teams (object_id, team_id) VALUES ($1, $2);
