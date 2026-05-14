-- name: ListInvoiceCorrectionsPaginated :many
SELECT
    io.id                                                       AS id,
    COALESCE(w.name, '')::text                                  AS supervisor_name,
    io.district_id                                              AS district_id,
    COALESCE(d.name, '')::text                                  AS district_name,
    COALESCE(o.name, '')::text                                  AS object_name,
    COALESCE(o.type, '')::text                                  AS object_type,
    io.team_id                                                  AS team_id,
    io.date_of_invoice                                          AS date_of_invoice,
    COALESCE(io.delivery_code, '')::text                        AS delivery_code,
    COALESCE(io.confirmed_by_operator, false)::boolean          AS confirmed_by_operator,
    COALESCE(w2.name, '')::text                                 AS team_leader_name
FROM invoice_objects AS io
INNER JOIN workers AS w ON w.id = io.supervisor_worker_id
LEFT JOIN districts AS d ON d.id = io.district_id
INNER JOIN objects AS o ON o.id = io.object_id
INNER JOIN team_leaders AS tl ON tl.team_id = io.team_id
INNER JOIN workers AS w2 ON w2.id = tl.leader_worker_id
WHERE
    io.project_id = $1
    AND COALESCE(io.confirmed_by_operator, false) = false
    AND ($2::bigint = 0 OR io.team_id = $2)
    AND ($3::bigint = 0 OR io.object_id = $3)
ORDER BY io.id DESC
LIMIT $4 OFFSET $5;

-- name: CountInvoiceCorrections :one
SELECT COUNT(*)::bigint
FROM invoice_objects
WHERE
    COALESCE(confirmed_by_operator, false) = false
    AND project_id = $1
    AND ($2::bigint = 0 OR team_id = $2)
    AND ($3::bigint = 0 OR object_id = $3);

-- name: ListInvoiceCorrectionMaterialsByInvoiceObjectID :many
SELECT
    invoice_materials.id                                            AS invoice_material_id,
    COALESCE(materials.name, '')::text                              AS material_name,
    materials.id                                                    AS material_id,
    COALESCE(invoice_materials.notes, '')::text                     AS notes,
    COALESCE(invoice_materials.amount, 0)::numeric                  AS material_amount
FROM invoice_materials
INNER JOIN material_costs ON material_costs.id = invoice_materials.material_cost_id
INNER JOIN materials ON materials.id = material_costs.material_id
WHERE
    invoice_materials.invoice_type = 'object'
    AND invoice_materials.invoice_id = $1
ORDER BY materials.id;

-- name: ListInvoiceCorrectionSerialNumberOfMaterialInTeam :many
-- The GORM repo's query had a typo (`team.id` instead of `teams.id`)
-- and would have raised "missing FROM-clause entry" at execution time.
-- Fixed in place since preserving would mean this endpoint always errors.
SELECT COALESCE(serial_numbers.code, '')::text AS code
FROM material_locations
INNER JOIN teams ON teams.id = material_locations.location_id
INNER JOIN serial_numbers ON serial_numbers.material_cost_id = material_locations.material_cost_id
INNER JOIN serial_number_locations ON serial_number_locations.serial_number_id = serial_numbers.id
INNER JOIN material_costs ON material_locations.material_cost_id = material_costs.id
INNER JOIN materials ON materials.id = material_costs.material_id
WHERE
    materials.project_id = $1
    AND materials.id = $2
    AND teams.id = $3
    AND material_locations.location_type = serial_number_locations.location_type
    AND material_locations.location_id = serial_number_locations.location_id;

-- name: ListInvoiceCorrectionUniqueObjects :many
SELECT
    objects.id                              AS id,
    COALESCE(objects.name, '')::text        AS object_name,
    COALESCE(objects.type, '')::text        AS object_type
FROM objects
WHERE objects.id IN (
    SELECT DISTINCT(invoice_objects.object_id)
    FROM invoice_objects
    WHERE invoice_objects.project_id = $1 AND invoice_objects.object_id IS NOT NULL
);

-- name: ListInvoiceCorrectionUniqueTeams :many
SELECT
    teams.id                                                                                  AS value,
    (COALESCE(teams.number, '') || ' (' || COALESCE(workers.name, '') || ')')::text           AS label
FROM teams
INNER JOIN team_leaders ON team_leaders.team_id = teams.id
INNER JOIN workers ON workers.id = team_leaders.leader_worker_id
WHERE teams.id IN (
    SELECT DISTINCT(invoice_objects.team_id)
    FROM invoice_objects
    WHERE invoice_objects.project_id = $1 AND invoice_objects.team_id IS NOT NULL
);

