-- name: ListInvoiceInputs :many
SELECT id, project_id, warehouse_manager_worker_id, released_worker_id,
       delivery_code, notes, date_of_invoice, confirmed
FROM invoice_inputs
ORDER BY id DESC;

-- name: ListInvoiceInputsPaginated :many
SELECT id, project_id, warehouse_manager_worker_id, released_worker_id,
       delivery_code, notes, date_of_invoice, confirmed
FROM invoice_inputs
ORDER BY id DESC
LIMIT $1 OFFSET $2;

-- name: GetInvoiceInput :one
SELECT id, project_id, warehouse_manager_worker_id, released_worker_id,
       delivery_code, notes, date_of_invoice, confirmed
FROM invoice_inputs
WHERE id = $1;

-- name: CreateInvoiceInput :one
INSERT INTO invoice_inputs (project_id, warehouse_manager_worker_id, released_worker_id,
                            delivery_code, notes, date_of_invoice, confirmed)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, project_id, warehouse_manager_worker_id, released_worker_id,
          delivery_code, notes, date_of_invoice, confirmed;

-- name: UpdateInvoiceInput :one
UPDATE invoice_inputs
SET project_id = $2, warehouse_manager_worker_id = $3, released_worker_id = $4,
    delivery_code = $5, notes = $6, date_of_invoice = $7, confirmed = $8
WHERE id = $1
RETURNING id, project_id, warehouse_manager_worker_id, released_worker_id,
          delivery_code, notes, date_of_invoice, confirmed;

-- name: DeleteInvoiceInput :exec
DELETE FROM invoice_inputs WHERE id = $1;

-- name: ConfirmInvoiceInput :exec
UPDATE invoice_inputs
SET confirmed = true
WHERE id = $1;

-- name: ListInvoiceInputsPaginatedFiltered :many
SELECT
    invoice_inputs.id                                                AS id,
    COALESCE(invoice_inputs.confirmed, false)::boolean               AS confirmation,
    COALESCE(invoice_inputs.delivery_code, '')::text                 AS delivery_code,
    COALESCE(warehouse_manager.name, '')::text                       AS warehouse_manager_name,
    COALESCE(released.name, '')::text                                AS released_name,
    invoice_inputs.date_of_invoice                                   AS date_of_invoice
FROM invoice_inputs
LEFT JOIN workers AS warehouse_manager ON warehouse_manager.id = invoice_inputs.warehouse_manager_worker_id
LEFT JOIN workers AS released ON released.id = invoice_inputs.released_worker_id
WHERE
    invoice_inputs.project_id = $1
    AND (NULLIF($2::bigint, 0) IS NULL OR warehouse_manager_worker_id = $2)
    AND (NULLIF($3::bigint, 0) IS NULL OR released_worker_id = $3)
    AND (NULLIF($4::text, '') IS NULL OR delivery_code = $4)
    AND (NULLIF($5::timestamptz, '0001-01-01 00:00:00 UTC') IS NULL OR $5 <= invoice_inputs.date_of_invoice)
    AND (NULLIF($6::timestamptz, '0001-01-01 00:00:00 UTC') IS NULL OR invoice_inputs.date_of_invoice <= $6)
ORDER BY invoice_inputs.id DESC
LIMIT $7 OFFSET $8;

-- name: CountInvoiceInputsFiltered :one
SELECT COUNT(*)::bigint
FROM invoice_inputs
LEFT JOIN workers AS warehouse_manager ON warehouse_manager.id = invoice_inputs.warehouse_manager_worker_id
LEFT JOIN workers AS released ON released.id = invoice_inputs.released_worker_id
WHERE
    invoice_inputs.project_id = $1
    AND (NULLIF($2::bigint, 0) IS NULL OR warehouse_manager_worker_id = $2)
    AND (NULLIF($3::bigint, 0) IS NULL OR released_worker_id = $3)
    AND (NULLIF($4::text, '') IS NULL OR delivery_code = $4)
    AND (NULLIF($5::timestamptz, '0001-01-01 00:00:00 UTC') IS NULL OR $5 <= invoice_inputs.date_of_invoice)
    AND (NULLIF($6::timestamptz, '0001-01-01 00:00:00 UTC') IS NULL OR invoice_inputs.date_of_invoice <= $6);

