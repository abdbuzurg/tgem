-- +goose Up

-- Phase 7 data cleanup: brings the database into a known clean state
-- BEFORE the phase-7 code fixes start running. Each section is a no-op on
-- a database that's never run the buggy code; on production it sweeps up
-- orphan rows, coalesces duplicates introduced by the bugs, and backfills
-- counters that previous code never created.
--
-- All operations are non-destructive of valid data. Orphan deletions
-- target rows that the application has been ignoring; the duplicate
-- coalesce preserves total amounts via SUM; the auction-duplicate cleanup
-- keeps the latest entry; the invoice_counts backfill matches existing
-- counts via COUNT(*).

-- +goose StatementBegin

-- =============================================================================
-- 0. Pre-cleanup: stage every `objects` row this migration will delete, then
--    remove all child references across the 9 FK-referencing tables. Without
--    this, the parent DELETEs in sections 1 and 4 fail with FK violations
--    against production-shape data.
--    Goose runs each migration in a single transaction, so the temp table
--    with ON COMMIT DROP is cleaned up automatically on success/rollback.
-- =============================================================================
CREATE TEMP TABLE _soon_to_be_deleted_objects ON COMMIT DROP AS
SELECT id FROM objects WHERE
    (type='kl04kv_objects'          AND object_detailed_id IS NOT NULL AND object_detailed_id NOT IN (SELECT id FROM kl04_kv_objects))         OR
    (type='mjd_objects'             AND object_detailed_id IS NOT NULL AND object_detailed_id NOT IN (SELECT id FROM mjd_objects))             OR
    (type='sip_objects'             AND object_detailed_id IS NOT NULL AND object_detailed_id NOT IN (SELECT id FROM s_ip_objects))            OR
    (type='stvt_objects'            AND object_detailed_id IS NOT NULL AND object_detailed_id NOT IN (SELECT id FROM stvt_objects))            OR
    (type='tp_objects'              AND object_detailed_id IS NOT NULL AND object_detailed_id NOT IN (SELECT id FROM tp_objects))              OR
    (type='substation_objects'      AND object_detailed_id IS NOT NULL AND object_detailed_id NOT IN (SELECT id FROM substation_objects))      OR
    (type='substation_cell_objects' AND object_detailed_id IS NOT NULL AND object_detailed_id NOT IN (SELECT id FROM substation_cell_objects));

DELETE FROM object_teams                                 WHERE object_id                  IN (SELECT id FROM _soon_to_be_deleted_objects);
DELETE FROM object_supervisors                           WHERE object_id                  IN (SELECT id FROM _soon_to_be_deleted_objects);
DELETE FROM invoice_objects                              WHERE object_id                  IN (SELECT id FROM _soon_to_be_deleted_objects);
DELETE FROM tp_nourashes_objects                         WHERE tp_object_id               IN (SELECT id FROM _soon_to_be_deleted_objects);
DELETE FROM substation_cell_nourashes_substation_objects WHERE substation_object_id       IN (SELECT id FROM _soon_to_be_deleted_objects);
DELETE FROM substation_cell_nourashes_substation_objects WHERE substation_cell_object_id  IN (SELECT id FROM _soon_to_be_deleted_objects);
DELETE FROM supervisor_objects                           WHERE object_id                  IN (SELECT id FROM _soon_to_be_deleted_objects);
DELETE FROM team_objects                                 WHERE object_id                  IN (SELECT id FROM _soon_to_be_deleted_objects);
DELETE FROM object_operations                            WHERE object_id                  IN (SELECT id FROM _soon_to_be_deleted_objects);

-- =============================================================================
-- 1. kl04kv_object Delete bug (6.21):
--    The buggy Delete filtered the parent `objects` row by
--    type='kl04_kv_objects' (with underscore) instead of 'kl04kv_objects',
--    so the parent row was never removed. Sweep up orphans.
-- =============================================================================
DELETE FROM objects
WHERE type = 'kl04kv_objects'
  AND object_detailed_id IS NOT NULL
  AND object_detailed_id NOT IN (SELECT id FROM kl04_kv_objects);

