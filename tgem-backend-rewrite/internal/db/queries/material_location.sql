-- name: ListMaterialLocations :many
SELECT id, project_id, material_cost_id, location_id, location_type, amount
FROM material_locations
ORDER BY id DESC;

-- name: ListMaterialLocationsPaginated :many
SELECT id, project_id, material_cost_id, location_id, location_type, amount
FROM material_locations
ORDER BY id DESC
LIMIT $1 OFFSET $2;

-- name: ListMaterialLocationsPaginatedFiltered :many
-- Fixed during phase 6 from the GORM-era SQL that targeted the wrong
-- table (`SELECT * FROM materials`). This query now correctly hits
-- material_locations.
SELECT id, project_id, material_cost_id, location_id, location_type, amount
FROM material_locations
WHERE
    (NULLIF($1::bigint, 0) IS NULL OR material_cost_id = $1)
    AND (NULLIF($2::bigint, 0) IS NULL OR location_id = $2)
    AND (NULLIF($3::text, '') IS NULL OR location_type = $3)
    AND (NULLIF($4::numeric, 0) IS NULL OR amount = $4)
ORDER BY id DESC
LIMIT $5 OFFSET $6;

-- name: GetMaterialLocation :one
SELECT id, project_id, material_cost_id, location_id, location_type, amount
FROM material_locations
WHERE id = $1;

-- name: CreateMaterialLocation :one
INSERT INTO material_locations (project_id, material_cost_id, location_id, location_type, amount)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, project_id, material_cost_id, location_id, location_type, amount;

-- name: UpdateMaterialLocation :exec
UPDATE material_locations
SET project_id = $2, material_cost_id = $3, location_id = $4, location_type = $5, amount = $6
WHERE id = $1;

-- name: DeleteMaterialLocation :exec
DELETE FROM material_locations WHERE id = $1;

-- name: CountMaterialLocations :one
SELECT COUNT(*)::bigint FROM material_locations;

-- name: ListUniqueObjectsForSelectFromMaterialLocations :many
SELECT
    objects.id                       AS id,
    COALESCE(objects.name, '')::text AS object_name,
    COALESCE(objects.type, '')::text AS object_type
FROM objects
WHERE objects.id IN (
    SELECT DISTINCT location_id
    FROM material_locations
    WHERE
        location_type = 'object'
        AND amount > 0
        AND material_locations.project_id = $1
        AND location_id IS NOT NULL
);

-- name: ListUniqueTeamsForSelectFromMaterialLocations :many
SELECT
    teams.id                          AS id,
    COALESCE(teams.number, '')::text  AS team_number,
    COALESCE(workers.name, '')::text  AS team_leader_name
FROM teams
INNER JOIN team_leaders ON team_leaders.team_id = teams.id
INNER JOIN workers ON team_leaders.leader_worker_id = workers.id
WHERE teams.id IN (
    SELECT DISTINCT location_id
    FROM material_locations
    WHERE
        location_type = 'team'
        AND amount > 0
        AND material_locations.project_id = $1
        AND location_id IS NOT NULL
);

-- name: ListUniqueMaterialsFromLocation :many
-- COALESCE on location_id: legacy warehouse rows can land with location_id IS NULL
-- while canonical writes use 0. Normalize so NULL is treated as 0 — matches the
-- semantics of ListBalanceReportData. Without this, materials sitting at
-- (warehouse, NULL) appear in balance reports but are unselectable in the picker.
SELECT DISTINCT
    materials.id, materials.category, materials.code, materials.name,
    materials.unit, materials.notes, materials.has_serial_number,
    materials.article, materials.project_id,
    materials.planned_amount_for_project, materials.show_planned_amount_in_report
FROM material_locations
INNER JOIN material_costs ON material_costs.id = material_locations.material_cost_id
INNER JOIN materials ON materials.id = material_costs.material_id
WHERE
    material_locations.project_id = $1
    AND COALESCE(material_locations.location_id, 0) = $2
    AND material_locations.location_type = $3
    AND material_locations.amount > 0
ORDER BY materials.id;

-- name: ListUniqueMaterialCostsFromLocation :many
SELECT DISTINCT
    material_costs.id, material_costs.material_id,
    material_costs.cost_prime, material_costs.cost_m19, material_costs.cost_with_customer