-- name: ListInvoiceInputsPaginatedByMaterials :many
SELECT
    invoice_inputs.id                                                AS id,
    COALESCE(invoice_inputs.confirmed, false)::boolean               AS confirmation,
    COALESCE(invoice_inputs.delivery_code, '')::text                 AS delivery_code,
    COALESCE(warehouse_manager.name, '')::text                       AS warehouse_manager_name,
    COALESCE(released.name, '')::text                                AS released_name,
    invoice_inputs.date_of_invoice                                   AS date_of_invoice
FROM invoice_inputs
LEFT JOIN workers AS warehouse_manager ON warehouse_manager.id = invoice_inputs.warehouse_manager_worker_id
LEFT JOIN workers AS released ON released.id = invoice_inputs.released_worker_id
WHERE
    invoice_inputs.project_id = $1
    AND invoice_inputs.id IN (
        SELECT invoice_materials.invoice_id
        FROM invoice_materials
        WHERE
            invoice_materials.project_id = $1
            AND invoice_materials.invoice_type = 'input'
            AND invoice_materials.material_cost_id IN (
                SELECT material_costs.id
                FROM material_costs
                WHERE material_costs.material_id = ANY(sqlc.arg(material_ids)::bigint[])
            )
    )
ORDER BY invoice_inputs.id DESC
LIMIT $2 OFFSET $3;

-- name: CountInvoiceInputsByMaterials :one
SELECT COUNT(*)::bigint
FROM invoice_inputs
LEFT JOIN workers AS warehouse_manager ON warehouse_manager.id = invoice_inputs.warehouse_manager_worker_id
LEFT JOIN workers AS released ON released.id = invoice_inputs.released_worker_id
WHERE
    invoice_inputs.project_id = $1
    AND invoice_inputs.id IN (
        SELECT invoice_materials.invoice_id
        FROM invoice_materials
        WHERE
            invoice_materials.project_id = $1
            AND invoice_materials.invoice_type = 'input'
            AND invoice_materials.material_cost_id IN (
                SELECT material_costs.id
                FROM material_costs
                WHERE material_costs.material_id = ANY(sqlc.arg(material_ids)::bigint[])
            )
    );

-- name: ListInvoiceInputUniqueDeliveryCodes :many
SELECT
    COALESCE(delivery_code, '')::text AS label,
    COALESCE(delivery_code, '')::text AS value
FROM invoice_inputs
WHERE project_id = $1
ORDER BY id DESC;

-- name: ListInvoiceInputUniqueWarehouseManagers :many
SELECT
    workers.id                       AS value,
    COALESCE(workers.name, '')::text AS label
FROM workers
WHERE workers.id IN (
    SELECT DISTINCT(invoice_inputs.warehouse_manager_worker_id)
    FROM invoice_inputs
    WHERE invoice_inputs.project_id = $1 AND invoice_inputs.warehouse_manager_worker_id IS NOT NULL
);

-- name: ListInvoiceInputUniqueReleasedWorkers :many
SELECT
    workers.id                       AS value,
    COALESCE(workers.name, '')::text AS label
FROM workers
WHERE workers.id IN (
    SELECT DISTINCT(invoice_inputs.released_worker_id)
    FROM invoice_inputs
    WHERE invoice_inputs.project_id = $1 AND invoice_inputs.released_worker_id IS NOT NULL
);

-- name: ListInvoiceInputReportFilterData :many
SELECT
    invoice_inputs.id                                AS id,
    COALESCE(warehouse_manager.name, '')::text       AS warehouse_manager_name,
    COALESCE(released.name, '')::text                AS released_name,
    COALESCE(invoice_inputs.delivery_code, '')::text AS delivery_code,
    COALESCE(invoice_inputs.notes, '')::text         AS notes,
    invoice_inputs.date_of_invoice                   AS date_of_invoice