-- =============================================================================
-- 2. mjd_object Update bug (6.22): tp_nourashes_objects rows for mjd targets
--    were tagged target_type='kl04kv_objects' (typo) on Update, leaving
--    legitimate 'mjd_objects' rows alongside misclassified 'kl04kv_objects'
--    rows for the same target_id. Re-tag misclassified rows.
--    Real kl04kv rows are untouched: the filter requires the target_id to
--    refer to an mjd_object's parent objects row.
-- =============================================================================
UPDATE tp_nourashes_objects
SET target_type = 'mjd_objects'
WHERE target_type = 'kl04kv_objects'
  AND target_id IN (
    SELECT id FROM objects WHERE type = 'mjd_objects'
  );

-- =============================================================================
-- 3. sip_object Delete bug (6.23): Delete didn't cascade tp_nourashes_objects.
--    Sweep up rows whose target_id no longer exists as a sip_object's parent.
-- =============================================================================
DELETE FROM tp_nourashes_objects
WHERE target_type = 'sip_objects'
  AND target_id IS NOT NULL
  AND target_id NOT IN (
    SELECT id FROM objects WHERE type = 'sip_objects'
  );

-- =============================================================================
-- 4. substation_object Delete bugs (6.25): the cascade chain referenced
--    tp_objects/'tp_objects' instead of substation_objects in three places,
--    so deleting a substation only removed the substation_objects detail
--    row — leaving its object_supervisors, object_teams, and parent objects
--    row orphaned. Sweep up orphans.
--    Note: this also catches orphans from any other source. Object_supervisors
--    and object_teams should always reference live objects, so this is a
--    conservative cleanup.
-- =============================================================================
DELETE FROM object_supervisors
WHERE object_id IS NOT NULL
  AND object_id NOT IN (SELECT id FROM objects);

DELETE FROM object_teams
WHERE object_id IS NOT NULL
  AND object_id NOT IN (SELECT id FROM objects);

DELETE FROM objects
WHERE type = 'substation_objects'
  AND object_detailed_id IS NOT NULL
  AND object_detailed_id NOT IN (SELECT id FROM substation_objects);

-- Same orphan-cleanup for kl04kv (already handled above) plus the other
-- object-family types — defensive sweep covering future incidents.
DELETE FROM objects
WHERE type = 'mjd_objects'
  AND object_detailed_id IS NOT NULL
  AND object_detailed_id NOT IN (SELECT id FROM mjd_objects);

DELETE FROM objects
WHERE type = 'sip_objects'
  AND object_detailed_id IS NOT NULL
  AND object_detailed_id NOT IN (SELECT id FROM s_ip_objects);

DELETE FROM objects
WHERE type = 'stvt_objects'
  AND object_detailed_id IS NOT NULL
  AND object_detailed_id NOT IN (SELECT id FROM stvt_objects);

DELETE FROM objects
WHERE type = 'tp_objects'
  AND object_detailed_id IS NOT NULL
  AND object_detailed_id NOT IN (SELECT id FROM tp_objects);

DELETE FROM objects
WHERE type = 'substation_cell_objects'
  AND object_detailed_id IS NOT NULL
  AND object_detailed_id NOT IN (SELECT id FROM substation_cell_objects);

-- =============================================================================
-- 5. invoice_output_out_of_project Confirmation bug (6.30): the inner-loop
--    typo set materialInWarehouseIndex (already used elsewhere) instead of
--    materialOutOfProjectIndex, so every invoice_material always appended
--    a NEW out-of-project material_locations row instead of merging into
--    an existing one. Coalesce duplicates: keep one row per
--    (project_id, material_cost_id, location_id, location_type='out-of-project'),
--    summing amounts across the duplicates.
-- =============================================================================
WITH grouped AS (
    SELECT
        MIN(id) AS keeper_id,
        SUM(amount) AS total_amount,
        project_id, material_cost_id, location_id, location_type
    FROM material_locations
    WHERE location_type = 'out-of-project'
    GROUP BY project_id, material_cost_id, location_id, location_type
    HAVING COUNT(*) > 1
)
UPDATE material_locations ml
SET amount = grouped.total_amount
FROM grouped
WHERE ml.id = grouped.keeper_id;

