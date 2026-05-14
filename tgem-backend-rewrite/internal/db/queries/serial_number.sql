-- name: ListSerialNumbers :many
SELECT id, project_id, material_cost_id, code
FROM serial_numbers;

-- name: ListSerialNumbersByMaterialCostID :many
SELECT id, project_id, material_cost_id, code
FROM serial_numbers
WHERE material_cost_id = $1;

-- name: CreateSerialNumber :one
INSERT INTO serial_numbers (project_id, material_cost_id, code)
VALUES ($1, $2, $3)
RETURNING id, project_id, material_cost_id, code;

-- name: UpdateSerialNumber :one
UPDATE serial_numbers
SET project_id = $2, material_cost_id = $3, code = $4
WHERE id = $1
RETURNING id, project_id, material_cost_id, code;

-- name: DeleteSerialNumber :exec
DELETE FROM serial_numbers WHERE id = $1;

-- name: GetSerialNumberCodesByMaterialIDAndLocation :many
SELECT COALESCE(serial_numbers.code, '')::text AS code
FROM material_locations
INNER JOIN serial_numbers ON serial_numbers.material_cost_id = material_locations.material_cost_id
INNER JOIN serial_number_locations ON serial_number_locations.serial_number_id = serial_numbers.id
INNER JOIN material_costs ON material_locations.material_cost_id = material_costs.id
INNER JOIN materials ON materials.id = material_costs.material_id
WHERE
    materials.project_id = $1
    AND materials.id = $2
    AND material_locations.location_type = serial_number_locations.location_type
    AND material_locations.location_type = $3
    AND material_locations.location_id = serial_number_locations.location_id
    AND material_locations.location_id = $4;