FROM material_locations
INNER JOIN material_costs ON material_costs.id = material_locations.material_cost_id
INNER JOIN materials ON materials.id = material_costs.material_id
WHERE
    material_locations.project_id = $1
    AND material_locations.location_type = $2
    AND COALESCE(material_locations.location_id, 0) = $3
    AND materials.id = $4;

-- name: GetUniqueMaterialTotalAmount :one
SELECT COALESCE(material_locations.amount, 0)::numeric AS amount
FROM material_locations
WHERE
    material_locations.project_id = $1
    AND material_locations.location_type = $2
    AND COALESCE(material_locations.location_id, 0) = $3
    AND material_locations.material_cost_id = $4
LIMIT 1;

-- name: ListBalanceReportData :many
SELECT
    COALESCE(material_locations.location_id, 0)::bigint    AS location_id,
    materials.id                                           AS material_id,
    COALESCE(materials.code, '')::text                     AS material_code,
    COALESCE(materials.name, '')::text                     AS material_name,
    COALESCE(materials.unit, '')::text                     AS material_unit,
    COALESCE(material_locations.amount, 0)::numeric        AS total_amount,
    COALESCE(material_defects.amount, 0)::numeric          AS defect_amount,
    COALESCE(material_costs.cost_m19, 0)::numeric          AS material_cost_m19,
    (COALESCE(material_locations.amount, 0) * COALESCE(material_costs.cost_m19, 0))::numeric AS total_cost,
    (COALESCE(material_defects.amount, 0)  * COALESCE(material_costs.cost_m19, 0))::numeric AS total_defect_cost
FROM material_locations
INNER JOIN material_costs ON material_costs.id = material_locations.material_cost_id
INNER JOIN materials ON materials.id = material_costs.material_id
LEFT JOIN material_defects ON material_defects.material_location_id = material_locations.id
WHERE
    material_locations.project_id = $1
    AND material_locations.location_type = $2
    AND (NULLIF($3::bigint, 0) IS NULL OR material_locations.location_id = $3)
    AND material_locations.amount <> 0
ORDER BY material_locations.id;

-- name: GetTotalAmountInLocation :one
-- Used by invoice_return.GetMaterialAmountByMaterialID — sums material_locations.amount
-- across all material_costs of the given material in the given location, where amount > 0.
SELECT COALESCE(SUM(material_locations.amount), 0)::numeric AS total_amount
FROM materials
INNER JOIN material_costs ON material_costs.material_id = materials.id
INNER JOIN material_locations ON material_locations.material_cost_id = material_costs.id
WHERE
    materials.project_id = $1
    AND materials.id = $2
    AND material_locations.location_type = $3
    AND COALESCE(material_locations.location_id, 0) = $4
    AND material_locations.amount > 0;

-- name: ListMaterialLocationsForInvoiceConfirmation :many
-- The GORM-era materialLocationRepo.GetMaterialsInLocationBasedOnInvoiceID
-- query: rows in a (location_type, location_id) location whose material_cost_id
-- matches one of the invoice's invoice_materials rows.
SELECT id, project_id, material_cost_id, location_id, location_type, amount
FROM material_locations
WHERE
    material_locations.location_type = $1
    AND COALESCE(material_locations.location_id, 0) = $2
    AND material_locations.material_cost_id IN (
        SELECT invoice_materials.material_cost_id
        FROM invoice_materials
        WHERE
            invoice_materials.invoice_type = $3
            AND invoice_materials.invoice_id = $4
    );

-- name: ListMaterialLocationsLive :many
SELECT
    materials.id                                            AS material_id,
    COALESCE(materials.name, '')::text                      AS material_name,
    COALESCE(materials.unit, '')::text                      AS material_unit,
    material_costs.id                                       AS material_cost_id,
    COALESCE(material_costs.cost_m19, 0)::numeric           AS material_cost_m19,
    COALESCE(material_locations.location_type, '')::text    AS location_type,
    COALESCE(material_locations.location_id, 0)::bigint     AS location_id,
    COALESCE(material_locations.amount, 0)::numeric         AS amount
FROM material_locations
INNER JOIN material_costs ON material_costs.id = material_locations.material_cost_id
INNER JOIN materials ON materials.id = material_costs.material_id
WHERE
    material_locations.location_type = $1
    AND material_locations.project_id = $2
    AND (NULLIF($3::bigint, 0) IS NULL OR material_locations.location_id = $3)
    AND (NULLIF($4::bigint, 0) IS NULL OR materials.id = $4);
