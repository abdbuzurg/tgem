-- name: ListWorkerAttendancesPaginated :many
SELECT
    worker_attendances.id                                          AS id,
    COALESCE(workers.name, '')::text                               AS worker_name,
    COALESCE(workers.company_worker_id, '')::text                  AS company_worker_id,
    worker_attendances.start                                       AS start,
    worker_attendances."end"                                       AS "end"
FROM worker_attendances
INNER JOIN workers ON workers.id = worker_attendances.worker_id
WHERE worker_attendances.project_id = $1;

-- name: CountWorkerAttendances :one
SELECT COUNT(*)::bigint FROM worker_attendances WHERE project_id = $1;

-- name: CreateWorkerAttendancesBatch :copyfrom
INSERT INTO worker_attendances (project_id, worker_id, start, "end")
VALUES ($1, $2, $3, $4);
