-- name: ListSerialNumberMovementsByInvoice :many
SELECT id, serial_number_id, project_id, invoice_id, invoice_type, is_defected, confirmation
FROM serial_number_movements
WHERE invoice_id = $1 AND invoice_type = $2;

-- name: CreateSerialNumberMovementsBatch :copyfrom
INSERT INTO serial_number_movements (serial_number_id, project_id, invoice_id, invoice_type, is_defected, confirmation)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: DeleteSerialNumberMovementsByInvoice :exec
DELETE FROM serial_number_movements
WHERE invoice_type = $1 AND invoice_id = $2;

-- name: ConfirmSerialNumberMovementsByInvoice :exec
UPDATE serial_number_movements
SET confirmation = true
WHERE invoice_id = $1 AND invoice_type = $2;
