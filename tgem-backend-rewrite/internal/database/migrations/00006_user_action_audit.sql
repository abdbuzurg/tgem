-- +goose Up
-- Activate the user-action audit log: extend user_actions with the metadata
-- the new RecordUserAction middleware emits, and add the indexes the admin
-- read endpoint needs to filter cheaply.
--
-- The admin.user_action resource type and its grants are seeded by migration
-- 00005, so no auth-table changes are needed here.
--
-- Idempotent: ADD COLUMN IF NOT EXISTS / CREATE INDEX IF NOT EXISTS.

-- +goose StatementBegin

ALTER TABLE user_actions ADD COLUMN IF NOT EXISTS http_method text;
ALTER TABLE user_actions ADD COLUMN IF NOT EXISTS request_ip  text;

CREATE INDEX IF NOT EXISTS user_actions_date_idx
    ON user_actions (date_of_action DESC);

CREATE INDEX IF NOT EXISTS user_actions_user_date_idx
    ON user_actions (user_id, date_of_action DESC);

CREATE INDEX IF NOT EXISTS user_actions_project_date_idx
    ON user_actions (project_id, date_of_action DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS user_actions_project_date_idx;
DROP INDEX IF EXISTS user_actions_user_date_idx;
DROP INDEX IF EXISTS user_actions_date_idx;

ALTER TABLE user_actions DROP COLUMN IF EXISTS request_ip;
ALTER TABLE user_actions DROP COLUMN IF EXISTS http_method;

-- +goose StatementEnd