-- name: ListInvoiceCorrectionReportData :many
SELECT
    invoice_objects.id                                          AS id,
    COALESCE(invoice_objects.delivery_code, '')::text           AS delivery_code,
    COALESCE(districts.name, '')::text                          AS district_name,
    COALESCE(objects.name, '')::text                            AS object_name,
    COALESCE(objects.type, '')::text                            AS object_type,
    COALESCE(teams.number, '')::text                            AS team_number,
    COALESCE(team_leader.name, '')::text                        AS team_leader_name,
    invoice_objects.date_of_invoice                             AS date_of_invoice,
    COALESCE(operator.name, '')::text                           AS operator_name,
    invoice_objects.date_of_correction                          AS date_of_correction
FROM invoice_objects
LEFT JOIN districts ON districts.id = invoice_objects.district_id
INNER JOIN objects ON objects.id = invoice_objects.object_id
INNER JOIN teams ON teams.id = invoice_objects.team_id
INNER JOIN team_leaders ON team_leaders.team_id = teams.id
INNER JOIN workers AS team_leader ON team_leader.id = team_leaders.leader_worker_id
INNER JOIN invoice_object_operators ON invoice_object_operators.invoice_object_id = invoice_objects.id
INNER JOIN workers AS operator ON operator.id = invoice_object_operators.operator_worker_id
WHERE
    invoice_objects.project_id = $1
    AND COALESCE(invoice_objects.confirmed_by_operator, false) = true
    AND (NULLIF($2::bigint, 0) IS NULL OR invoice_objects.object_id = $2)
    AND (NULLIF($3::bigint, 0) IS NULL OR invoice_objects.team_id = $3)
    AND (NULLIF($4::timestamptz, '0001-01-01 00:00:00 UTC') IS NULL OR $4 <= invoice_objects.date_of_invoice)
    AND (NULLIF($5::timestamptz, '0001-01-01 00:00:00 UTC') IS NULL OR invoice_objects.date_of_invoice <= $5)
ORDER BY invoice_objects.id DESC;

-- name: ListInvoiceCorrectionOperationsByInvoiceObjectID :many
SELECT
    operations.id                                          AS operation_id,
    COALESCE(operations.name, '')::text                    AS operation_name,
    COALESCE(invoice_operations.amount, 0)::numeric        AS amount,
    COALESCE(materials.name, '')::text                     AS material_name
FROM invoice_operations
INNER JOIN operations ON operations.id = invoice_operations.operation_id
FULL JOIN operation_materials ON operation_materials.operation_id = operations.id
FULL JOIN materials ON materials.id = operation_materials.material_id
WHERE
    invoice_operations.invoice_id = $1
    AND invoice_operations.invoice_type = 'object';

-- name: ListInvoiceCorrectionTeamsForSearch :many
SELECT
    (COALESCE(w.name, '') || ' (' || COALESCE(t.number, '') || ')')::text AS label,
    t.id                                                                  AS value
FROM teams t
INNER JOIN team_leaders t2 ON t2.team_id = t.id
INNER JOIN workers w ON w.id = t2.leader_worker_id
WHERE t.id IN (
    SELECT DISTINCT team_id
    FROM invoice_objects i
    WHERE i.project_id = $1 AND i.team_id IS NOT NULL
);

-- name: ListInvoiceCorrectionObjectsForSearch :many
SELECT
    COALESCE(o.name, '')::text   AS label,
    o.id                          AS value
FROM objects o
WHERE o.id IN (
    SELECT DISTINCT object_id
    FROM invoice_objects i
    WHERE i.project_id = $1 AND i.object_id IS NOT NULL
);

-- name: UpdateInvoiceObjectConfirmation :exec
UPDATE invoice_objects
SET confirmed_by_operator = $2,
    date_of_correction = $3
WHERE id = $1;

-- name: CreateInvoiceObjectOperator :exec
INSERT INTO invoice_object_operators (operator_worker_id, invoice_object_id)
VALUES ($1, $2);

-- name: GetMaterialLocationByCostAndLocation :one
-- Used by invoice_correction's GetByMaterialCostIDOrCreate idiom: looks up
-- a material_locations row matching all four scope columns.
SELECT id, project_id, material_cost_id, location_id, location_type, amount
FROM material_locations
WHERE
    project_id = $1
    AND material_cost_id = $2
    AND location_type = $3
    AND location_id = $4
LIMIT 1;

-- name: GetTotalAmountInTeamsByTeamNumber :one
SELECT COALESCE(SUM(material_locations.amount), 0)::numeric AS total_amount
FROM material_locations
INNER JOIN teams ON material_locations.location_id = teams.id
INNER JOIN material_costs ON material_costs.id = material_locations.material_cost_id
INNER JOIN materials ON materials.id = material_costs.material_id
WHERE
    materials.project_id = $1
    AND materials.id = $2
    AND material_locations.location_type = 'team'
    AND teams.number = $3;
