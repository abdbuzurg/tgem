-- name: ListMaterialCosts :many
SELECT id, material_id, cost_prime, cost_m19, cost_with_customer
FROM material_costs
ORDER BY id DESC;

-- name: GetMaterialCost :one
SELECT id, material_id, cost_prime, cost_m19, cost_with_customer
FROM material_costs
WHERE id = $1;

-- name: ListMaterialCostsByMaterialID :many
SELECT id, material_id, cost_prime, cost_m19, cost_with_customer
FROM material_costs
WHERE material_id = $1;

-- name: ListMaterialCostsByMaterialIDSorted :many
SELECT id, material_id, cost_prime, cost_m19, cost_with_customer
FROM material_costs
WHERE material_id = $1
ORDER BY cost_m19 DESC;

-- name: GetMaterialCostByCostM19AndMaterialID :one
SELECT id, material_id, cost_prime, cost_m19, cost_with_customer
FROM material_costs
WHERE material_id = $1 AND cost_m19 = $2
LIMIT 1;

-- name: ListMaterialCostsViewByProject :many
SELECT
    material_costs.id                       AS id,
    material_costs.cost_prime               AS cost_prime,
    material_costs.cost_m19                 AS cost_m19,
    material_costs.cost_with_customer       AS cost_with_customer,
    COALESCE(materials.name, '')::text      AS material_name
FROM material_costs
INNER JOIN materials ON materials.id = material_costs.material_id
WHERE materials.project_id = $1
ORDER BY material_costs.id DESC
LIMIT $2 OFFSET $3;

-- name: ListMaterialCostsViewFiltered :many
SELECT
    material_costs.id                       AS id,
    material_costs.cost_prime               AS cost_prime,
    material_costs.cost_m19                 AS cost_m19,
    material_costs.cost_with_customer       AS cost_with_customer,
    COALESCE(materials.name, '')::text      AS material_name
FROM material_costs
INNER JOIN materials ON materials.id = material_costs.material_id
WHERE
    materials.project_id = $1
    AND (NULLIF($2::text, '') IS NULL OR materials.name = $2)
ORDER BY material_costs.id DESC
LIMIT $3 OFFSET $4;

-- name: CountMaterialCostsFiltered :one
SELECT COUNT(*)::bigint
FROM material_costs
INNER JOIN materials ON materials.id = material_costs.material_id
WHERE
    materials.project_id = $1
    AND (NULLIF($2::text, '') IS NULL OR materials.name = $2);

-- name: CreateMaterialCost :one
INSERT INTO material_costs (material_id, cost_prime, cost_m19, cost_with_customer)
VALUES ($1, $2, $3, $4)
RETURNING id, material_id, cost_prime, cost_m19, cost_with_customer;

-- name: UpdateMaterialCost :one
UPDATE material_costs
SET material_id = $2, cost_prime = $3, cost_m19 = $4, cost_with_customer = $5
WHERE id = $1
RETURNING id, material_id, cost_prime, cost_m19, cost_with_customer;

-- name: DeleteMaterialCost :exec
DELETE FROM material_costs WHERE id = $1;

-- name: CreateMaterialCostsBatch :copyfrom
INSERT INTO material_costs (material_id, cost_prime, cost_m19, cost_with_customer)
VALUES ($1, $2, $3, $4);
