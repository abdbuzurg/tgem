-- name: CountInvoiceInputs :one
SELECT COUNT(*)::bigint FROM invoice_inputs WHERE project_id = $1;

-- name: CountInvoiceOutputs :one
SELECT COUNT(*)::bigint FROM invoice_outputs WHERE project_id = $1;

-- name: CountInvoiceReturns :one
SELECT COUNT(*)::bigint FROM invoice_returns WHERE project_id = $1;

-- name: CountInvoiceWriteOffs :one
SELECT COUNT(*)::bigint FROM invoice_write_offs WHERE project_id = $1;

-- name: ListInvoiceInputUniqueCreators :many
SELECT DISTINCT(COALESCE(released_worker_id, 0)::bigint)
FROM invoice_inputs
WHERE project_id = $1;

-- name: CountInvoiceInputCreatorInvoices :one
SELECT COUNT(*)::bigint
FROM invoice_inputs
WHERE project_id = $1 AND released_worker_id = $2;

-- name: ListInvoiceOutputUniqueCreators :many
SELECT DISTINCT(COALESCE(released_worker_id, 0)::bigint)
FROM invoice_outputs
WHERE project_id = $1;

-- name: CountInvoiceOutputCreatorInvoices :one
SELECT COUNT(*)::bigint
FROM invoice_outputs
WHERE project_id = $1 AND released_worker_id = $2;

-- name: CountMaterialInInvoices :many
SELECT
    COALESCE(invoice_materials.amount, 0)::float8       AS amount,
    COALESCE(invoice_materials.invoice_type, '')::text  AS invoice_type
FROM invoice_materials
INNER JOIN material_costs ON invoice_materials.material_cost_id = material_costs.id
INNER JOIN materials ON materials.id = material_costs.material_id
WHERE materials.id = $1;

-- name: CountMaterialInLocations :many
SELECT
    COALESCE(material_locations.amount, 0)::float8        AS amount,
    COALESCE(material_locations.location_type, '')::text  AS location_type
FROM material_locations
INNER JOIN material_costs ON material_locations.material_cost_id = material_costs.id
INNER JOIN materials ON materials.id = material_costs.material_id
WHERE materials.id = $1;
