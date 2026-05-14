-- =============================================================================
-- Phase 1 — Best-effort rollback
-- =============================================================================
-- This file undoes the Phase 1 forward migration in reverse order
-- (00006 → 00005 → 00004 → 00003 → 00002 → un-stamp Goose). Each section is
-- annotated LOSSLESS or LOSSY.
--
-- LOSSLESS = the schema and data can be returned to the pre-migration state
--            with no information loss.
-- LOSSY    = some information cannot be recovered by SQL alone. The only
--            reliable rollback for these is pg_restore from the pre-migration
--            backup (taken in runbook step 3.2).
--
-- DO NOT run this file as-is in one shot. It is meant to be reviewed and
-- executed section-by-section by an operator who is rolling back. The
-- runbook (MIGRATION/03-rollback.md) selects which sections apply based on
-- how far the forward migration progressed.
--
-- If the forward migration completed (Goose at version 6) and the rollback
-- target is "back to legacy production schema," the recommended path is
-- pg_restore from the pre-migration backup. SQL rollback is provided here
-- for cases where data written between migration and rollback must be
-- preserved (i.e. the rewrite was serving traffic for some minutes/hours
-- and you want to keep its writes but undo the schema changes).
-- =============================================================================


-- =============================================================================
-- §1. Undo 00006_user_action_audit.sql  —  LOSSLESS
-- =============================================================================
-- Drops the three indexes and the two columns. Any data in http_method or
-- request_ip is discarded; those columns are an audit annotation, not
-- business state, so this is functionally lossless from the application's
-- perspective.

BEGIN;
DROP INDEX IF EXISTS user_actions_project_date_idx;
DROP INDEX IF EXISTS user_actions_user_date_idx;
DROP INDEX IF EXISTS user_actions_date_idx;
ALTER TABLE user_actions DROP COLUMN IF EXISTS request_ip;
ALTER TABLE user_actions DROP COLUMN IF EXISTS http_method;
COMMIT;


-- =============================================================================
-- §2. Undo 00005_permissions_v2_foundation.sql  —  LOSSLESS
-- =============================================================================
-- The v1 permissions/resources/roles tables remained authoritative throughout
-- the migration window (the rewrite's permissions phase 4 is not part of
-- THIS migration). The v2 tables (resource_types, permission_actions,
-- role_grants, user_roles) and roles.code are pure additions, never read
-- by the legacy code and never written-to by the user during the cutover
-- window unless the rewrite's new admin UI was used.
--
-- LOSSY caveat: if the rewrite's phase-3 admin UI WAS used after cutover
-- and operators customized role_grants or user_roles, those customizations
-- are dropped here. The pre-migration backup is the authoritative recovery.

BEGIN;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS role_grants;
ALTER TABLE roles DROP CONSTRAINT IF EXISTS roles_code_unique;
ALTER TABLE roles DROP COLUMN IF EXISTS code;
DROP TABLE IF EXISTS resource_types;
DROP TABLE IF EXISTS permission_actions;
COMMIT;


-- =============================================================================
-- §3. Undo 00004_split_output_out_of_project_counter.sql  —  LOSSY
-- =============================================================================
-- Removes all invoice_counts rows where invoice_type = 'output-out-of-project'.
-- LOSSY because if any out-of-project invoices were issued post-migration,
-- they incremented the counter and the next Create would now produce a
-- duplicate delivery code. Mitigations:
--   - Stop the rewrite container before running this.
--   - Verify no rows have been created in invoice_output_out_of_projects
--     with delivery codes that depend on this counter
--     (timestamp > the cutover timestamp).
--   - If verifying is too costly, pg_restore is the safer path.

BEGIN;
DELETE FROM invoice_counts WHERE invoice_type = 'output-out-of-project';
COMMIT;


-- =============================================================================
-- §4. Undo 00003_align_correction_resource_url.sql  —  LOSSLESS
-- =============================================================================
-- One row. Conditional update. Safe to run; idempotent if not still pointing
-- at the new URL.

BEGIN;
UPDATE resources
SET url = '/invoice-correction'
WHERE name = 'Корректировка оператора'
  AND url = '/correction';
COMMIT;


