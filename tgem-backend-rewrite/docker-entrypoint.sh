#!/bin/sh
# Templates /app/configurations/App_dev.yaml from env vars (sourced from
# .env.backend by docker-compose), then exec's the API binary.
#
# Why: the rewrite's viper config loader reads a hardcoded yaml path with
# no env-var override (internal/config/config.go). Until that's changed in
# the rewrite, the container synthesizes the yaml at start time.
#
# Required env vars:
#   DB_HOST DB_USERNAME DB_PASSWORD DB_NAME JWT_SECRET
#
# Optional env vars (defaults shown):
#   APP_HOST=0.0.0.0    bind address; 0.0.0.0 for container, 127.0.0.1 for host
#   APP_PORT=5000       HTTP port
#   DB_PORT=5432        Postgres port
#   FILES_PATH=./files  relative path for legacy Files.Path config
#
# The script fails fast if a required var is unset, so docker-compose
# surfaces the misconfiguration in its restart loop instead of producing a
# corrupt config.
set -eu

: "${APP_HOST:=0.0.0.0}"
: "${APP_PORT:=5000}"
: "${DB_HOST:?DB_HOST is required (use host.docker.internal for host Postgres)}"
: "${DB_PORT:=5432}"
: "${DB_USERNAME:?DB_USERNAME is required}"
: "${DB_PASSWORD:?DB_PASSWORD is required}"
: "${DB_NAME:?DB_NAME is required}"
: "${JWT_SECRET:?JWT_SECRET is required (must equal the legacy production value to keep sessions valid through cutover)}"
: "${FILES_PATH:=./files}"

# Reject characters that would break double-quoted YAML scalars. If your
# password actually needs these, change the yaml below to a block scalar.
for var_name in DB_PASSWORD JWT_SECRET DB_USERNAME DB_NAME; do
    eval "value=\${${var_name}}"
    case "$value" in
        *'"'*|*'\'*)
            echo "ERROR: ${var_name} contains a double-quote or backslash; escape or simplify it." >&2
            exit 1
            ;;
    esac
done

cat > /app/configurations/App_dev.yaml <<EOF
App:
  Host: "${APP_HOST}"
  Port: ${APP_PORT}

Database:
  Host: "${DB_HOST}"
  Port: ${DB_PORT}
  Username: "${DB_USERNAME}"
  Password: "${DB_PASSWORD}"
  DBName: "${DB_NAME}"

Files:
  Path: "${FILES_PATH}"

Jwt:
  Secret: "${JWT_SECRET}"
EOF

# Log enough to confirm the file landed, but never the secrets.
echo "[entrypoint] wrote /app/configurations/App_dev.yaml (APP_HOST=${APP_HOST} APP_PORT=${APP_PORT} DB_HOST=${DB_HOST} DB_NAME=${DB_NAME})"

exec "$@"
