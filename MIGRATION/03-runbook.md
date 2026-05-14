# Phase 3 — Cutover runbook

This is the literal command-by-command sequence for moving production from the
legacy pm2-on-host backend + `/var/www/dist` frontend to the Docker-Compose
stack. The system stays live throughout; the only user-visible interruption is
a ~10–30 second window in step §3.6 when API requests fail while pm2 hands
off port 5000 to the container.

Every destructive or hard-to-reverse step has its rollback command directly
below it. The contiguous, "abort everything" version lives in
[`03-rollback.md`](./03-rollback.md).

## Conventions

- `<placeholder>` denotes a value you must substitute. Don't paste it
  literally.
- `$` prefix is the operator's shell prompt; lines without `$` are file
  excerpts or expected output.
- Commands assume the operator is on the production host, in the repo root
  (`/home/<user>/tgem` or wherever you cloned it).
- Run each `psql` command via the existing host Postgres superuser session;
  the database name and user come from the live `App_dev.yaml`.

## Variables to capture at the top of the session

Before running anything, capture these in shell so the rest of the runbook
can reference them:

```sh
# Set once at session start. Adjust paths and names to match your server.
export TGEM_REPO=~/tgem
export TGEM_DATE=$(date +%Y%m%d-%H%M%S)
export TGEM_BACKUP=~/backups/tgem-${TGEM_DATE}
export TGEM_NGINX_SITE=/etc/nginx/sites-available/<file>       # FILL IN
export TGEM_PM2_PROCESS=<legacy-backend-name>                  # FILL IN, e.g. "tgem-backend"
export TGEM_DB_USER=<user>                                     # FILL IN, from live App_dev.yaml
export TGEM_DB_NAME=<db>                                       # FILL IN, from live App_dev.yaml
export TGEM_DOMAIN=<your.public.domain>                        # FILL IN
echo "session: ${TGEM_DATE}, backup: ${TGEM_BACKUP}"
```

If any value is uncertain, **STOP** and verify before continuing.

---

## §3.1 Pre-flight

These checks confirm the host is ready. If any fails, stop and fix before
proceeding.

### 3.1.1 Tooling

```sh
$ docker --version
$ docker compose version
$ psql --version
$ pg_dump --version
$ nginx -v
```

Required: Docker Engine ≥ 24.0 (for `host-gateway` mapping), `docker compose`
v2 (not v1 `docker-compose`), psql/pg_dump matching the running Postgres
major version, nginx already managing the live site.

### 3.1.2 Postgres reachability from the host

```sh
$ psql -h 127.0.0.1 -U ${TGEM_DB_USER} -d ${TGEM_DB_NAME} -c 'select 1'
 ?column?
----------
        1
(1 row)
```

### 3.1.3 Postgres reachable from the Docker bridge

The backend container connects as `host.docker.internal` → host gateway.
Verify two things:

