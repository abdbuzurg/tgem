-- name: CreateTeamLeader :exec
INSERT INTO team_leaders (team_id, leader_worker_id) VALUES ($1, $2);

-- name: DeleteTeamLeadersByTeamID :exec
DELETE FROM team_leaders WHERE team_id = $1;

-- name: CreateTeamLeadersBatch :copyfrom
INSERT INTO team_leaders (team_id, leader_worker_id) VALUES ($1, $2);
