-- name: ListInvoiceMaterialsByInvoice :many
SELECT id, project_id, material_cost_id, invoice_id, invoice_type, is_defected, amount, notes
FROM invoice_materials
WHERE invoice_id = $1 AND invoice_type = $2 AND project_id = $3;

-- name: GetInvoiceMaterialByMaterialCostID :one
SELECT id, project_id, material_cost_id, invoice_id, invoice_type, is_defected, amount, notes
FROM invoice_materials
WHERE material_cost_id = $1 AND invoice_type = $2 AND invoice_id = $3
LIMIT 1;

-- name: DeleteInvoiceMaterialsByInvoice :exec
DELETE FROM invoice_materials
WHERE invoice_type = $1 AND invoice_id = $2;

-- name: CreateInvoiceMaterialsBatch :copyfrom
INSERT INTO invoice_materials (project_id, material_cost_id, invoice_id, invoice_type, is_defected, amount, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: ListInvoiceMaterialsWithoutSerialNumbers :many
SELECT
    invoice_materials.id                       AS id,
    COALESCE(materials.name, '')::text         AS material_name,
    COALESCE(materials.unit, '')::text         AS material_unit,
    COALESCE(invoice_materials.is_defected, false)::boolean AS is_defected,
    COALESCE(material_costs.cost_m19, 0)::numeric           AS cost_m19,
    COALESCE(invoice_materials.amount, 0)::numeric          AS amount,
    COALESCE(invoice_materials.notes, '')::text             AS notes
FROM invoice_materials
INNER JOIN material_costs ON material_costs.id = invoice_materials.material_cost_id
INNER JOIN materials ON materials.id = material_costs.material_id
WHERE
    invoice_materials.invoice_type = $1
    AND invoice_materials.invoice_id = $2
    AND invoice_materials.project_id = $3
    AND COALESCE(materials.has_serial_number, false) = false
ORDER BY materials.name DESC;

-- name: ListInvoiceMaterialsWithSerialNumbers :many
SELECT
    invoice_materials.id                                    AS id,
    COALESCE(materials.name, '')::text                      AS material_name,
    COALESCE(materials.unit, '')::text                      AS material_unit,
    COALESCE(invoice_materials.is_defected, false)::boolean AS is_defected,
    COALESCE(material_costs.cost_m19, 0)::numeric           AS cost_m19,
    COALESCE(serial_numbers.code, '')::text                 AS serial_number,
    COALESCE(invoice_materials.amount, 0)::numeric          AS amount,
    COALESCE(invoice_materials.notes, '')::text             AS notes
FROM invoice_materials
INNER JOIN serial_number_movements ON serial_number_movements.invoice_id = invoice_materials.invoice_id
INNER JOIN serial_numbers ON serial_numbers.id = serial_number_movements.serial_number_id
INNER JOIN material_costs ON material_costs.id = serial_numbers.material_cost_id
INNER JOIN materials ON materials.id = material_costs.material_id
WHERE
    invoice_materials.project_id = serial_numbers.project_id
    AND invoice_materials.project_id = serial_number_movements.project_id
    AND invoice_materials.invoice_type = serial_number_movements.invoice_type
    AND invoice_materials.material_cost_id = serial_numbers.material_cost_id
    AND invoice_materials.is_defected = serial_number_movements.is_defected
    AND invoice_materials.invoice_type = $1
    AND invoice_materials.invoice_id = $2
    AND invoice_materials.project_id = $3
    AND COALESCE(materials.has_serial_number, false) = true
ORDER BY materials.name DESC;

-- name: ListInvoiceMaterialsDataForReport :many
SELECT
    invoice_materials.id                                AS invoice_material_id,
    materials.id                                        AS material_id,
    COALESCE(materials.name, '')::text                  AS material_name,
    COALESCE(materials.unit, '')::text                  AS material_unit,
    COALESCE(materials.category, '')::text              AS material_category,
    COALESCE(material_costs.cost_prime, 0)::numeric     AS material_cost_prime,
    COALESCE(material_costs.cost_m19, 0)::numeric       AS material_cost_m19,
    COALESCE(material_costs.cost_with_customer, 0)::numeric AS material_cost_with_customer,
    COALESCE(invoice_materials.amount, 0)::numeric      AS invoice_material_amount,
    COALESCE(invoice_materials.notes, '')::text         AS invoice_material_notes
FROM invoice_materials
INNER JOIN material_costs ON material_costs.id = invoice_materials.material_cost_id
INNER JOIN materials ON materials.id = material_costs.material_id
WHERE
    invoice_type = $1
    AND invoice_id = $2;
