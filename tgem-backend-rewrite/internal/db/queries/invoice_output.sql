-- name: ListInvoiceOutputs :many
SELECT id, district_id, project_id, warehouse_manager_worker_id, released_worker_id,
       recipient_worker_id, team_id, delivery_code, date_of_invoice, notes, confirmation
FROM invoice_outputs
ORDER BY id DESC;

-- name: ListInvoiceOutputsPaginated :many
SELECT id, district_id, project_id, warehouse_manager_worker_id, released_worker_id,
       recipient_worker_id, team_id, delivery_code, date_of_invoice, notes, confirmation
FROM invoice_outputs
ORDER BY id DESC
LIMIT $1 OFFSET $2;

-- name: GetInvoiceOutput :one
SELECT id, district_id, project_id, warehouse_manager_worker_id, released_worker_id,
       recipient_worker_id, team_id, delivery_code, date_of_invoice, notes, confirmation
FROM invoice_outputs
WHERE id = $1;

-- name: GetInvoiceOutputByDeliveryCode :one
SELECT id, district_id, project_id, warehouse_manager_worker_id, released_worker_id,
       recipient_worker_id, team_id, delivery_code, date_of_invoice, notes, confirmation
FROM invoice_outputs
WHERE delivery_code = $1
  AND project_id = $2
LIMIT 1;

