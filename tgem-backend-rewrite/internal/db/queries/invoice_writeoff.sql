-- name: ListInvoiceWriteOffs :many
SELECT id, project_id, released_worker_id, write_off_type, write_off_location_id,
       delivery_code, date_of_invoice, confirmation, date_of_confirmation, notes
FROM invoice_write_offs
ORDER BY id DESC;

-- name: GetInvoiceWriteOff :one
SELECT id, project_id, released_worker_id, write_off_type, write_off_location_id,
       delivery_code, date_of_invoice, confirmation, date_of_confirmation, notes
FROM invoice_write_offs
WHERE id = $1;

-- name: CreateInvoiceWriteOff :one
INSERT INTO invoice_write_offs (
    project_id, released_worker_id, write_off_type, write_off_location_id,
    delivery_code, date_of_invoice, confirmation, date_of_confirmation, notes
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, project_id, released_worker_id, write_off_type, write_off_location_id,
          delivery_code, date_of_invoice, confirmation, date_of_confirmation, notes;

-- name: UpdateInvoiceWriteOff :one
UPDATE invoice_write_offs
SET project_id = $2, released_worker_id = $3, write_off_type = $4,
    write_off_location_id = $5, delivery_code = $6, date_of_invoice = $7,
    confirmation = $8, date_of_confirmation = $9, notes = $10
WHERE id = $1
RETURNING id, project_id, released_worker_id, write_off_type, write_off_location_id,
          delivery_code, date_of_invoice, confirmation, date_of_confirmation, notes;

-- name: DeleteInvoiceWriteOff :exec
DELETE FROM invoice_write_offs WHERE id = $1;

-- name: ListInvoiceWriteOffsPaginated :many
SELECT
    invoice_write_offs.id                                       AS id,
    COALESCE(invoice_write_offs.write_off_type, '')::text       AS write_off_type,
    COALESCE(invoice_write_offs.write_off_location_id, 0)::bigint AS write_off_location_id,
    invoice_write_offs.released_worker_id                       AS released_worker_id,
    COALESCE(workers.name, '')::text                            AS released_worker_name,
    COALESCE(invoice_write_offs.delivery_code, '')::text        AS delivery_code,
    invoice_write_offs.date_of_invoice                          AS date_of_invoice,
    COALESCE(invoice_write_offs.confirmation, false)::boolean   AS confirmation,
    invoice_write_offs.date_of_confirmation                     AS date_of_confirmation
FROM invoice_write_offs
INNER JOIN workers ON workers.id = invoice_write_offs.released_worker_id
WHERE
    invoice_write_offs.project_id = $1
    AND (NULLIF($2::text, '') IS NULL OR invoice_write_offs.write_off_type = $2)
ORDER BY invoice_write_offs.id DESC
LIMIT $3 OFFSET $4;

-- name: CountInvoiceWriteOffsFiltered :one
SELECT COUNT(*)::bigint
FROM invoice_write_offs
WHERE
    invoice_write_offs.project_id = $1
    AND (NULLIF($2::text, '') IS NULL OR invoice_write_offs.write_off_type = $2);

-- name: ListInvoiceWriteOffMaterialsForEdit :many
SELECT
    materials.id                                                  AS material_id,
    COALESCE(materials.name, '')::text                            AS material_name,
    COALESCE(materials.unit, '')::text                            AS unit,
    COALESCE(invoice_materials.amount, 0)::numeric                AS amount,
    material_costs.id                                             AS material_cost_id,
    COALESCE(material_costs.cost_m19, 0)::numeric                 AS material_cost,
    COALESCE(invoice_materials.notes, '')::text                   AS notes,
    COALESCE(materials.has_serial_number, false)::boolean         AS has_serial_number,
    COALESCE(material_locations.amount, 0)::numeric               AS location_amount
FROM invoice_materials
INNER JOIN material_costs ON invoice_materials.material_cost_id = material_costs.id
INNER JOIN materials ON material_costs.material_id = materials.id
INNER JOIN material_locations ON material_locations.material_cost_id = invoice_materials.material_cost_id
WHERE
    invoice_materials.invoice_type = 'writeoff'
    AND invoice_materials.invoice_id = $1
    AND material_locations.location_type = $2
    AND material_locations.location_id = $3
    AND invoice_materials.project_id = $4
ORDER BY materials.id;

-- name: ListInvoiceWriteOffReportData :many
SELECT
    invoice_write_offs.id                                       AS id,
    COALESCE(invoice_write_offs.delivery_code, '')::text        AS delivery_code,
    COALESCE(released.name, '')::text                           AS released_worker_name,
    invoice_write_offs.date_of_invoice                          AS date_of_invoice
FROM invoice_write_offs
INNER JOIN workers AS released ON released.id = invoice_write_offs.released_worker_id
WHERE
    invoice_write_offs.project_id = $1
    AND invoice_write_offs.write_off_type = $2
    AND COALESCE(invoice_write_offs.confirmation, false) = true
    AND (NULLIF($3::timestamptz, '0001-01-01 00:00:00 UTC') IS NULL OR $3 <= invoice_write_offs.date_of_invoice)
    AND (NULLIF($4::timestamptz, '0001-01-01 00:00:00 UTC') IS NULL OR invoice_write_offs.date_of_invoice <= $4)
ORDER BY invoice_write_offs.id DESC;

-- name: UpdateInvoiceWriteOffConfirmation :exec
UPDATE invoice_write_offs
SET confirmation = $2, date_of_confirmation = $3
WHERE id = $1;

-- name: ListMaterialLocationsByLocationType :many
-- Used by invoice_writeoff Confirmation: gets all material_locations rows
-- for a given polymorphic location_type (e.g. 'writeoff-warehouse',
-- 'loss-team', etc — these are the writeoff sentinel location types).
SELECT id, project_id, material_cost_id, location_id, location_type, amount
FROM material_locations
WHERE location_type = $1;
