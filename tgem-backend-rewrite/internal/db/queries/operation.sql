-- name: GetOperation :one
SELECT id, project_id, name, code, cost_prime, cost_with_customer,
       planned_amount_for_project, show_planned_amount_in_report
FROM operations
WHERE id = $1;

-- name: GetOperationByName :one
SELECT id, project_id, name, code, cost_prime, cost_with_customer,
       planned_amount_for_project, show_planned_amount_in_report
FROM operations
WHERE name = $1 AND project_id = $2
LIMIT 1;

-- name: ListOperationsWithoutMaterials :many
SELECT id, project_id, name, code, cost_prime, cost_with_customer,
       planned_amount_for_project, show_planned_amount_in_report
FROM operations
WHERE
    operations.project_id = $1
    AND operations.id NOT IN (SELECT operation_materials.operation_id FROM operation_materials);

-- name: ListOperationsByProject :many
SELECT
    operations.id                                                              AS id,
    COALESCE(operations.name, '')::text                                        AS name,
    COALESCE(operations.code, '')::text                                        AS code,
    operations.cost_prime                                                      AS cost_prime,
    operations.cost_with_customer                                              AS cost_with_customer,
    COALESCE(materials.id, 0)::bigint                                          AS material_id,
    COALESCE(materials.name, '')::text                                         AS material_name
FROM operations
FULL JOIN operation_materials ON operation_materials.operation_id = operations.id
FULL JOIN materials ON operation_materials.material_id = materials.id
WHERE operations.project_id = $1;

-- name: ListOperationsPaginated :many
SELECT
    operations.id                                                              AS id,
    COALESCE(operations.name, '')::text                                        AS name,
    COALESCE(operations.code, '')::text                                        AS code,
    operations.cost_prime                                                      AS cost_prime,
    operations.cost_with_customer                                              AS cost_with_customer,
    COALESCE(operations.planned_amount_for_project, 0)::float8                 AS planned_amount_for_project,
    COALESCE(operations.show_planned_amount_in_report, false)::boolean         AS show_planned_amount_in_report,
    COALESCE(materials.id, 0)::bigint                                          AS material_id,
    COALESCE(materials.name, '')::text                                         AS material_name
FROM operations
FULL JOIN operation_materials ON operation_materials.operation_id = operations.id
FULL JOIN materials ON operation_materials.material_id = materials.id
WHERE
    operations.project_id = $1
    AND (NULLIF($2::text, '') IS NULL OR operations.name = $2)
    AND (NULLIF($3::text, '') IS NULL OR operations.code = $3)
    AND (NULLIF($4::bigint, 0) IS NULL OR materials.id = $4)
ORDER BY operations.id DESC
LIMIT $5 OFFSET $6;

-- name: CountOperationsFiltered :one
SELECT COUNT(*)::bigint
FROM operations
FULL JOIN operation_materials ON operation_materials.operation_id = operations.id
FULL JOIN materials ON operation_materials.material_id = materials.id
WHERE
    operations.project_id = $1
    AND (NULLIF($2::text, '') IS NULL OR operations.name = $2)
    AND (NULLIF($3::text, '') IS NULL OR operations.code = $3)
    AND (NULLIF($4::bigint, 0) IS NULL OR materials.id = $4);

-- name: CreateOperation :one
INSERT INTO operations (project_id, name, code, cost_prime, cost_with_customer,
                        planned_amount_for_project, show_planned_amount_in_report)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, project_id, name, code, cost_prime, cost_with_customer,
          planned_amount_for_project, show_planned_amount_in_report;

-- name: UpdateOperation :exec
UPDATE operations
SET project_id = $2, name = $3, code = $4, cost_prime = $5, cost_with_customer = $6,
    planned_amount_for_project = $7, show_planned_amount_in_report = $8
WHERE id = $1;

-- name: DeleteOperation :exec
DELETE FROM operations WHERE id = $1;