-- =============================================================================
-- §5. Undo 00002_phase7_data_cleanup.sql  —  LOSSY (mostly)
-- =============================================================================
-- The Up migration deleted orphan rows, coalesced duplicates, retagged
-- misclassified rows, and backfilled invoice_counts. None of those data
-- changes can be undone by SQL alone.
--
-- LOSSLESS portion: drop the auction_participant_prices unique index.
-- Required if you intend to restart the legacy backend, because the legacy
-- code path that creates duplicate rows under concurrency would now hit a
-- constraint violation.

BEGIN;
DROP INDEX IF EXISTS auction_participant_prices_item_user_uq;
COMMIT;

-- The data-cleanup steps below are NO-OPS — there is no SQL that resurrects
-- a deleted orphan or re-duplicates a coalesced row. To get pre-migration
-- data back, pg_restore from the pre-migration backup is the only path.
-- The runbook MIGRATION/03-rollback.md prescribes pg_restore as the
-- canonical rollback for this migration's data half.


-- =============================================================================
-- §6. Un-stamp Goose  —  LOSSLESS
-- =============================================================================
-- After running §1..§5, the database is structurally back to pre-migration
-- (modulo §5's data losses). Remove the Goose tracking table so a future
-- forward migration can stamp it again cleanly.

BEGIN;
DROP TABLE IF EXISTS public.goose_db_version;
COMMIT;


-- =============================================================================
-- §7. Verification after rollback
-- =============================================================================
-- Run these to confirm the database is back to pre-migration shape.

SELECT 'R_goose_table_absent' AS check_name, count(*) AS value
FROM information_schema.tables
WHERE table_schema = 'public' AND table_name = 'goose_db_version';
-- expected: 0

SELECT 'R_v2_tables_absent' AS check_name, count(*) AS value
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_name IN ('resource_types','permission_actions','role_grants','user_roles');
-- expected: 0

SELECT 'R_roles_code_absent' AS check_name, count(*) AS value
FROM information_schema.columns
WHERE table_schema = 'public' AND table_name = 'roles' AND column_name = 'code';
-- expected: 0

SELECT 'R_user_actions_audit_absent' AS check_name, count(*) AS value
FROM information_schema.columns
WHERE table_schema = 'public' AND table_name = 'user_actions'
  AND column_name IN ('http_method','request_ip');
-- expected: 0

SELECT 'R_app_unique_index_absent' AS check_name, count(*) AS value
FROM pg_indexes
WHERE schemaname = 'public' AND tablename = 'auction_participant_prices'
  AND indexname = 'auction_participant_prices_item_user_uq';
-- expected: 0

SELECT 'R_oop_counter_rows' AS check_name, count(*) AS value
FROM invoice_counts WHERE invoice_type = 'output-out-of-project';
-- expected: 0 (rolled back), unless production used that counter pre-migration


-- =============================================================================
-- §8. Notes on pg_restore as the canonical rollback
-- =============================================================================
-- If a full rollback is needed and you cannot tolerate the §3 and §5 data
-- losses, restore from the pre-migration backup taken in runbook step 3.2:
--
--   sudo systemctl stop nginx                     # halt traffic
--   docker compose stop backend                   # halt the rewrite
--   psql -h localhost -U <superuser> -c \
--       "REVOKE CONNECT ON DATABASE <db> FROM <app_user>"   # kick app sessions
--   psql -h localhost -U <superuser> -c \
--       "SELECT pg_terminate_backend(pid) FROM pg_stat_activity \
--        WHERE datname = '<db>' AND pid <> pg_backend_pid()"
--   dropdb -h localhost -U <superuser> <db>
--   createdb -h localhost -U <superuser> -O <app_user> <db>
--   pg_restore -h localhost -U <superuser> -d <db> ~/backups/tgem-<date>/db.dump
--   psql -h localhost -U <superuser> -c \
--       "GRANT CONNECT ON DATABASE <db> TO <app_user>"
--   pm2 start <legacy-backend-name>                # bring legacy back up
--   sudo systemctl start nginx
--
-- Any writes that hit the rewrite after cutover are lost by pg_restore.
-- That is the explicit trade-off of choosing pg_restore over piecemeal SQL.
