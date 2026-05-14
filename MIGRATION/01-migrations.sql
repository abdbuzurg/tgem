-- =============================================================================
-- Phase 1 — Forward migration
-- =============================================================================
-- This file is consumed in three pieces by the runbook (MIGRATION/03-runbook.md):
--   §A: pre-flight verification (read-only; safe to re-run)
--   §B: Goose baseline stamp (one-time; the cutover-day mutation)
--   §C: which migrations the rewrite container applies automatically (reference)
--   §D: optional manual goose-CLI invocation (fallback if you don't want
--       the container to migrate on first boot)
--   §E: post-migration verification (read-only)
--
-- Run §A and §B against the LIVE production database via psql before starting
-- the rewrite container. The container's MigrateUp then handles §C. Run §E
-- against the live database after the container has started successfully.
--
-- Idempotence: §A and §E are SELECTs. §B uses CREATE TABLE IF NOT EXISTS and
-- ON CONFLICT DO NOTHING — re-running it after a successful stamp is a no-op.
-- =============================================================================


-- =============================================================================
-- §A. Pre-flight verification (run BEFORE any mutation)
-- =============================================================================
-- All five queries must return the documented expected count. Anything else
-- is a STOP condition — read MIGRATION/01-schema-diff.md §5 BLOCKERS.

-- A1. The 50 baseline tables exist. Expected: 50.
SELECT 'A1_table_count' AS check_name, count(*) AS value
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_name IN (
    'auction_items','auction_packages','auction_participant_prices','auctions',
    'districts','invoice_counts','invoice_inputs','invoice_materials',
    'invoice_object_operators','invoice_objects','invoice_operations',
    'invoice_output_out_of_projects','invoice_outputs','invoice_returns',
    'invoice_write_offs','kl04_kv_objects','material_costs','material_defects',
    'material_locations','materials','mjd_objects','object_supervisors',
    'object_teams','objects','operation_materials','operations',
    'operator_error_founds','permissions','project_progress_materials',
    'project_progress_operations','projects','resources','roles','s_ip_objects',
    'serial_number_locations','serial_number_movements','serial_numbers',
    'stvt_objects','substation_cell_nourashes_substation_objects',
    'substation_cell_objects','substation_objects','team_leaders','teams',
    'tp_nourashes_objects','tp_objects','user_actions','user_in_projects',
    'users','worker_attendances','workers'
  );

-- A2. Goose tracking table is not yet present. Expected: 0.
SELECT 'A2_goose_table_absent' AS check_name, count(*) AS value
FROM information_schema.tables
WHERE table_schema = 'public' AND table_name = 'goose_db_version';

-- A3. Permissions-v2 tables not yet present. Expected: 0.
SELECT 'A3_v2_tables_absent' AS check_name, count(*) AS value
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_name IN ('resource_types','permission_actions','role_grants','user_roles');

-- A4. roles.code column not yet present. Expected: 0.
SELECT 'A4_roles_code_absent' AS check_name, count(*) AS value
FROM information_schema.columns
WHERE table_schema = 'public' AND table_name = 'roles' AND column_name = 'code';

-- A5. user_actions.http_method not yet present. Expected: 0.
SELECT 'A5_user_actions_http_method_absent' AS check_name, count(*) AS value
FROM information_schema.columns
WHERE table_schema = 'public' AND table_name = 'user_actions' AND column_name = 'http_method';

-- A6. Row-count snapshot for blast-radius awareness (informational, not a check).
-- Capture these for the cutover incident log so post-migration deltas are
-- attributable.
SELECT 'A6_users_count'           AS check_name, count(*) AS value FROM users;
SELECT 'A6_roles_count'           AS check_name, count(*) AS value FROM roles;
SELECT 'A6_projects_count'        AS check_name, count(*) AS value FROM projects;
SELECT 'A6_invoice_counts_count'  AS check_name, count(*) AS value FROM invoice_counts;
SELECT 'A6_objects_count'         AS check_name, count(*) AS value FROM objects;
SELECT 'A6_user_actions_count'    AS check_name, count(*) AS value FROM user_actions;


-- =============================================================================
-- §B. Goose baseline stamp (ONE-TIME MUTATION)
-- =============================================================================
-- This is the only schema change Phase 1 makes by hand. Everything else is
-- handled by the rewrite container's embedded MigrateUp once it sees the
-- stamp.
--
-- Why this is necessary: the rewrite uses Goose with 00001_initial_schema.sql
-- as a *full schema baseline*. If we let MigrateUp run against the existing
-- production schema, it would try to CREATE TABLE every legacy table and
-- error out at the first "already exists." The fix is to tell Goose "version
-- 1 is already applied" before the binary touches the database.
--
-- The shape of goose_db_version below matches Goose v3's table schema
-- (github.com/pressly/goose/v3). The version_id=0 row is Goose's convention
-- for "initial empty state." The version_id=1 row says "00001 done."
-- 00002..00006 will be applied by MigrateUp because their version_id rows
-- are absent.

BEGIN;

CREATE TABLE IF NOT EXISTS public.goose_db_version (
    id          serial      PRIMARY KEY,
    version_id  bigint      NOT NULL,
    is_applied  boolean     NOT NULL,
    tstamp      timestamp   NULL DEFAULT now()
);

-- Use INSERT...SELECT WHERE NOT EXISTS instead of ON CONFLICT because there
-- is no UNIQUE constraint on (version_id, is_applied) in Goose's schema.
INSERT INTO public.goose_db_version (version_id, is_applied)
SELECT 0, true
WHERE NOT EXISTS (
    SELECT 1 FROM public.goose_db_version
    WHERE version_id = 0 AND is_applied = true
);

