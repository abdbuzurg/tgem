-- name: CreateInvoiceCount :exec
INSERT INTO invoice_counts (project_id, invoice_type, count) VALUES ($1, $2, $3);

-- name: GetInvoiceCount :one
SELECT COALESCE(count, 0)::bigint AS count
FROM invoice_counts
WHERE invoice_type = $1 AND project_id = $2
LIMIT 1;

-- name: IncrementInvoiceCount :exec
UPDATE invoice_counts
SET count = COALESCE(count, 0) + 1
WHERE invoice_type = $1 AND project_id = $2;

-- name: IncrementInvoiceCountBy :exec
UPDATE invoice_counts
SET count = COALESCE(count, 0) + sqlc.arg(amount)::bigint
WHERE invoice_type = $1 AND project_id = $2;
