-- name: GetInvoiceObject :one
SELECT id, district_id, delivery_code, project_id, supervisor_worker_id,
       object_id, team_id, date_of_invoice, confirmed_by_operator,
       date_of_correction
FROM invoice_objects
WHERE id = $1;

-- name: CreateInvoiceObject :one
INSERT INTO invoice_objects (
    district_id, delivery_code, project_id, supervisor_worker_id,
    object_id, team_id, date_of_invoice, confirmed_by_operator,
    date_of_correction
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, district_id, delivery_code, project_id, supervisor_worker_id,
          object_id, team_id, date_of_invoice, confirmed_by_operator,
          date_of_correction;

-- name: DeleteInvoiceObject :exec
DELETE FROM invoice_objects WHERE id = $1;

-- name: CountInvoiceObjectsByProject :one
SELECT COUNT(*)::bigint FROM invoice_objects WHERE project_id = $1;

-- name: GetInvoiceObjectDescriptiveDataByID :one
SELECT
    invoice_objects.id                                          AS id,
    COALESCE(workers.name, '')::text                            AS supervisor_name,
    invoice_objects.district_id                                 AS district_id,
    COALESCE(districts.name, '')::text                          AS district_name,
    COALESCE(objects.name, '')::text                            AS object_name,
    COALESCE(objects.type, '')::text                            AS object_type,
    COALESCE(teams.number, '')::text                            AS team_number,
    invoice_objects.date_of_invoice                             AS date_of_invoice,
    COALESCE(invoice_objects.delivery_code, '')::text           AS delivery_code,
    COALESCE(invoice_objects.confirmed_by_operator, false)::boolean AS confirmed_by_operator
FROM invoice_objects
INNER JOIN workers ON workers.id = invoice_objects.supervisor_worker_id
LEFT JOIN districts ON districts.id = invoice_objects.district_id
INNER JOIN objects ON objects.id = invoice_objects.object_id
INNER JOIN teams ON teams.id = invoice_objects.team_id
WHERE invoice_objects.id = $1
  AND invoice_objects.project_id = $2;

-- name: ListInvoiceObjectsPaginated :many
SELECT
    invoice_objects.id                                          AS id,
    COALESCE(workers.name, '')::text                            AS supervisor_name,
    invoice_objects.district_id                                 AS district_id,
    COALESCE(districts.name, '')::text                          AS district_name,
    COALESCE(objects.name, '')::text                            AS object_name,
    COALESCE(objects.type, '')::text                            AS object_type,
    COALESCE(teams.number, '')::text                            AS team_number,
    invoice_objects.date_of_invoice                             AS date_of_invoice,
    COALESCE(invoice_objects.delivery_code, '')::text           AS delivery_code,
    COALESCE(invoice_objects.confirmed_by_operator, false)::boolean AS confirmed_by_operator
FROM invoice_objects
INNER JOIN workers ON workers.id = invoice_objects.supervisor_worker_id
LEFT JOIN districts ON districts.id = invoice_objects.district_id
INNER JOIN objects ON objects.id = invoice_objects.object_id
INNER JOIN teams ON teams.id = invoice_objects.team_id
WHERE invoice_objects.project_id = $1
ORDER BY invoice_objects.id DESC
LIMIT $2 OFFSET $3;

-- name: ListInvoiceObjectsForCorrection :many
SELECT
    invoice_objects.id                                          AS id,
    COALESCE(workers.name, '')::text                            AS supervisor_name,
    invoice_objects.district_id                                 AS district_id,
    COALESCE(districts.name, '')::text                          AS district_name,
    COALESCE(objects.name, '')::text                            AS object_name,
    teams.id                                                    AS team_id,
    COALESCE(teams.number, '')::text                            AS team_number,
    invoice_objects.date_of_invoice                             AS date_of_invoice,
    COALESCE(invoice_objects.delivery_code, '')::text           AS delivery_code,
    COALESCE(invoice_objects.confirmed_by_operator, false)::boolean AS confirmed_by_operator
FROM invoice_objects
INNER JOIN workers ON workers.id = invoice_objects.supervisor_worker_id
LEFT JOIN districts ON districts.id = invoice_objects.district_id
INNER JOIN objects ON objects.id = invoice_objects.object_id
INNER JOIN teams ON teams.id = invoice_objects.team_id
WHERE
    invoice_objects.project_id = $1
    AND COALESCE(invoice_objects.confirmed_by_operator, false) = false;

-- name: ListInvoiceObjectTeamsByObjectID :many
SELECT
    teams.id                                AS id,
    teams.project_id                        AS project_id,
    COALESCE(teams.number, '')::text        AS number,
    COALESCE(teams.mobile_number, '')::text AS mobile_number,
    COALESCE(teams.company, '')::text       AS company
FROM object_teams
INNER JOIN teams ON teams.id = object_teams.team_id
WHERE object_teams.object_id = $1;

-- name: ListInvoiceObjectOperationsBasedOnMaterialsInTeam :many
SELECT
    operations.id                                                AS id,
    operations.project_id                                        AS project_id,
    COALESCE(operations.name, '')::text                          AS name,
    COALESCE(operations.code, '')::text                          AS code,
    operations.cost_prime                                        AS cost_prime,
    operations.cost_with_customer                                AS cost_with_customer
FROM operations
INNER JOIN operation_materials ON operation_materials.operation_id = operations.id
WHERE operation_materials.material_id IN (
    SELECT materials.id
    FROM material_locations
    INNER JOIN material_costs ON material_locations.material_cost_id = material_costs.id
    INNER JOIN materials ON material_costs.material_id = materials.id
    WHERE
        material_locations.location_type = 'team'
        AND material_locations.location_id = $1
);

-- name: CreateInvoiceOperationsBatch :copyfrom
INSERT INTO invoice_operations (project_id, operation_id, invoice_id, invoice_type, amount, notes)
VALUES ($1, $2, $3, $4, $5, $6);
