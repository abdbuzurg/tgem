-- name: CreateUserInProject :exec
INSERT INTO user_in_projects (user_id, project_id) VALUES ($1, $2);

-- name: DeleteUserInProjectsByProjectID :exec
DELETE FROM user_in_projects WHERE project_id = $1;

-- name: ListUserInProjectsByUserID :many
SELECT id, project_id, user_id
FROM user_in_projects
WHERE user_id = $1;

-- name: ListProjectNamesByUserID :many
SELECT COALESCE(name, '')::text
FROM projects
WHERE projects.id IN (
    SELECT project_id
    FROM user_in_projects
    WHERE user_id = $1
);

-- name: DeleteUserInProjectsByUserID :exec
DELETE FROM user_in_projects WHERE user_id = $1;

-- name: CreateUserInProjectsBatch :copyfrom
INSERT INTO user_in_projects (user_id, project_id) VALUES ($1, $2);
