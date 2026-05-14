-- name: ListWorkersByProject :many
SELECT id, project_id, name, company_worker_id, job_title_in_company,
       job_title_in_project, mobile_number
FROM workers
WHERE id <> 1 AND project_id = $1
ORDER BY id DESC;

-- name: ListWorkersPaginatedFiltered :many
SELECT id, project_id, name, company_worker_id, job_title_in_company,
       job_title_in_project, mobile_number
FROM workers
WHERE
    project_id = $1
    AND (NULLIF($2::text, '') IS NULL OR name = $2)
    AND (NULLIF($3::text, '') IS NULL OR mobile_number = $3)
    AND (NULLIF($4::text, '') IS NULL OR job_title_in_company = $4)
    AND (NULLIF($5::text, '') IS NULL OR job_title_in_project = $5)
    AND (NULLIF($6::text, '') IS NULL OR company_worker_id = $6)
ORDER BY id DESC
LIMIT $7 OFFSET $8;

-- name: ListWorkersByJobTitleInProject :many
SELECT id, project_id, name, company_worker_id, job_title_in_company,
       job_title_in_project, mobile_number
FROM workers
WHERE job_title_in_project = $1 AND project_id = $2
ORDER BY id DESC;

-- name: GetWorker :one
SELECT id, project_id, name, company_worker_id, job_title_in_company,
       job_title_in_project, mobile_number
FROM workers
WHERE id = $1;

-- name: CountWorkersFiltered :one
SELECT COUNT(*)::bigint
FROM workers
WHERE
    project_id = $1
    AND (NULLIF($2::text, '') IS NULL OR name = $2)
    AND (NULLIF($3::text, '') IS NULL OR mobile_number = $3)
    AND (NULLIF($4::text, '') IS NULL OR job_title_in_company = $4)
    AND (NULLIF($5::text, '') IS NULL OR job_title_in_project = $5);

-- name: CreateWorker :one
INSERT INTO workers (project_id, name, company_worker_id, job_title_in_company,
                     job_title_in_project, mobile_number)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, project_id, name, company_worker_id, job_title_in_company,
          job_title_in_project, mobile_number;

-- name: UpdateWorker :one
UPDATE workers
SET project_id = $2, name = $3, company_worker_id = $4,
    job_title_in_company = $5, job_title_in_project = $6, mobile_number = $7
WHERE id = $1
RETURNING id, project_id, name, company_worker_id, job_title_in_company,
          job_title_in_project, mobile_number;

-- name: DeleteWorker :exec
DELETE FROM workers WHERE id = $1;

-- name: CreateWorkersBatch :copyfrom
INSERT INTO workers (project_id, name, company_worker_id, job_title_in_company,
                     job_title_in_project, mobile_number)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: ListDistinctWorkerNames :many
SELECT DISTINCT(COALESCE(name, '')::text)
FROM workers
WHERE project_id = $1 AND COALESCE(name, '') <> '';

-- name: ListDistinctWorkerJobTitlesInCompany :many
SELECT DISTINCT(COALESCE(job_title_in_company, '')::text)
FROM workers
WHERE project_id = $1 AND COALESCE(job_title_in_company, '') <> '';

-- name: ListDistinctWorkerJobTitlesInProject :many
SELECT DISTINCT(COALESCE(job_title_in_project, '')::text)
FROM workers
WHERE project_id = $1 AND COALESCE(job_title_in_project, '') <> '';

-- name: ListDistinctWorkerCompanyIDs :many
SELECT DISTINCT(COALESCE(company_worker_id, '')::text)
FROM workers
WHERE project_id = $1 AND COALESCE(company_worker_id, '') <> '';

-- name: ListDistinctWorkerMobileNumbers :many
SELECT DISTINCT(COALESCE(mobile_number, '')::text)
FROM workers
WHERE project_id = $1 AND COALESCE(mobile_number, '') <> '';

-- name: GetWorkerByCompanyID :one
SELECT id, project_id, name, company_worker_id, job_title_in_company,
       job_title_in_project, mobile_number
FROM workers
WHERE company_worker_id = $1
LIMIT 1;

-- name: GetWorkerByName :one
SELECT id, project_id, name, company_worker_id, job_title_in_company,
       job_title_in_project, mobile_number
FROM workers
WHERE name = $1
LIMIT 1;
