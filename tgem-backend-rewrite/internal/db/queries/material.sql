-- name: ListMaterialsByProject :many
SELECT id, category, code, name, unit, notes, has_serial_number, article,
       project_id, planned_amount_for_project, show_planned_amount_in_report
FROM materials
WHERE project_id = $1
ORDER BY id DESC;

-- name: GetMaterialByProjectAndName :one
SELECT id, category, code, name, unit, notes, has_serial_number, article,
       project_id, planned_amount_for_project, show_planned_amount_in_report
FROM materials
WHERE name = $2 AND project_id = $1
LIMIT 1;

-- name: GetMaterialByMaterialCostID :one
SELECT id, category, code, name, unit, notes, has_serial_number, article,
       project_id, planned_amount_for_project, show_planned_amount_in_report
FROM materials
WHERE materials.id IN (
    SELECT material_costs.material_id FROM material_costs WHERE material_costs.id = $1
)
LIMIT 1;

-- name: ListMaterialsByProjectPaginated :many
SELECT id, category, code, name, unit, notes, has_serial_number, article,
       project_id, planned_amount_for_project, show_planned_amount_in_report
FROM materials
WHERE project_id = $1
ORDER BY id DESC
LIMIT $2 OFFSET $3;

-- name: ListMaterialsPaginatedFiltered :many
SELECT id, category, code, name, unit, notes, has_serial_number, article,
       project_id, planned_amount_for_project, show_planned_amount_in_report
FROM materials
WHERE
    project_id = $1
    AND (NULLIF($2::text, '') IS NULL OR category = $2)
    AND (NULLIF($3::text, '') IS NULL OR code = $3)
    AND (NULLIF($4::text, '') IS NULL OR name = $4)
    AND (NULLIF($5::text, '') IS NULL OR unit = $5)
ORDER BY id DESC
LIMIT $6 OFFSET $7;

-- name: GetMaterial :one
SELECT id, category, code, name, unit, notes, has_serial_number, article,
       project_id, planned_amount_for_project, show_planned_amount_in_report
FROM materials
WHERE id = $1;

-- name: CountMaterialsFiltered :one
SELECT COUNT(*)::bigint
FROM materials
WHERE
    project_id = $1
    AND (NULLIF($2::text, '') IS NULL OR category = $2)
    AND (NULLIF($3::text, '') IS NULL OR code = $3)
    AND (NULLIF($4::text, '') IS NULL OR name = $4)
    AND (NULLIF($5::text, '') IS NULL OR unit = $5);

-- name: CreateMaterial :one
INSERT INTO materials (category, code, name, unit, notes, has_serial_number, article,
                       project_id, planned_amount_for_project, show_planned_amount_in_report)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, category, code, name, unit, notes, has_serial_number, article,
          project_id, planned_amount_for_project, show_planned_amount_in_report;

-- name: UpdateMaterial :one
UPDATE materials
SET category = $2, code = $3, name = $4, unit = $5, notes = $6,
    has_serial_number = $7, article = $8, project_id = $9,
    planned_amount_for_project = $10, show_planned_amount_in_report = $11
WHERE id = $1
RETURNING id, category, code, name, unit, notes, has_serial_number, article,
          project_id, planned_amount_for_project, show_planned_amount_in_report;

-- name: DeleteMaterial :exec
DELETE FROM materials WHERE id = $1;

-- name: CreateMaterialsBatch :copyfrom
INSERT INTO materials (category, code, name, unit, notes, has_serial_number, article,
                       project_id, planned_amount_for_project, show_planned_amount_in_report)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
