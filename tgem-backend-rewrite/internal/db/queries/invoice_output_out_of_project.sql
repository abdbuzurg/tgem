-- name: GetInvoiceOutputOutOfProject :one
SELECT id, project_id, delivery_code, released_worker_id, name_of_project,
       date_of_invoice, notes, confirmation
FROM invoice_output_out_of_projects
WHERE id = $1;

-- name: GetInvoiceOutputOutOfProjectByDeliveryCode :one
SELECT id, project_id, delivery_code, released_worker_id, name_of_project,
       date_of_invoice, notes, confirmation
FROM invoice_output_out_of_projects
WHERE delivery_code = $1
  AND project_id = $2
LIMIT 1;

-- name: ListInvoiceOutputOutOfProjectsPaginated :many
SELECT
    invoice_output_out_of_projects.id                              AS id,
    COALESCE(invoice_output_out_of_projects.name_of_project, '')::text  AS name_of_project,
    COALESCE(invoice_output_out_of_projects.delivery_code, '')::text    AS delivery_code,
    COALESCE(workers.name, '')::text                               AS released_worker_name,
    invoice_output_out_of_projects.date_of_invoice                 AS date_of_invoice,
    COALESCE(invoice_output_out_of_projects.confirmation, false)::boolean AS confirmation
FROM invoice_output_out_of_projects
INNER JOIN workers ON invoice_output_out_of_projects.released_worker_id = workers.id
WHERE
    invoice_output_out_of_projects.project_id = $1
    AND (NULLIF($2::text, '') IS NULL OR name_of_project = $2)
    AND (NULLIF($3::bigint, 0) IS NULL OR released_worker_id = $3)
ORDER BY invoice_output_out_of_projects.id DESC
LIMIT $4 OFFSET $5;

-- name: CountInvoiceOutputOutOfProjects :one
SELECT COUNT(*)::bigint
FROM invoice_output_out_of_projects
WHERE
    project_id = $1
    AND (NULLIF($2::text, '') IS NULL OR name_of_project = $2)
    AND (NULLIF($3::bigint, 0) IS NULL OR released_worker_id = $3);

-- name: CreateInvoiceOutputOutOfProject :one
INSERT INTO invoice_output_out_of_projects (
    project_id, delivery_code, released_worker_id, name_of_project,
    date_of_invoice, notes, confirmation
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, project_id, delivery_code, released_worker_id, name_of_project,
          date_of_invoice, notes, confirmation;

-- name: UpdateInvoiceOutputOutOfProject :one
UPDATE invoice_output_out_of_projects
SET project_id = $2, delivery_code = $3, released_worker_id = $4,
    name_of_project = $5, date_of_invoice = $6, notes = $7, confirmation = $8
WHERE id = $1
RETURNING id, project_id, delivery_code, released_worker_id, name_of_project,
          date_of_invoice, notes, confirmation;

-- name: DeleteInvoiceOutputOutOfProject :exec
DELETE FROM invoice_output_out_of_projects WHERE id = $1;

-- name: ConfirmInvoiceOutputOutOfProject :exec
UPDATE invoice_output_out_of_projects
SET confirmation = true
WHERE id = $1;

-- name: ListInvoiceOutputOutOfProjectMaterialsForEdit :many
SELECT
    materials.id                                                  AS material_id,
    COALESCE(materials.name, '')::text                            AS material_name,
    COALESCE(materials.unit, '')::text                            AS material_unit,
    COALESCE(material_locations.amount, 0)::numeric               AS warehouse_amount,
    COALESCE(invoice_materials.amount, 0)::numeric                AS amount,
    COALESCE(invoice_materials.notes, '')::text                   AS notes,
    COALESCE(materials.has_serial_number, false)::boolean         AS has_serial_number
FROM invoice_materials
INNER JOIN material_costs ON material_costs.id = invoice_materials.material_cost_id
INNER JOIN materials ON materials.id = material_costs.material_id
INNER JOIN material_locations ON material_locations.material_cost_id = invoice_materials.material_cost_id
WHERE
    material_locations.location_type = 'warehouse'
    AND invoice_materials.invoice_type = 'output-out-of-project'
    AND invoice_materials.invoice_id = $1
    AND invoice_materials.project_id = $2
ORDER BY materials.id;

-- name: ListInvoiceOutputOutOfProjectUniqueNameOfProjects :many
SELECT DISTINCT(COALESCE(invoice_output_out_of_projects.name_of_project, '')::text)
FROM invoice_output_out_of_projects
WHERE invoice_output_out_of_projects.project_id = $1;

-- name: ListInvoiceOutputOutOfProjectReportData :many
SELECT
    invoice_output_out_of_projects.id                              AS id,
    COALESCE(invoice_output_out_of_projects.name_of_project, '')::text  AS name_of_project,
    COALESCE(invoice_output_out_of_projects.delivery_code, '')::text    AS delivery_code,
    COALESCE(workers.name, '')::text                               AS released_worker_name,
    invoice_output_out_of_projects.date_of_invoice                 AS date_of_invoice
FROM invoice_output_out_of_projects
INNER JOIN workers ON invoice_output_out_of_projects.released_worker_id = workers.id
WHERE
    invoice_output_out_of_projects.project_id = $1
    AND COALESCE(invoice_output_out_of_projects.confirmation, false) = true
    AND (NULLIF($2::timestamptz, '0001-01-01 00:00:00 UTC') IS NULL OR $2 <= invoice_output_out_of_projects.date_of_invoice)
    AND (NULLIF($3::timestamptz, '0001-01-01 00:00:00 UTC') IS NULL OR invoice_output_out_of_projects.date_of_invoice <= $3)
ORDER BY invoice_output_out_of_projects.id DESC;