FROM invoice_inputs
LEFT JOIN workers AS warehouse_manager ON warehouse_manager.id = invoice_inputs.warehouse_manager_worker_id
LEFT JOIN workers AS released ON released.id = invoice_inputs.released_worker_id
WHERE
    invoice_inputs.project_id = $1
    AND (NULLIF($2::text, '') IS NULL OR invoice_inputs.delivery_code = $2)
    AND (NULLIF($3::bigint, 0) IS NULL OR invoice_inputs.released_worker_id = $3)
    AND (NULLIF($4::bigint, 0) IS NULL OR invoice_inputs.warehouse_manager_worker_id = $4)
    AND (NULLIF($5::timestamptz, '0001-01-01 00:00:00 UTC') IS NULL OR $5 <= invoice_inputs.date_of_invoice)
    AND (NULLIF($6::timestamptz, '0001-01-01 00:00:00 UTC') IS NULL OR invoice_inputs.date_of_invoice <= $6)
ORDER BY invoice_inputs.id DESC;

-- name: ListInvoiceInputMaterialsForEdit :many
SELECT
    materials.id                                            AS material_id,
    COALESCE(materials.name, '')::text                      AS material_name,
    COALESCE(materials.unit, '')::text                      AS unit,
    COALESCE(invoice_materials.amount, 0)::numeric          AS amount,
    material_costs.id                                       AS material_cost_id,
    COALESCE(material_costs.cost_m19, 0)::numeric           AS material_cost,
    COALESCE(invoice_materials.notes, '')::text             AS notes,
    COALESCE(materials.has_serial_number, false)::boolean   AS has_serial_number
FROM invoice_materials
INNER JOIN material_costs ON invoice_materials.material_cost_id = material_costs.id
INNER JOIN materials ON material_costs.material_id = materials.id
WHERE
    invoice_materials.invoice_type = 'input'
    AND invoice_materials.invoice_id = $1
    AND invoice_materials.project_id = $2;

-- name: ListInvoiceInputAllDeliveryCodes :many
SELECT COALESCE(delivery_code, '')::text AS delivery_code
FROM invoice_inputs
WHERE project_id = $1;

-- name: ListInvoiceInputAllWarehouseManagers :many
SELECT DISTINCT
    workers.id                       AS value,
    COALESCE(workers.name, '')::text AS label
FROM invoice_inputs
INNER JOIN workers ON workers.id = invoice_inputs.warehouse_manager_worker_id
WHERE invoice_inputs.project_id = $1;

-- name: ListInvoiceInputAllReleasedWorkers :many
SELECT DISTINCT
    workers.id                       AS value,
    COALESCE(workers.name, '')::text AS label
FROM invoice_inputs
INNER JOIN workers ON workers.id = invoice_inputs.released_worker_id
WHERE invoice_inputs.project_id = $1;

-- name: ListInvoiceInputAllMaterialsThatExist :many
SELECT DISTINCT
    materials.id                       AS value,
    COALESCE(materials.name, '')::text AS label
FROM invoice_materials
INNER JOIN material_costs ON invoice_materials.material_cost_id = material_costs.id
INNER JOIN materials ON materials.id = material_costs.material_id
WHERE
    invoice_materials.project_id = $1
    AND invoice_materials.invoice_type = 'input';

-- name: UpsertMaterialLocationByID :exec
-- Used by invoice_input.Confirmation: when an existing material_locations
-- row is found for (project_id, material_cost_id, location_type='warehouse'),
-- its amount is updated to the supplied total. The GORM-era code did this
-- via clause.OnConflict on id; here it's an explicit UPDATE since the id
-- is already known at this point.
UPDATE material_locations
SET amount = sqlc.arg(amount)::numeric
WHERE id = sqlc.arg(id);