a. `listen_addresses` in `postgresql.conf` accepts connections from outside
   loopback. Either `*` or an entry that includes the docker bridge subnet
   (typically `172.16.0.0/12`):

   ```sh
   $ sudo grep -E '^\s*listen_addresses' /etc/postgresql/*/main/postgresql.conf
   listen_addresses = 'localhost'        # ← needs change if this
   ```

   If it's `localhost` only, edit to `listen_addresses = 'localhost,172.17.0.1'`
   (use the host's docker0 IP — get it with `ip -4 addr show docker0`) or
   `listen_addresses = '*'`. Reload:

   ```sh
   $ sudo systemctl reload postgresql
   ```

   Rollback for this step: revert the file, reload again.

b. `pg_hba.conf` accepts md5/scram auth from the bridge subnet:

   ```sh
   $ sudo grep -E '^host' /etc/postgresql/*/main/pg_hba.conf
   ```

   Look for a line allowing the app user from the docker bridge CIDR. If
   absent, add a line like:

   ```
   host    <db>    <app_user>    172.17.0.0/16    scram-sha-256
   ```

   (Use the exact `<db>` and `<app_user>` names from the live config. Match
   the auth method to the rest of the file — `md5` if that's what's used,
   `scram-sha-256` for modern Postgres.)

   Reload:

   ```sh
   $ sudo systemctl reload postgresql
   ```

   Rollback: revert the line, reload again.

After both, test from a transient Alpine container that mimics the backend's
network position:

```sh
$ docker run --rm --add-host host.docker.internal:host-gateway alpine:3.20 \
      sh -c 'apk add --no-cache postgresql-client >/dev/null && \
             PGPASSWORD=<password> psql \
             -h host.docker.internal -U '"${TGEM_DB_USER}"' -d '"${TGEM_DB_NAME}"' \
             -c "select 1"'
```

Substitute the live password (don't put it in your shell history if you can
avoid it — use a `.pgpass` file or `read -s`).

### 3.1.4 Production schema-drift check (LIGHTWEIGHT)

Run the §A pre-flight queries from `01-migrations.sql`. The minimum subset:

```sh
$ psql -h 127.0.0.1 -U ${TGEM_DB_USER} -d ${TGEM_DB_NAME} \
       -f MIGRATION/01-migrations.sql 2>&1 | \
       grep -A1 -E 'A[1-6]_'
```

Required outputs (per `01-schema-diff.md` §4.1):

- `A1_table_count` → **50**
- `A2_goose_table_absent` → **0**
- `A3_v2_tables_absent` → **0**
- `A4_roles_code_absent` → **0**
- `A5_user_actions_http_method_absent` → **0**

Anything else is a **STOP** condition. Read `01-schema-diff.md` §5 BLOCKERS.

### 3.1.5 Production schema-drift check (THOROUGH — recommended)

`pg_dump` the production schema and compare to the rewrite's frozen baseline:

```sh
$ pg_dump -h 127.0.0.1 -U ${TGEM_DB_USER} -d ${TGEM_DB_NAME} \
          --schema-only --no-owner --no-privileges \
          > ${TGEM_BACKUP}.preflight.sql || mkdir -p ${TGEM_BACKUP} && \
  pg_dump -h 127.0.0.1 -U ${TGEM_DB_USER} -d ${TGEM_DB_NAME} \
          --schema-only --no-owner --no-privileges \
          > ${TGEM_BACKUP}/preflight-live-schema.sql

$ diff -u \
      <(grep -E '^(CREATE TABLE|ALTER TABLE.*ADD COLUMN|CREATE INDEX|CREATE UNIQUE INDEX)' \
            ${TGEM_REPO}/tgem-backend-rewrite/internal/database/migrations/00001_initial_schema.sql | sort) \
      <(grep -E '^(CREATE TABLE|ALTER TABLE.*ADD COLUMN|CREATE INDEX|CREATE UNIQUE INDEX)' \
            ${TGEM_BACKUP}/preflight-live-schema.sql | sort) \
      > ${TGEM_BACKUP}/preflight-schema.diff || true

$ wc -l ${TGEM_BACKUP}/preflight-schema.diff
```

Expected: a small diff. The legacy AutoMigrate sometimes leaves residual
columns from removed model fields; those show up here as `+` lines on the
live side. They are **harmless** — the rewrite ignores them.

A `-` line (something the rewrite expects, missing from live) is a **STOP**.

### 3.1.6 Cutover-window check

The legacy backend runs a cron at `55 23 * * *` in `Asia/Dushanbe` time
that snapshots project progress. Cutting over during this window risks a
half-finished snapshot.

```sh
$ TZ=Asia/Dushanbe date
```

If the current Dushanbe time is between **23:50 and 00:10**, wait until
00:30 local Dushanbe time before continuing. Set a reminder.

### 3.1.7 Record state for the rollback path

```sh
$ mkdir -p ${TGEM_BACKUP}
$ pm2 list                                              > ${TGEM_BACKUP}/pm2.before.txt
$ pm2 save                                                                  # rewrites ~/.pm2/dump.pm2
$ cp ~/.pm2/dump.pm2                                      ${TGEM_BACKUP}/dump.pm2
$ (cd _legacy/tgem-backend  && git rev-parse HEAD)      > ${TGEM_BACKUP}/legacy-backend.sha
$ (cd _legacy/tgem-front    && git rev-parse HEAD)      > ${TGEM_BACKUP}/legacy-front.sha
$ (cd tgem-backend-rewrite  && git rev-parse HEAD)      > ${TGEM_BACKUP}/rewrite-backend.sha
$ (cd tgem-front-rewrite    && git rev-parse HEAD)      > ${TGEM_BACKUP}/rewrite-front.sha
```

---

## §3.2 Backup  (DESTRUCTIVE-SAFE BOUNDARY)

Everything before this step has been read-only. From here on, the system is
about to change. The backup taken here is the single point of truth for
"recover to known-good state."

### 3.2.1 Database

```sh
$ pg_dump -h 127.0.0.1 -U ${TGEM_DB_USER} -Fc ${TGEM_DB_NAME} \
          > ${TGEM_BACKUP}/db.dump
$ ls -lh ${TGEM_BACKUP}/db.dump      # sanity check size
```

Verify restorability without touching the live DB:

```sh
$ pg_restore --list ${TGEM_BACKUP}/db.dump | head -5
```

(Should show a `dbname:` line and table-of-contents entries. If the dump is
corrupt, this errors out.)

### 3.2.2 Frontend static files

```sh
$ sudo tar czf ${TGEM_BACKUP}/var-www-dist.tgz /var/www/dist
$ ls -lh ${TGEM_BACKUP}/var-www-dist.tgz
```

### 3.2.3 Nginx site file (TLS directives included)

```sh
$ sudo cp ${TGEM_NGINX_SITE} ${TGEM_BACKUP}/nginx.site.legacy
# If the enabled link is a separate file:
$ sudo readlink -f /etc/nginx/sites-enabled/$(basename ${TGEM_NGINX_SITE}) \
      > ${TGEM_BACKUP}/nginx.enabled.target
```

### 3.2.4 Live App_dev.yaml (legacy backend config — contains JWT_SECRET)

You need this both for the .env.backend (§3.3) and for the rollback path.

```sh
$ sudo cp _legacy/tgem-backend/configurations/App_dev.yaml \
        ${TGEM_BACKUP}/App_dev.yaml.legacy
$ sudo chmod 600 ${TGEM_BACKUP}/App_dev.yaml.legacy           # contains secrets
```

If the deployed legacy backend lives in a different path than `_legacy/`,
substitute that path.

### 3.2.5 Verify backup integrity

```sh
$ ls -la ${TGEM_BACKUP}/
total ...
-rw------- 1 ... App_dev.yaml.legacy
-rw-r--r-- 1 ... db.dump
-rw-r--r-- 1 ... dump.pm2
-rw-r--r-- 1 ... legacy-backend.sha
-rw-r--r-- 1 ... legacy-front.sha
-rw-r--r-- 1 ... nginx.site.legacy
-rw-r--r-- 1 ... pm2.before.txt
-rw-r--r-- 1 ... preflight-live-schema.sql      # if §3.1.5 was run
-rw-r--r-- 1 ... rewrite-backend.sha
-rw-r--r-- 1 ... rewrite-front.sha
-rw-r--r-- 1 ... var-www-dist.tgz
```

---

## §3.3 Build images

### 3.3.1 Compose env file

```sh
$ cd ${TGEM_REPO}
$ cp env.backend.example .env.backend
$ chmod 600 .env.backend                  # contains secrets
$ $EDITOR .env.backend
```

Fill in:

- `DB_USERNAME`, `DB_PASSWORD`, `DB_NAME` — copy from
  `${TGEM_BACKUP}/App_dev.yaml.legacy`.
- `JWT_SECRET` — copy from `${TGEM_BACKUP}/App_dev.yaml.legacy`. **MUST
  match** to preserve existing logged-in sessions.
- `DB_HOST` — leave as `host.docker.internal`.
- `APP_HOST` — leave as `0.0.0.0`.
- `APP_PORT` — leave as `5000`.
- `AUTH_PERMISSIONS_ENFORCE` — leave as `0` (log-only).

Sanity check that the file parses:

```sh
$ docker compose config | grep -A20 backend
```

The `environment:` block of the rendered config should reflect your values
(passwords appear in clear in this output — don't share the output).

### 3.3.2 Build

```sh
$ docker compose build
```

This may take 5–10 minutes the first time (the Go build is the long part).
Verify the images exist:

```sh
$ docker images | grep tgem
tgem-backend     latest   ...
tgem-frontend    latest   ...
```

Rollback at this point: nothing — no host state has changed.

---

## §3.4 Apply DB migrations  (FIRST DB MUTATION)

The actual migration runs in two stages:

1. **Goose stamp** — single SQL transaction, applied here by hand.
2. **Migrations `00002`–`00006`** — applied automatically by the backend
   container's embedded `MigrateUp` when it starts in §3.5.

### 3.4.1 Apply the Goose stamp (§B from 01-migrations.sql)

```sh
$ psql -h 127.0.0.1 -U ${TGEM_DB_USER} -d ${TGEM_DB_NAME} -1 <<'SQL'
CREATE TABLE IF NOT EXISTS public.goose_db_version (
    id          serial      PRIMARY KEY,
    version_id  bigint      NOT NULL,
    is_applied  boolean     NOT NULL,
    tstamp      timestamp   NULL DEFAULT now()
);

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
SQL
```

`-1` runs the whole input as a single transaction; if anything fails, the
table doesn't get created and you can retry without partial state.

Verify the stamp:

```sh
$ psql -h 127.0.0.1 -U ${TGEM_DB_USER} -d ${TGEM_DB_NAME} \
       -c "SELECT MAX(version_id) FROM goose_db_version WHERE is_applied;"
 max
-----
   1
```

Expected: `1`.

**Rollback at this point** (before any container starts): drop the goose
table.

```sh
$ psql -h 127.0.0.1 -U ${TGEM_DB_USER} -d ${TGEM_DB_NAME} \
       -c "DROP TABLE public.goose_db_version;"
```

---

## §3.5 Bring up the new stack  (legacy still owns port 5000)

```sh
$ docker compose up -d
```

### 3.5.1 Frontend health (should be immediately green)

```sh
$ docker compose ps frontend
NAME             ...    STATUS         PORTS
tgem-frontend    ...    Up X seconds   127.0.0.1:8080->80/tcp

$ curl -I http://127.0.0.1:8080/
HTTP/1.1 200 OK
Server: nginx/1.27.x
Cache-Control: no-cache, no-store, must-revalidate
Content-Type: text/html

$ curl -s http://127.0.0.1:8080/ | head -5
<!doctype html>
<html lang="ru">
  ...
```

The HTML must be the rewrite's `index.html` — confirm by reading the
shipped title or the hashed asset filenames in the `<script>` tags.

### 3.5.2 Backend status (EXPECTED to be unhealthy until §3.6)

```sh
$ docker compose ps backend
NAME             ...    STATUS                              PORTS
tgem-backend     ...    Restarting (1) X seconds ago         (no ports)

$ docker compose logs --tail=30 backend
[entrypoint] wrote /app/configurations/App_dev.yaml (...)
2026/05/14 ... listen tcp 0.0.0.0:5000: bind: address already in use
```

The "address already in use" error is **expected** — pm2 still owns port
5000. The container is restart-looping; it'll bind successfully the moment
pm2 releases the port in §3.6.

If the log shows anything *other* than that (DB connection error, missing
env var, panic) — **STOP** and resolve before continuing. The most likely
issues:

- DB connection refused → pg_hba.conf still doesn't include the bridge
  CIDR; revisit §3.1.3.
- DB authentication failed → wrong DB_PASSWORD in .env.backend; revisit
  §3.3.1.
- Migration failure (`pq: relation X already exists`) → goose stamp
  didn't take; revisit §3.4.1.

**Rollback at this point:**

```sh
$ docker compose down
```

(named volumes survive; add `-v` if you want to wipe them too — but on a
fresh install they have no data yet, so it doesn't matter)

Then drop the goose table per §3.4.1's rollback.

---

## §3.6 Switch the API to the container  (THE CUTOVER MOMENT)

This is the only step with user-visible impact: API requests fail for
roughly 10–30 seconds while pm2 stops and the container binds.

### 3.6.1 Stop the legacy backend

```sh
$ pm2 stop ${TGEM_PM2_PROCESS}
$ pm2 save
```

### 3.6.2 Wait for the container to bind

```sh
$ for i in 1 2 3 4 5 6 7 8 9 10; do
      docker compose ps backend | grep -q "Up" && break
      sleep 2
  done
$ docker compose ps backend
NAME             ...    STATUS         PORTS
tgem-backend     ...    Up X seconds   127.0.0.1:5000->5000/tcp
```

If after 30 seconds the container still isn't `Up`, check the logs (the
expected error is gone; any new error is a real problem):

```sh
$ docker compose logs --tail=50 backend
```

### 3.6.3 Verify backend is serving

```sh
$ curl -i http://127.0.0.1:5000/api/sign-in -X POST \
       -H 'Content-Type: application/json' \
       -d '{"username":"<invalid>","password":"<invalid>"}'
```

Expected: HTTP 200 with a JSON envelope `{"data":null,"error":"...","success":false,...}`
(the rewrite returns errors via the envelope, not via status code). Any
HTTP 5xx is a problem.

Verify through nginx (still on the old `/var/www/dist` for the SPA, but the
`/api/` block has been pointing at port 5000 the whole time):

```sh
$ curl -i https://${TGEM_DOMAIN}/api/sign-in -X POST \
       -H 'Content-Type: application/json' \
       -d '{"username":"<invalid>","password":"<invalid>"}'
```

### 3.6.4 Verify the migrations ran (00002–00006)

Container's MigrateUp logs should show six "OK" lines (one per migration).
The post-migration verification queries from `01-migrations.sql` §E
confirm the database state:

```sh
$ psql -h 127.0.0.1 -U ${TGEM_DB_USER} -d ${TGEM_DB_NAME} <<'SQL'
SELECT MAX(version_id)               AS max_applied_version FROM goose_db_version WHERE is_applied;
SELECT count(*)                      AS v2_tables           FROM information_schema.tables
   WHERE table_schema='public' AND table_name IN ('resource_types','permission_actions','role_grants','user_roles');
SELECT count(*) AS permission_actions FROM permission_actions;
SELECT count(*) AS resource_types     FROM resource_types;
SELECT count(*) AS roles_with_null_code FROM roles WHERE code IS NULL;
SELECT (SELECT count(*) FROM invoice_counts WHERE invoice_type='output-out-of-project')
     = (SELECT count(*) FROM projects) AS oop_counter_seeded;
SQL
```

Expected:
- `max_applied_version` = 6
- `v2_tables` = 4
- `permission_actions` = 9
- `resource_types` = 40
- `roles_with_null_code` = 0
- `oop_counter_seeded` = `t`

If anything is off, see `01-migrations.sql` §E for the full set of
diagnostic queries.

### 3.6.5 Rollback at this point  (LEGACY → API restore)

API requests are now served by the rewrite; SPA is still legacy
`/var/www/dist`. To go back to "all legacy":

```sh
$ docker compose stop backend                            # frees port 5000
$ pm2 start ${TGEM_PM2_PROCESS}                          # legacy binds 5000 again
$ pm2 save
```

The DB migrations (`00002`–`00006`) have already run. The legacy backend
will work against the migrated schema because the migrations were
additive-only (new tables, new columns, new indexes, plus one one-row
UPDATE on `resources.url`). The legacy code never reads any of the new
columns. **No DB rollback is needed** to revert the *runtime* to legacy.

If you also want to revert the DB schema, see `03-rollback.md`. Note that
`00002` deleted orphan rows that are unrecoverable except via `pg_restore`.

---

## §3.7 Switch the frontend  (last cutover step)

The host nginx is still serving the legacy SPA from `/var/www/dist`. We
swap it to proxy to the frontend container on `127.0.0.1:8080`.

### 3.7.1 Edit the nginx site file

Two `location` blocks change; everything else (TLS, server_name, listen
directives) stays.

```sh
$ sudo $EDITOR ${TGEM_NGINX_SITE}
```

Apply the two changes from [`nginx.conf.new`](./nginx.conf.new):

1. Replace the existing `location / { root /var/www/dist; ... }` block
   with the new `location / { proxy_pass http://127.0.0.1:8080; ... }`
   block.
2. (Optional but recommended) Replace the existing `location /api/`
   block with the new one that adds `client_max_body_size 25m`,
   `proxy_http_version 1.1`, and longer proxy timeouts. The `proxy_pass`
   target stays at `http://127.0.0.1:5000`.

Do NOT touch any `listen`, `ssl_*`, `server_name`, `add_header
Strict-Transport-Security`, or other site-wide directive.

### 3.7.2 Validate and reload

```sh
$ sudo nginx -t
nginx: the configuration file /etc/nginx/nginx.conf syntax is ok
nginx: configuration file /etc/nginx/nginx.conf test is successful

$ sudo systemctl reload nginx
```

If `nginx -t` fails, fix the file before reloading — a failed reload is
a no-op and traffic continues on the old config, but it leaves the file
in a broken state and the next graceful reload will fail too.

### 3.7.3 Verify

```sh
$ curl -I https://${TGEM_DOMAIN}/
HTTP/2 200
server: nginx/...
# The response should now be from the proxied path — no Last-Modified
# header (static-file nginx adds one; the proxied response usually doesn't
# unless the frontend nginx adds it). The Cache-Control header for the
# root path should be "no-cache, no-store, must-revalidate" (set by the
# frontend container's nginx for /index.html).

$ curl -I https://${TGEM_DOMAIN}/assets/index-XXXX.js
HTTP/2 200
cache-control: public, max-age=31536000, immutable
```

(Substitute a real hashed asset filename — get one from the HTML.)

### 3.7.4 Rollback at this point  (NGINX → frontend restore)

```sh
$ sudo cp ${TGEM_BACKUP}/nginx.site.legacy ${TGEM_NGINX_SITE}
$ sudo nginx -t
$ sudo systemctl reload nginx
```

The legacy `/var/www/dist` is untouched on the host disk, so the rollback
is instant. If you've also reverted the API in §3.6.5, that completes the
return to pre-cutover state at the runtime level.

---

## §3.8 Post-cutover verification

Run through `04-smoke-tests.md` in full. Key paths to hit first:

- Log in with a known-good account → check that the token works on the
  new backend (no re-login required → confirms JWT_SECRET is right).
- Open `/invoice/object/:id` for an existing project → confirms polymorphic
  queries still work.
- Visit `/admin/user-actions` as superadmin → confirms 00006 migration and
  the new admin route.
- Trigger one Excel export and one Excel import → confirms templates COPYed
  into image, storage volume writable, and nginx body-size/timeouts are
  sized correctly.

If anything in §3.8 fails, stop and decide: fix forward (most issues are
config tweaks) or roll back (`03-rollback.md`).

---

## §3.9 Cooldown (7 days)

Don't remove any of the following until at least 7 days have passed without
incident:

- `${TGEM_BACKUP}` — entire directory, especially `db.dump` and
  `App_dev.yaml.legacy`.
- The legacy pm2 process (it's stopped, not deleted): `pm2 list` should
  still show it as `stopped`. Don't `pm2 delete <name>` until cooldown
  ends.
- `/var/www/dist` — leave on disk. Disk is cheap; rollback time is not.
- The legacy backend binary at `_legacy/tgem-backend/` — don't delete.
- The pre-migration App_dev.yaml on the production server (its container
  equivalent is now generated by the entrypoint).

After 7 days, an optional cleanup runbook (`05-cleanup.md`) covers
removing all of the above. That file does not yet exist in this migration
package; create it on day 7 once you've decided the cutover is permanent.

---

## Quick reference — full sequence

```sh
# Pre-flight + setup (no mutations)
export TGEM_REPO=~/tgem TGEM_DATE=$(date +%Y%m%d-%H%M%S) TGEM_BACKUP=~/backups/tgem-${TGEM_DATE} \
       TGEM_NGINX_SITE=/etc/nginx/sites-available/<file> TGEM_PM2_PROCESS=<name> \
       TGEM_DB_USER=<user> TGEM_DB_NAME=<db> TGEM_DOMAIN=<domain>
docker --version && docker compose version && psql --version
psql -h 127.0.0.1 -U ${TGEM_DB_USER} -d ${TGEM_DB_NAME} -c 'select 1'
psql -h 127.0.0.1 -U ${TGEM_DB_USER} -d ${TGEM_DB_NAME} -f ${TGEM_REPO}/MIGRATION/01-migrations.sql | grep -A1 -E 'A[1-6]_'
TZ=Asia/Dushanbe date

# Backup (defensive)
mkdir -p ${TGEM_BACKUP}
pg_dump -h 127.0.0.1 -U ${TGEM_DB_USER} -Fc ${TGEM_DB_NAME} > ${TGEM_BACKUP}/db.dump
sudo tar czf ${TGEM_BACKUP}/var-www-dist.tgz /var/www/dist
sudo cp ${TGEM_NGINX_SITE} ${TGEM_BACKUP}/nginx.site.legacy
sudo cp _legacy/tgem-backend/configurations/App_dev.yaml ${TGEM_BACKUP}/App_dev.yaml.legacy
pm2 save && cp ~/.pm2/dump.pm2 ${TGEM_BACKUP}/dump.pm2

# Build
cd ${TGEM_REPO}
cp env.backend.example .env.backend && chmod 600 .env.backend && $EDITOR .env.backend
docker compose build

# Migrate (stamp Goose only; container does 2..6)
psql -h 127.0.0.1 -U ${TGEM_DB_USER} -d ${TGEM_DB_NAME} -1 -c "$(sed -n '/§B/,/§C/p' ${TGEM_REPO}/MIGRATION/01-migrations.sql)"

# Start stack (backend will restart-loop until pm2 stops)
docker compose up -d
docker compose ps

# Cutover API
pm2 stop ${TGEM_PM2_PROCESS} && pm2 save
sleep 5 && docker compose ps backend

# Verify §E from 01-migrations.sql
psql -h 127.0.0.1 -U ${TGEM_DB_USER} -d ${TGEM_DB_NAME} -c "SELECT MAX(version_id) FROM goose_db_version WHERE is_applied;"

# Cutover frontend
sudo $EDITOR ${TGEM_NGINX_SITE}                        # splice in MIGRATION/nginx.conf.new blocks
sudo nginx -t && sudo systemctl reload nginx
curl -I https://${TGEM_DOMAIN}/

# Smoke tests
$EDITOR MIGRATION/04-smoke-tests.md
```