DELETE FROM material_locations ml
WHERE ml.location_type = 'out-of-project'
  AND ml.id <> (
    SELECT MIN(id)
    FROM material_locations
    WHERE location_type = 'out-of-project'
      AND project_id IS NOT DISTINCT FROM ml.project_id
      AND material_cost_id IS NOT DISTINCT FROM ml.material_cost_id
      AND location_id IS NOT DISTINCT FROM ml.location_id
  );

-- =============================================================================
-- 6. invoice_input/output/return/writeoff GetInvoiceCount ErrNoRows fold (6.28):
--    project Create now seeds invoice_counts rows on project creation, but
--    older projects never had them. When invoice_counts row is missing for
--    a (project_id, invoice_type), the increment query is a silent no-op
--    and Create-side delivery-code generation produces duplicates. Backfill
--    rows for all (project, type) pairs missing them, with count = number
--    of existing invoices of that type.
-- =============================================================================
INSERT INTO invoice_counts (project_id, invoice_type, count)
SELECT projects.id, 'input',
       COALESCE((SELECT COUNT(*)::bigint FROM invoice_inputs WHERE project_id = projects.id), 0)
FROM projects
WHERE NOT EXISTS (
    SELECT 1 FROM invoice_counts
    WHERE invoice_counts.project_id = projects.id
      AND invoice_counts.invoice_type = 'input'
);

-- The 'output' counter is shared between invoice_outputs and
-- invoice_output_out_of_projects (both increment 'output' in their Create
-- paths).
INSERT INTO invoice_counts (project_id, invoice_type, count)
SELECT projects.id, 'output',
       COALESCE((SELECT COUNT(*)::bigint FROM invoice_outputs WHERE project_id = projects.id), 0)
       + COALESCE((SELECT COUNT(*)::bigint FROM invoice_output_out_of_projects WHERE project_id = projects.id), 0)
FROM projects
WHERE NOT EXISTS (
    SELECT 1 FROM invoice_counts
    WHERE invoice_counts.project_id = projects.id
      AND invoice_counts.invoice_type = 'output'
);

INSERT INTO invoice_counts (project_id, invoice_type, count)
SELECT projects.id, 'return',
       COALESCE((SELECT COUNT(*)::bigint FROM invoice_returns WHERE project_id = projects.id), 0)
FROM projects
WHERE NOT EXISTS (
    SELECT 1 FROM invoice_counts
    WHERE invoice_counts.project_id = projects.id
      AND invoice_counts.invoice_type = 'return'
);

INSERT INTO invoice_counts (project_id, invoice_type, count)
SELECT projects.id, 'writeoff',
       COALESCE((SELECT COUNT(*)::bigint FROM invoice_write_offs WHERE project_id = projects.id), 0)
FROM projects
WHERE NOT EXISTS (
    SELECT 1 FROM invoice_counts
    WHERE invoice_counts.project_id = projects.id
      AND invoice_counts.invoice_type = 'writeoff'
);

-- =============================================================================
-- 7. auction_participant_prices race condition (6.4):
--    GORM FirstOrCreate-in-a-loop with no DB-level uniqueness allowed
--    duplicate rows for (auction_item_id, user_id) under concurrent writes.
--    Keep the latest row per pair (highest id) and add a unique index so
--    future inserts can use ON CONFLICT.
-- =============================================================================
DELETE FROM auction_participant_prices
WHERE id NOT IN (
    SELECT MAX(id)
    FROM auction_participant_prices
    GROUP BY auction_item_id, user_id
);

CREATE UNIQUE INDEX IF NOT EXISTS auction_participant_prices_item_user_uq
    ON auction_participant_prices (auction_item_id, user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- The Down migration cannot resurrect deleted orphan rows or split coalesced
-- rows back into duplicates — those operations are by design irreversible
-- (the rows were unreachable to begin with). The unique index can be
-- dropped, and the backfilled invoice_counts rows can be removed by
-- counting on the same condition (count = number of existing invoices),
-- but that's only safe if no new invoices were created since. To keep
-- this migration safely reversible-on-empty-changes, Down only drops the
-- index — the data cleanup steps are no-ops on rollback.
DROP INDEX IF EXISTS auction_participant_prices_item_user_uq;
-- +goose StatementEnd
