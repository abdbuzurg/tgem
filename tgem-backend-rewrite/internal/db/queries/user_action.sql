-- name: ListUserActionsByUserID :many
SELECT id, action_url, action_type, action_id, action_status,
       action_status_message, http_method, request_ip,
       user_id, project_id, date_of_action
FROM user_actions
WHERE user_id = $1
ORDER BY id DESC;

-- name: CreateUserAction :one
INSERT INTO user_actions (action_url, action_type, action_id, action_status,
                          action_status_message, http_method, request_ip,
                          user_id, project_id, date_of_action)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, action_url, action_type, action_id, action_status,
          action_status_message, http_method, request_ip,
          user_id, project_id, date_of_action;

-- name: ListUserActionsPaginated :many
-- Optional filters: pass 0 for bigint, '' for text, or sentinel time bounds to disable.
-- DateFrom defaults to '0001-01-01' (no lower bound), DateTo defaults to '9999-12-31'
-- (no upper bound) — caller substitutes these when the filter is unset.
SELECT
    user_actions.id                                     AS id,
    COALESCE(user_actions.action_url, '')::text         AS action_url,
    COALESCE(user_actions.action_type, '')::text        AS action_type,
    COALESCE(user_actions.action_id, 0)::bigint         AS action_id,
    COALESCE(user_actions.action_status, false)::boolean AS action_status,
    COALESCE(user_actions.action_status_message, '')::text AS action_status_message,
    COALESCE(user_actions.http_method, '')::text        AS http_method,
    COALESCE(user_actions.request_ip, '')::text         AS request_ip,
    COALESCE(user_actions.user_id, 0)::bigint           AS user_id,
    COALESCE(user_actions.project_id, 0)::bigint        AS project_id,
    COALESCE(users.username, '')::text                  AS username,
    user_actions.date_of_action                         AS date_of_action
FROM user_actions
LEFT JOIN users ON users.id = user_actions.user_id
WHERE
    (NULLIF($1::bigint, 0) IS NULL OR user_actions.user_id = $1)
    AND (NULLIF($2::bigint, 0) IS NULL OR user_actions.project_id = $2)
    AND (NULLIF($3::text, '') IS NULL OR user_actions.action_type = $3)
    AND (NULLIF($4::text, '') IS NULL OR user_actions.http_method = $4)
    AND ($5::boolean IS FALSE OR user_actions.action_status = $6)
    AND user_actions.date_of_action >= $7
    AND user_actions.date_of_action <= $8
ORDER BY user_actions.date_of_action DESC, user_actions.id DESC
LIMIT $9 OFFSET $10;

-- name: CountUserActionsPaginated :one
SELECT COUNT(*)::bigint
FROM user_actions
WHERE
    (NULLIF($1::bigint, 0) IS NULL OR user_actions.user_id = $1)
    AND (NULLIF($2::bigint, 0) IS NULL OR user_actions.project_id = $2)
    AND (NULLIF($3::text, '') IS NULL OR user_actions.action_type = $3)
    AND (NULLIF($4::text, '') IS NULL OR user_actions.http_method = $4)
    AND ($5::boolean IS FALSE OR user_actions.action_status = $6)
    AND user_actions.date_of_action >= $7
    AND user_actions.date_of_action <= $8;

-- name: ListUserOptionsForAudit :many
-- Powers the audit-log page's user filter — admins type a name or a login
-- and pick from the dropdown. Returns a flat (id, username, workerName)
-- list, sorted by worker name then username for stable display.
SELECT
    users.id                                  AS id,
    COALESCE(users.username, '')::text        AS username,
    COALESCE(workers.name, '')::text          AS worker_name
FROM users
LEFT JOIN workers ON workers.id = users.worker_id
ORDER BY worker_name, username;
