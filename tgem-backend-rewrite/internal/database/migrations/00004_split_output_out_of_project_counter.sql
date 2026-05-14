-- +goose Up
-- Split the output-out-of-project counter from the regular output counter.
-- Previously both invoice_output_usecase and invoice_output_out_of_project_usecase
-- read/incremented the same invoice_counts row keyed on (project_id, 'output'),
-- so out-of-project invoices burned slots in the regular output namespace —
-- visible in project listings as unexpected delivery codes (the "G-2"/"О-02"
-- bug reported in project Турсунзода).
--
-- The code now uses InvoiceType = 'output-out-of-project' AND a distinct
-- prefix "ОВ" for newly issued codes, so the two flavors no longer collide.
-- Already-issued codes are left untouched (renumbering would break printed/
-- exported references). This migration just seeds the new counter row at 0
-- for each existing project.
--
-- Idempotent: only inserts when the row is missing.

-- +goose StatementBegin
INSERT INTO invoice_counts (project_id, invoice_type, count)
SELECT projects.id, 'output-out-of-project', 0
FROM projects
WHERE NOT EXISTS (
    SELECT 1 FROM invoice_counts
    WHERE invoice_counts.project_id = projects.id
      AND invoice_counts.invoice_type = 'output-out-of-project'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM invoice_counts WHERE invoice_type = 'output-out-of-project';
-- +goose StatementEnd
