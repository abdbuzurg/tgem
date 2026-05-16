-- name: ListInvoiceReturns :many
SELECT id, project_id, district_id, returner_type, returner_id,
       acceptor_type, acceptor_id, accepted_by_worker_id, date_of_invoice,
       notes, delivery_code, confirmation
FROM invoice_returns
ORDER BY id DESC;

-- name: GetInvoiceReturn :one
SELECT id, project_id, district_id, returner_type, returner_id,
       acceptor_type, acceptor_id, accepted_by_worker_id, date_of_invoice,
       notes, delivery_code, confirmation
FROM invoice_returns
WHERE id = $1;

-- name: GetInvoiceReturnByDeliveryCode :one
SELECT id, project_id, district_id, returner_type, returner_id,
       acceptor_type, acceptor_id, accepted_by_worker_id, date_of_invoice,
       notes, delivery_code, confirmation
FROM invoice_returns
WHERE delivery_code = $1
  AND project_id = $2
LIMIT 1;

-- name: CreateInvoiceReturn :one
INSERT INTO invoice_returns (
    project_id, district_id, returner_type, returner_id,
    acceptor_type, acceptor_id, accepted_by_worker_id, date_of_invoice,
    notes, delivery_code, confirmation
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, project_id, district_id, returner_type, returner_id,
          acceptor_type, acceptor_id, accepted_by_worker_id, date_of_invoice,
          notes, delivery_code, confirmation;

-- name: UpdateInvoiceReturn :one
UPDATE invoice_returns
SET project_id = $2, district_id = $3, returner_type = $4, returner_id = $5,
    acceptor_type = $6, acceptor_id = $7, accepted_by_worker_id = $8,
    date_of_invoice = $9, notes = $10, delivery_code = $11, confirmation = $12
WHERE id = $1
RETURNING id, project_id, district_id, returner_type, returner_id,
          acceptor_type, acceptor_id, accepted_by_worker_id, date_of_invoice,
          notes, delivery_code, confirmation;

-- name: DeleteInvoiceReturn :exec
DELETE FROM invoice_returns WHERE id = $1;

-- name: ConfirmInvoiceReturn :exec
UPDATE invoice_returns
SET confirmation = true
WHERE id = $1;

-- name: CountInvoiceReturnsBasedOnType :one
SELECT COUNT(*)::bigint
FROM invoice_returns
WHERE project_id = $1 AND returner_type = $2;

-- name: CountInvoiceReturnsByProject :one
SELECT COUNT(*)::bigint FROM invoice_returns WHERE project_id = $1;

-- name: ListInvoiceReturnsPaginatedTeam :many
SELECT
    invoice_returns.id                                  AS id,
    COALESCE(invoice_returns.delivery_code, '')::text   AS delivery_code,
    COALESCE(districts.name, '')::text                  AS district_name,
    COALESCE(teams.number, '')::text                    AS team_number,
    COALESCE(workers.name, '')::text                    AS team_leader_name,
    COALESCE(acceptor_worker.name, '')::text            AS acceptor_name,
    invoice_returns.date_of_invoice                     AS date_of_invoice,
    COALESCE(invoice_returns.confirmation, false)::boolean AS confirmation
FROM invoice_returns
INNER JOIN districts ON districts.id = invoice_returns.district_id
INNER JOIN teams ON teams.id = invoice_returns.returner_id
LEFT JOIN team_leaders ON team_leaders.team_id = teams.id
LEFT JOIN workers ON workers.id = team_leaders.leader_worker_id
LEFT JOIN workers AS acceptor_worker ON acceptor_worker.id = invoice_returns.accepted_by_worker_id
WHERE
    invoice_returns.project_id = $1
    AND invoice_returns.returner_type = 'team'
ORDER BY invoice_returns.id DESC
LIMIT $2 OFFSET $3;

-- name: ListInvoiceReturnsPaginatedObject :many
SELECT
    invoice_returns.id                                  AS id,
    COALESCE(invoice_returns.delivery_code, '')::text   AS delivery_code,
    COALESCE(districts.name, '')::text                  AS district_name,
    COALESCE(objects.name, '')::text                    AS object_name,
    COALESCE(objects.type, '')::text                    AS object_type,
    COALESCE(workers.name, '')::text                    AS object_supervisor_name,
    COALESCE(teams.number, '')::text                    AS team_number,
    COALESCE(leader.name, '')::text                     AS team_leader_name,
    COALESCE(acceptor_worker.name, '')::text            AS acceptor_name,
    invoice_returns.date_of_invoice                     AS date_of_invoice,
    COALESCE(invoice_returns.confirmation, false)::boolean AS confirmation
FROM invoice_returns
INNER JOIN districts ON districts.id = invoice_returns.district_id
INNER JOIN objects ON objects.id = invoice_returns.returner_id
LEFT JOIN object_supervisors ON objects.id = object_supervisors.object_id
LEFT JOIN workers ON workers.id = object_supervisors.supervisor_worker_id
INNER JOIN teams ON teams.id = invoice_returns.acceptor_id
LEFT JOIN team_leaders ON team_leaders.team_id = teams.id
LEFT JOIN workers AS leader ON leader.id = team_leaders.leader_worker_id
LEFT JOIN workers AS acceptor_worker ON acceptor_worker.id = invoice_returns.accepted_by_worker_id
WHERE
    invoice_returns.project_id = $1
    AND invoice_returns.returner_type = 'object'
ORDER BY invoice_returns.id DESC
LIMIT $2 OFFSET $3;

-- name: ListInvoiceReturnUniqueCodes :many
SELECT DISTINCT(COALESCE(delivery_code, '')::text)
FROM invoice_returns
WHERE project_id = $1;

-- name: ListInvoiceReturnUniqueTeams :many
SELECT DISTINCT(COALESCE(returner_id, 0)::bigint)
FROM invoice_returns
WHERE returner_type = 'team' AND project_id = $1 AND returner_id IS NOT NULL;

-- name: ListInvoiceReturnUniqueObjects :many
SELECT DISTINCT(COALESCE(returner_id, 0)::bigint)
FROM invoice_returns
WHERE returner_type = 'object' AND project_id = $1 AND returner_id IS NOT NULL;

-- name: ListInvoiceReturnReportData :many
SELECT id, project_id, district_id, returner_type, returner_id,
       acceptor_type, acceptor_id, accepted_by_worker_id, date_of_invoice,
       notes, delivery_code, confirmation
FROM invoice_returns
WHERE
    project_id = $1
    AND (NULLIF($2::text, '') IS NULL OR delivery_code = $2)
    AND (NULLIF($3::text, '') IS NULL OR returner_type = $3)
    AND (NULLIF($4::bigint, 0) IS NULL OR returner_id = $4)
    AND (NULLIF($5::timestamptz, '0001-01-01 00:00:00 UTC') IS NULL OR $5 <= date_of_invoice)
    AND (NULLIF($6::timestamptz, '0001-01-01 00:00:00 UTC') IS NULL OR date_of_invoice <= $6)
ORDER BY id DESC;

-- name: ListInvoiceReturnMaterialsForEdit :many
SELECT
    materials.id                                            AS material_id,
    COALESCE(materials.name, '')::text                      AS material_name,
    COALESCE(materials.unit, '')::text                      AS unit,
    COALESCE(invoice_materials.amount, 0)::numeric          AS amount,
    COALESCE(invoice_materials.notes, '')::text             AS notes,
    COALESCE(materials.has_serial_number, false)::boolean   AS has_serial_number,
    COALESCE(invoice_materials.is_defected, false)::boolean AS is_defective,
    COALESCE(material_locations.amount, 0)::numeric         AS holder_amount
FROM invoice_materials
INNER JOIN material_costs ON invoice_materials.material_cost_id = material_costs.id
INNER JOIN materials ON material_costs.material_id = materials.id
INNER JOIN material_locations ON material_locations.material_cost_id = invoice_materials.material_cost_id
WHERE
    invoice_materials.invoice_type = 'return'
    AND invoice_materials.invoice_id = $1
    AND material_locations.location_type = $2
    AND material_locations.location_id = $3
    AND invoice_materials.project_id = $4
ORDER BY materials.id;

-- name: ListMaterialAmountReverseSortedByCostM19InLocation :many
-- The Asc-by-cost variant used by invoice_return Create/Update (returner
-- gives back the lowest-cost material first).
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
ORDER BY material_costs.cost_m19;

-- name: GetMaterialDefectByMaterialLocationID :one
SELECT id, amount, material_location_id
FROM material_defects
WHERE material_location_id = $1
LIMIT 1;

-- name: UpsertMaterialDefectByID :exec
UPDATE material_defects
SET amount = sqlc.arg(amount)::numeric,
    material_location_id = sqlc.arg(material_location_id)::bigint
WHERE id = sqlc.arg(id);

-- name: CreateMaterialDefect :one
INSERT INTO material_defects (amount, material_location_id)
VALUES ($1, $2)
RETURNING id, amount, material_location_id;

-- name: ConfirmSerialNumberLocationsByReturnInvoice :exec
-- After invoice_return confirmation, serial_number_locations rows for the
-- invoice's serial numbers are flipped to the team acceptor.
UPDATE serial_number_locations
SET location_type = 'team', location_id = sqlc.arg(acceptor_id)::bigint
WHERE serial_number_locations.serial_number_id IN (
    SELECT serial_number_movements.serial_number_id
    FROM serial_number_movements
    WHERE
        serial_number_movements.invoice_type = 'return'
        AND serial_number_movements.invoice_id = sqlc.arg(invoice_id)::bigint
);