-- name: CreateInvoiceOutput :one
INSERT INTO invoice_outputs (
    district_id, project_id, warehouse_manager_worker_id, released_worker_id,
    recipient_worker_id, team_id, delivery_code, date_of_invoice, notes, confirmation
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, district_id, project_id, warehouse_manager_worker_id, released_worker_id,
          recipient_worker_id, team_id, delivery_code, date_of_invoice, notes, confirmation;

-- name: UpdateInvoiceOutput :one
UPDATE invoice_outputs
SET district_id = $2, project_id = $3, warehouse_manager_worker_id = $4,
    released_worker_id = $5, recipient_worker_id = $6, team_id = $7,
    delivery_code = $8, date_of_invoice = $9, notes = $10, confirmation = $11
WHERE id = $1
RETURNING id, district_id, project_id, warehouse_manager_worker_id, released_worker_id,
          recipient_worker_id, team_id, delivery_code, date_of_invoice, notes, confirmation;

-- name: DeleteInvoiceOutput :exec
DELETE FROM invoice_outputs WHERE id = $1;

-- name: ConfirmInvoiceOutput :exec
UPDATE invoice_outputs
SET confirmation = true
WHERE id = $1;

-- name: CountInvoiceOutputsByProject :one
SELECT COUNT(*)::bigint FROM invoice_outputs WHERE project_id = $1;

-- name: ListInvoiceOutputsPaginatedFiltered :many
SELECT
    invoice_outputs.id                                  AS id,
    COALESCE(invoice_outputs.delivery_code, '')::text   AS delivery_code,
    COALESCE(districts.name, '')::text                  AS district_name,
    districts.id                                        AS district_id,
    COALESCE(teams.number, '')::text                    AS team_name,
    teams.id                                            AS team_id,
    COALESCE(warehouse_manager.id, 0)::bigint           AS warehouse_manager_id,
    COALESCE(warehouse_manager.name, '')::text          AS warehouse_manager_name,
    COALESCE(released.name, '')::text                   AS released_name,
    COALESCE(recipient.id, 0)::bigint                   AS recipient_id,
    COALESCE(recipient.name, '')::text                  AS recipient_name,
    invoice_outputs.date_of_invoice                     AS date_of_invoice,
    COALESCE(invoice_outputs.confirmation, false)::boolean AS confirmation,
    COALESCE(invoice_outputs.notes, '')::text           AS notes
FROM invoice_outputs
INNER JOIN districts ON districts.id = invoice_outputs.district_id
INNER JOIN teams ON teams.id = invoice_outputs.team_id
LEFT JOIN workers AS warehouse_manager ON warehouse_manager.id = invoice_outputs.warehouse_manager_worker_id
LEFT JOIN workers AS released ON released.id = invoice_outputs.released_worker_id
LEFT JOIN workers AS recipient ON recipient.id = invoice_outputs.recipient_worker_id
WHERE
    invoice_outputs.project_id = $1
    AND (NULLIF($2::bigint, 0) IS NULL OR invoice_outputs.district_id = $2)
    AND (NULLIF($3::bigint, 0) IS NULL OR invoice_outputs.warehouse_manager_worker_id = $3)
    AND (NULLIF($4::bigint, 0) IS NULL OR invoice_outputs.released_worker_id = $4)
    AND (NULLIF($5::bigint, 0) IS NULL OR invoice_outputs.recipient_worker_id = $5)
    AND (NULLIF($6::bigint, 0) IS NULL OR invoice_outputs.team_id = $6)
    AND (NULLIF($7::text, '') IS NULL OR invoice_outputs.delivery_code = $7)
ORDER BY invoice_outputs.id DESC
LIMIT $8 OFFSET $9;

-- name: ListInvoiceOutputUniqueCodes :many
SELECT
    COALESCE(delivery_code, '')::text AS label,
    COALESCE(delivery_code, '')::text AS value
FROM invoice_outputs
WHERE project_id = $1
ORDER BY id DESC;

-- name: ListInvoiceOutputUniqueWarehouseManagers :many
SELECT
    workers.id                       AS value,
    COALESCE(workers.name, '')::text AS label
FROM workers
WHERE workers.id IN (
    SELECT DISTINCT(invoice_outputs.warehouse_manager_worker_id)
    FROM invoice_outputs
    WHERE invoice_outputs.project_id = $1 AND invoice_outputs.warehouse_manager_worker_id IS NOT NULL
);

-- name: ListInvoiceOutputUniqueRecieved :many
SELECT
    workers.id                       AS value,
    COALESCE(workers.name, '')::text AS label
FROM workers
WHERE workers.id IN (
    SELECT DISTINCT(invoice_outputs.recipient_worker_id)
    FROM invoice_outputs
    WHERE invoice_outputs.project_id = $1 AND invoice_outputs.recipient_worker_id IS NOT NULL
);

-- name: ListInvoiceOutputUniqueDistricts :many
SELECT
    districts.id                       AS value,
    COALESCE(districts.name, '')::text AS label
FROM districts
WHERE districts.project_id = $1;

-- name: ListInvoiceOutputUniqueTeams :many
SELECT
    teams.id                                                                                  AS value,
    (COALESCE(teams.number, '') || ' (' || COALESCE(workers.name, '') || ')')::text           AS label
FROM teams
LEFT JOIN team_leaders ON team_leaders.team_id = teams.id
LEFT JOIN workers ON workers.id = team_leaders.leader_worker_id
WHERE teams.id IN (
    SELECT DISTINCT(invoice_outputs.team_id)
    FROM invoice_outputs
    WHERE invoice_outputs.project_id = $1 AND invoice_outputs.team_id IS NOT NULL
);

-- name: ListInvoiceOutputReportFilterData :many
SELECT
    invoice_outputs.id                                  AS id,
    COALESCE(invoice_outputs.delivery_code, '')::text   AS delivery_code,
    COALESCE(warehouse_manager.name, '')::text          AS warehouse_manager_name,
    COALESCE(recipient_worker.name, '')::text           AS recipient_name,
    COALESCE(teams.number, '')::text                    AS team_number,
    COALESCE(leader_worker.name, '')::text              AS team_leader_name,
    invoice_outputs.date_of_invoice                     AS date_of_invoice
FROM invoice_outputs
LEFT JOIN workers AS warehouse_manager ON warehouse_manager.id = invoice_outputs.warehouse_manager_worker_id
LEFT JOIN workers AS recipient_worker ON recipient_worker.id = invoice_outputs.recipient_worker_id
INNER JOIN teams ON teams.id = invoice_outputs.team_id
LEFT JOIN team_leaders ON teams.id = team_leaders.team_id
LEFT JOIN workers AS leader_worker ON leader_worker.id = team_leaders.leader_worker_id
WHERE
    invoice_outputs.project_id = $1
    AND COALESCE(invoice_outputs.confirmation, false) = true
    AND (NULLIF($2::text, '') IS NULL OR invoice_outputs.delivery_code = $2)
    AND (NULLIF($3::bigint, 0) IS NULL OR invoice_outputs.recipient_worker_id = $3)
    AND (NULLIF($4::bigint, 0) IS NULL OR invoice_outputs.warehouse_manager_worker_id = $4)
    AND (NULLIF($5::bigint, 0) IS NULL OR invoice_outputs.team_id = $5)
    AND (NULLIF($6::timestamptz, '0001-01-01 00:00:00 UTC') IS NULL OR $6 <= invoice_outputs.date_of_invoice)
    AND (NULLIF($7::timestamptz, '0001-01-01 00:00:00 UTC') IS NULL OR invoice_outputs.date_of_invoice <= $7)
ORDER BY invoice_outputs.id DESC;

-- name: ListInvoiceOutputAvailableMaterialsInWarehouse :many
SELECT
    materials.id                                                  AS id,
    COALESCE(materials.name, '')::text                            AS name,
    COALESCE(materials.unit, '')::text                            AS unit,
    COALESCE(materials.has_serial_number, false)::boolean         AS has_serial_number,
    COALESCE(material_locations.amount, 0)::numeric               AS amount
FROM material_locations
INNER JOIN material_costs ON material_costs.id = material_locations.material_cost_id
INNER JOIN materials ON materials.id = material_costs.material_id
WHERE
    material_locations.project_id = $1
    AND material_locations.location_type = 'warehouse'
    AND material_locations.amount > 0
ORDER BY materials.name;

-- name: ListInvoiceOutputMaterialsForEdit :many
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
    AND invoice_materials.invoice_type = 'output'
    AND invoice_materials.invoice_id = $1
    AND invoice_materials.project_id = $2;

-- name: ListInvoiceOutputMaterialDataForReport :many
SELECT
    materials.id                                        AS material_id,
    COALESCE(materials.name, '')::text                  AS material_name,
    COALESCE(materials.unit, '')::text                  AS material_unit,
    COALESCE(materials.code, '')::text                  AS material_code,
    COALESCE(material_costs.cost_m19, 0)::numeric       AS material_cost_m19,
    COALESCE(invoice_materials.notes, '')::text         AS notes,
    COALESCE(invoice_materials.amount, 0)::numeric      AS amount
FROM invoice_materials
INNER JOIN material_costs ON material_costs.id = invoice_materials.material_cost_id
INNER JOIN materials ON materials.id = material_costs.material_id
WHERE
    invoice_materials.invoice_type = 'output'
    AND invoice_materials.invoice_id = $1;

-- name: GetInvoiceOutputDataForExcel :one
SELECT
    invoice_outputs.id                                  AS id,
    COALESCE(projects.name, '')::text                   AS project_name,
    COALESCE(invoice_outputs.delivery_code, '')::text   AS delivery_code,
    COALESCE(districts.name, '')::text                  AS district_name,
    COALESCE(team_leader.name, '')::text                AS team_leader_name,
    COALESCE(warehouse_manager.name, '')::text          AS warehouse_manager_name,
    COALESCE(released.name, '')::text                   AS released_name,
    COALESCE(recipient.name, '')::text                  AS recipient_name,
    invoice_outputs.date_of_invoice                     AS date_of_invoice
FROM invoice_outputs
INNER JOIN projects ON projects.id = invoice_outputs.project_id
INNER JOIN districts ON districts.id = invoice_outputs.district_id
INNER JOIN teams ON teams.id = invoice_outputs.team_id
LEFT JOIN team_leaders ON team_leaders.team_id = teams.id
LEFT JOIN workers AS team_leader ON team_leader.id = team_leaders.leader_worker_id
LEFT JOIN workers AS warehouse_manager ON warehouse_manager.id = invoice_outputs.warehouse_manager_worker_id
LEFT JOIN workers AS released ON released.id = invoice_outputs.released_worker_id
LEFT JOIN workers AS recipient ON recipient.id = invoice_outputs.recipient_worker_id
WHERE invoice_outputs.id = $1
ORDER BY team_leaders.id DESC
LIMIT 1;

-- name: ListMaterialAmountSortedByCostM19InLocation :many
SELECT
    materials.id                                            AS material_id,
    material_costs.id                                       AS material_cost_id,
    COALESCE(material_costs.cost_m19, 0)::numeric           AS material_cost_m19,
    COALESCE(material_locations.amount, 0)::numeric         AS material_amount
FROM material_locations
INNER JOIN material_costs ON material_costs.id = material_locations.material_cost_id
INNER JOIN materials ON materials.id = material_costs.material_id
WHERE
    material_locations.project_id = $1
    AND material_locations.location_type = $2
    AND material_locations.location_id = $3
    AND materials.id = $4
    AND material_locations.amount > 0
ORDER BY material_costs.cost_m19 DESC;

-- name: ListMaterialCostIDAndSerialNumberIDByCodes :many
SELECT
    material_costs.id                                       AS material_cost_id,
    serial_numbers.id                                       AS serial_number_id,
    serial_number_locations.id                              AS serial_number_location_id
FROM serial_numbers
INNER JOIN material_costs ON material_costs.id = serial_numbers.material_cost_id
INNER JOIN materials ON materials.id = material_costs.material_id
INNER JOIN serial_number_locations ON serial_number_locations.serial_number_id = serial_numbers.id
WHERE
    materials.id = $1
    AND serial_number_locations.location_type = $2
    AND serial_number_locations.location_id = $3
    AND serial_numbers.code = ANY(sqlc.arg(codes)::text[])
ORDER BY material_costs.id;

-- name: ListSerialNumberCodesByMaterialID :many
SELECT COALESCE(serial_numbers.code, '')::text AS code
FROM materials
INNER JOIN material_costs ON material_costs.material_id = materials.id
INNER JOIN serial_numbers ON material_costs.id = serial_numbers.material_cost_id
INNER JOIN serial_number_locations ON serial_number_locations.serial_number_id = serial_numbers.id
WHERE
    materials.project_id = serial_numbers.project_id
    AND materials.project_id = serial_number_locations.project_id
    AND materials.project_id = $1
    AND materials.id = $2
    AND serial_number_locations.location_type = $3
    AND serial_number_locations.location_id = 0;

-- name: GetTotalAmountInWarehouse :one
SELECT COALESCE(SUM(material_locations.amount), 0)::numeric AS total_amount
FROM materials
INNER JOIN material_costs ON material_costs.material_id = materials.id
INNER JOIN material_locations ON material_locations.material_cost_id = material_costs.id
WHERE
    material_locations.project_id = $1
    AND material_locations.location_type = 'warehouse'
    AND materials.id = $2;

-- name: ConfirmSerialNumberLocationsByOutputInvoice :exec
-- After invoice_output confirmation, serial_number_locations rows for the
-- invoice's serial numbers are flipped from warehouse to team.
UPDATE serial_number_locations
SET location_type = 'team', location_id = sqlc.arg(team_id)::bigint
WHERE serial_number_locations.serial_number_id IN (
    SELECT serial_number_movements.serial_number_id
    FROM serial_number_movements
    WHERE
        serial_number_movements.invoice_type = 'output'
        AND serial_number_movements.invoice_id = sqlc.arg(invoice_id)::bigint
);