INSERT INTO public.goose_db_version (version_id, is_applied)
SELECT 1, true
WHERE NOT EXISTS (
    SELECT 1 FROM public.goose_db_version
    WHERE version_id = 1 AND is_applied = true
);

COMMIT;

-- Post-stamp sanity check: the stamp should show version 1 as the latest
-- applied. Expected: 1.
SELECT 'B_max_applied_version' AS check_name, COALESCE(MAX(version_id), -1) AS value
FROM public.goose_db_version
WHERE is_applied = true;


-- =============================================================================
-- §C. Migrations the rewrite container applies automatically (REFERENCE ONLY)
-- =============================================================================
-- The container's MigrateUp() (tgem-backend-rewrite/internal/database/migrate.go)
-- runs at every startup. After §B has stamped 00001 as applied, MigrateUp
-- will see versions 2..6 are pending and apply them in order:
--
--   00002_phase7_data_cleanup.sql
--   00003_align_correction_resource_url.sql
--   00004_split_output_out_of_project_counter.sql
--   00005_permissions_v2_foundation.sql
--   00006_user_action_audit.sql
--
-- The container's logs will show one Goose line per migration. On success
-- the row count in goose_db_version reaches 7 (version 0,1,2,3,4,5,6 — all
-- is_applied=true). On failure the transaction rolls back, the binary
-- log.Fatals, and Docker restart-loops the container. See §E.
--
-- DO NOT copy the migration bodies here — they live in the rewrite repo and
-- must stay the single source of truth.


-- =============================================================================
-- §D. Optional: manual goose-CLI fallback
-- =============================================================================
-- Use this path instead of §C if you want migrations applied BEFORE the
-- rewrite container starts, so you can verify them in isolation. Requires
-- the goose CLI v3 (https://github.com/pressly/goose).
--
-- Prerequisite: §A and §B have already run; the goose_db_version table
-- exists and shows version 1 as applied.
--
-- The migrations directory is read from the rewrite source tree (the
-- container has them embedded in its binary, but a host-side goose run
-- reads them off disk).
--
-- Replace the connection parameters with the production values.
--
--   export DATABASE_URL="postgres://USER:PASSWORD@127.0.0.1:5432/DBNAME?sslmode=disable"
--   cd /path/to/repo
--   goose -dir ./tgem-backend-rewrite/internal/database/migrations \
--         postgres "$DATABASE_URL" status
--   # ^ expected output: 00001 applied; 00002..00006 pending
--   goose -dir ./tgem-backend-rewrite/internal/database/migrations \
--         postgres "$DATABASE_URL" up
--   goose -dir ./tgem-backend-rewrite/internal/database/migrations \
--         postgres "$DATABASE_URL" status
--   # ^ expected output: all six applied


-- =============================================================================
-- §E. Post-migration verification (run AFTER §C or §D succeeds)
-- =============================================================================

-- E1. Goose reports the highest applied version. Expected: 6.
SELECT 'E1_max_applied_version' AS check_name, MAX(version_id) AS value
FROM goose_db_version WHERE is_applied = true;

-- E2. The four permissions-v2 tables now exist. Expected: 4.
SELECT 'E2_v2_tables_present' AS check_name, count(*) AS value
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_name IN ('resource_types','permission_actions','role_grants','user_roles');

-- E3. Seeded permission actions. Expected: 9.
SELECT 'E3_permission_actions_seeded' AS check_name, count(*) AS value
FROM permission_actions;

-- E4. Seeded resource types. Expected: 40.
SELECT 'E4_resource_types_seeded' AS check_name, count(*) AS value
FROM resource_types;

-- E5. user_roles backfilled. Expected: same as count of users with role_id.
SELECT 'E5_user_roles_backfill_match' AS check_name,
       (SELECT count(*) FROM user_roles WHERE project_id IS NULL)
       = (SELECT count(*) FROM users    WHERE role_id IS NOT NULL) AS value;

-- E6. roles.code is NOT NULL and populated.
SELECT 'E6_roles_code_no_nulls' AS check_name, count(*) AS value
FROM roles WHERE code IS NULL;
-- expected: 0

-- E7. output-out-of-project counter seeded for every project.
SELECT 'E7_oop_counter_per_project' AS check_name,
       (SELECT count(*) FROM invoice_counts WHERE invoice_type = 'output-out-of-project')
       = (SELECT count(*) FROM projects) AS value;

-- E8. Correction resource URL aligned (idempotent, may already be the new value).
SELECT 'E8_correction_url_present' AS check_name, count(*) AS value
FROM resources WHERE name = 'Корректировка оператора' AND url = '/correction';
-- expected: 1 (legacy seed) or 0 if production has a customized resource name

-- E9. user_actions audit columns present.
SELECT 'E9_user_actions_audit_cols' AS check_name, count(*) AS value
FROM information_schema.columns
WHERE table_schema = 'public' AND table_name = 'user_actions'
  AND column_name IN ('http_method','request_ip');
-- expected: 2

-- E10. user_actions indexes present.
SELECT 'E10_user_actions_indexes' AS check_name, count(*) AS value
FROM pg_indexes
WHERE schemaname = 'public' AND tablename = 'user_actions'
  AND indexname IN ('user_actions_date_idx','user_actions_user_date_idx','user_actions_project_date_idx');
-- expected: 3

-- E11. auction_participant_prices unique index present.
SELECT 'E11_app_unique_index' AS check_name, count(*) AS value
FROM pg_indexes
WHERE schemaname = 'public' AND tablename = 'auction_participant_prices'
  AND indexname = 'auction_participant_prices_item_user_uq';
-- expected: 1
