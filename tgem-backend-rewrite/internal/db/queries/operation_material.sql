-- name: CreateOperationMaterial :exec
INSERT INTO operation_materials (operation_id, material_id) VALUES ($1, $2);

-- name: DeleteOperationMaterialsByOperationID :exec
DELETE FROM operation_materials WHERE operation_id = $1;

-- name: GetOperationMaterialByOperationID :one
SELECT id, operation_id, material_id
FROM operation_materials
WHERE operation_id = $1
LIMIT 1;
