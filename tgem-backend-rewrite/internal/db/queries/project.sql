-- name: ListProjects :many
SELECT id, name, client, budget, budget_currency, description,
       signed_date_of_contract, date_start, date_end, project_manager
FROM projects
ORDER BY id DESC;

-- name: ListProjectsExcludeAdmin :many
SELECT id, name, client, budget, budget_currency, description,
       signed_date_of_contract, date_start, date_end, project_manager
FROM projects
WHERE name <> 'Администрирование'
ORDER BY id DESC
LIMIT $1 OFFSET $2;

-- name: GetProject :one
SELECT id, name, client, budget, budget_currency, description,
       signed_date_of_contract, date_start, date_end, project_manager
FROM projects
WHERE id = $1;

-- name: GetProjectName :one
SELECT COALESCE(name, '')::text AS name FROM projects WHERE id = $1;

-- name: CreateProject :one
INSERT INTO projects (name, client, budget, budget_currency, description,
                      signed_date_of_contract, date_start, date_end, project_manager)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, name, client, budget, budget_currency, description,
          signed_date_of_contract, date_start, date_end, project_manager;

-- name: UpdateProject :one
UPDATE projects
SET name = $2, client = $3, budget = $4, budget_currency = $5, description = $6,
    signed_date_of_contract = $7, date_start = $8, date_end = $9, project_manager = $10
WHERE id = $1
RETURNING id, name, client, budget, budget_currency, description,
          signed_date_of_contract, date_start, date_end, project_manager;

-- name: DeleteProject :exec
DELETE FROM projects WHERE id = $1;

-- name: CountProjects :one
SELECT COUNT(*) FROM projects;
