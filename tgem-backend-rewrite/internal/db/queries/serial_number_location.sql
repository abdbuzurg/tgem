-- name: ListSerialNumberLocations :many
SELECT id, serial_number_id, project_id, location_id, location_type
FROM serial_number_locations;

-- name: CreateSerialNumberLocationsBatch :copyfrom
INSERT INTO serial_number_locations (serial_number_id, project_id, location_id, location_type)
VALUES ($1, $2, $3, $4);
