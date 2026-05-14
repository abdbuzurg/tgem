# Phase 3 — Rollback runbook

A single contiguous procedure that returns the system to the pre-cutover
state, from wherever the forward runbook stopped.

The runbook (`03-runbook.md`) has small rollback snippets at the end of
each section that undo *only that section*. This file is for the larger
abort: "we don't trust the cutover, take the whole thing back to legacy."

## Decision flow — what to roll back

Identify the furthest point you reached in the forward runbook, then run
the matching procedure below. The procedures cascade: if you reached §3.7,
run §R1 → §R2 → §R3 → §R4 → §R5 in order. Don't skip steps unless the
step's preconditions are trivially false (e.g. skip §R3 if you never
edited nginx).

| Reached in forward runbook         | Run these sections           |
| ---------------------------------- | ---------------------------- |
| §3.1 pre-flight                    | nothing — no changes made    |
| §3.2 backup                        | nothing — backup is read-only|
| §3.3 build images                  | nothing — host unchanged     |
| §3.4 Goose stamp                   | §R5                          |
| §3.5 docker compose up             | §R4 → §R5                    |
| §3.6 pm2 stop + container binds    | §R2 → §R4 → §R5 (DB optional)|
| §3.7 nginx swap                    | §R1 → §R2 → §R4 → §R5 (DB optional) |
| §3.8 verification failed           | same as §3.7 row              |

Sections in this file:

- **§R1** — restore the host nginx site file (puts SPA back on legacy `dist`).
- **§R2** — restart legacy backend, stop the container (puts API back on pm2).
- **§R3** — optional: roll back the DB schema migrations (lossy for `00002`).
- **§R4** — bring down the compose stack cleanly.
- **§R5** — un-stamp Goose.

## Conventions

Same `$TGEM_*` shell variables as the forward runbook. If you haven't
re-exported them in this rollback shell session:

```sh
export TGEM_REPO=~/tgem
export TGEM_BACKUP=~/backups/tgem-<DATE>             # FILL IN the dated dir
export TGEM_NGINX_SITE=/etc/nginx/sites-available/<file>
export TGEM_PM2_PROCESS=<legacy-backend-name>
export TGEM_DB_USER=<user>
export TGEM_DB_NAME=<db>
```

---

## §R1 — Restore the host nginx site file

```sh
$ sudo cp ${TGEM_BACKUP}/nginx.site.legacy ${TGEM_NGINX_SITE}
$ sudo nginx -t
nginx: ... is successful
$ sudo systemctl reload nginx
```

The legacy `/var/www/dist` is still on disk (not removed during cutover),
so this re-serves it immediately. TLS and certbot directives in the backup
file are intact — same file that was live before.

Verify:

```sh
$ curl -I https://${TGEM_DOMAIN}/
# Should show static-file response headers (Last-Modified, etc.) rather
# than the proxied response.
```

If you stopped here (no API rollback), the SPA is legacy but the API is
still the rewrite. The two are compatible because both speak the same
HTTP envelope contract.

---

## §R2 — Restart legacy backend, stop the container

This step assumes you're at the point in time where `docker compose up`
started the backend container AND `pm2 stop` was run.

### R2.1 Free port 5000

```sh
$ docker compose stop backend
$ docker compose ps backend          # should show "Exit X" or be absent
```

### R2.2 Re-start the legacy

```sh
$ pm2 start ${TGEM_PM2_PROCESS}
$ pm2 save
$ pm2 list                           # should show ${TGEM_PM2_PROCESS} online
```

If pm2 doesn't have the process registered anymore (e.g. you `pm2 delete`d
it accidentally), restore from the saved dump:

```sh
$ cp ${TGEM_BACKUP}/dump.pm2 ~/.pm2/dump.pm2
$ pm2 resurrect
$ pm2 list
```

### R2.3 Verify

```sh
$ curl -i http://127.0.0.1:5000/api/sign-in -X POST \
       -H 'Content-Type: application/json' -d '{}'
```

Expected: HTTP 200 with an envelope. The legacy backend's response shape
is identical to the rewrite's, so you can't distinguish them from the
response alone — confirm via the pm2 log:

```sh
$ pm2 logs ${TGEM_PM2_PROCESS} --lines 20
```

You should see legacy-style log lines (the legacy uses the same GORM
logger.Info, so it's verbose SQL on stdout — but the binary path in the
process listing will be the legacy build, not the container).

---

## §R3 — (Optional) Roll back the DB schema migrations

The rewrite's migrations 00002–00006 were all designed to coexist with the
legacy code. The legacy backend works against the migrated schema because:

- `00002` deleted orphan rows the legacy ignored anyway.
- `00003` updated one `resources.url` row; the legacy reads
  `resources.url` and uses whichever value is there.
- `00004` added new `invoice_counts` rows for `invoice_type='output-out-of-project'`;
  the legacy doesn't query that invoice type so the rows are inert.
- `00005` added entirely new tables (`resource_types`, `permission_actions`,
  `role_grants`, `user_roles`) and one new column (`roles.code`); the
  legacy never reads them.
- `00006` added two new columns (`user_actions.http_method`, `request_ip`)
  and three indexes; the legacy never writes to those columns.

**So in most cases you can skip §R3 entirely.** Running the legacy
backend against the migrated DB is fine.

If you nonetheless want to revert the schema (e.g. an unexpected app-tier
bug surfaces from one of the additions), there are two paths:

### R3.A — Piecewise SQL rollback (lossy for 00002 data)

Run [`01-rollback.sql`](./01-rollback.sql) section by section:

```sh
$ psql -h 127.0.0.1 -U ${TGEM_DB_USER} -d ${TGEM_DB_NAME} \
       -f ${TGEM_REPO}/MIGRATION/01-rollback.sql
```

The file is annotated section-by-section with LOSSLESS vs LOSSY.
`00002`'s data cleanup is unrecoverable except via §R3.B. `00004`'s
counter delete is lossy if the rewrite served traffic for long enough to
issue out-of-project invoices (very unlikely in a rollback window).

### R3.B — Full restore from backup (canonical)

The pre-migration backup taken in `03-runbook.md` §3.2 is the
authoritative pre-migration state. Restoring it loses any data the
rewrite wrote post-cutover.

```sh
# 1. Stop everything that might write to the DB.
$ sudo systemctl stop nginx                                  # halt traffic
$ pm2 stop ${TGEM_PM2_PROCESS} 2>/dev/null || true
$ docker compose stop backend 2>/dev/null || true

# 2. Kick existing app sessions.
$ psql -h 127.0.0.1 -U <postgres_superuser> -d postgres <<SQL
REVOKE CONNECT ON DATABASE ${TGEM_DB_NAME} FROM ${TGEM_DB_USER};
SELECT pg_terminate_backend(pid) FROM pg_stat_activity
 WHERE datname = '${TGEM_DB_NAME}' AND pid <> pg_backend_pid();
SQL

# 3. Drop and recreate the database.
$ dropdb   -h 127.0.0.1 -U <postgres_superuser> ${TGEM_DB_NAME}
$ createdb -h 127.0.0.1 -U <postgres_superuser> -O ${TGEM_DB_USER} ${TGEM_DB_NAME}

# 4. Restore from the dump.
$ pg_restore -h 127.0.0.1 -U <postgres_superuser> -d ${TGEM_DB_NAME} \
             --no-owner --no-privileges \
             ${TGEM_BACKUP}/db.dump

# 5. Re-grant.
$ psql -h 127.0.0.1 -U <postgres_superuser> -d postgres \
       -c "GRANT CONNECT ON DATABASE ${TGEM_DB_NAME} TO ${TGEM_DB_USER};"

# 6. Verify.
$ psql -h 127.0.0.1 -U ${TGEM_DB_USER} -d ${TGEM_DB_NAME} \
       -c "SELECT count(*) FROM users;"
# Compare against the pre-cutover row count (recorded in
# ${TGEM_BACKUP}/preflight-... if §3.1 captured it, or pre-known).

# 7. Now resume the legacy stack.
$ pm2 start ${TGEM_PM2_PROCESS}
$ sudo systemctl start nginx
```

---

## §R4 — Bring down the compose stack

```sh
$ cd ${TGEM_REPO}
$ docker compose down
```

`down` removes containers and the bridge network but **preserves named
volumes** by default. The volumes (`backend_storage`, `backend_files`) are
empty on a fresh install, so leaving them in place costs nothing and
makes a future re-attempt cheaper.

If you want to wipe everything (including the volumes):

```sh
$ docker compose down -v
$ docker image rm tgem-backend:latest tgem-frontend:latest
```

This is destructive of the container's storage/files but does not touch
the host Postgres or the host filesystem outside Docker.

---

## §R5 — Un-stamp Goose

After the schema has been rolled back (§R3) OR the database has been
restored from backup (§R3.B), remove the Goose tracking table:

```sh
$ psql -h 127.0.0.1 -U ${TGEM_DB_USER} -d ${TGEM_DB_NAME} \
       -c "DROP TABLE IF EXISTS public.goose_db_version;"
```

This puts the database back into an "untouched by Goose" state. If you
restored from backup, the table was wiped by `pg_restore` anyway — this
command becomes a no-op, which is fine.

**Important:** if you ran §R3.B (pg_restore), do NOT keep the
goose_db_version table from the post-migration state — it would mark
nothing-applied versions as applied, and a future re-attempt would
silently skip migrations.

---

## §R6 — Post-rollback verification

```sh
# Legacy backend is up
$ pm2 list | grep ${TGEM_PM2_PROCESS}
$ curl -i http://127.0.0.1:5000/api/sign-in -X POST -d '{}' -H 'Content-Type: application/json'

# Container is down
$ docker compose ps                                    # should be empty or all "Exit"

# Nginx is serving the legacy SPA
$ curl -I https://${TGEM_DOMAIN}/
$ curl -s https://${TGEM_DOMAIN}/ | grep -i 'тгэм\|<title>'

# DB is back to legacy shape (only if you ran §R3)
$ psql -h 127.0.0.1 -U ${TGEM_DB_USER} -d ${TGEM_DB_NAME} <<SQL
SELECT count(*) FROM information_schema.tables
 WHERE table_schema='public' AND table_name='goose_db_version';
-- expected 0
SELECT count(*) FROM information_schema.tables
 WHERE table_schema='public' AND table_name IN ('resource_types','permission_actions','role_grants','user_roles');
-- expected 0
SELECT count(*) FROM information_schema.columns
 WHERE table_schema='public' AND table_name='roles' AND column_name='code';
-- expected 0
SQL
```

---

## What you cannot get back from this rollback

Even after a full §R1 → §R5 rollback:

- **Time spent during cutover.** The ~10–30 second port-5000 handoff
  window had failed API requests; users may have hit error toasts. They'll
  retry; no automated recovery is needed.
- **Writes that hit the rewrite if you used §R3.B (pg_restore).** Any data
  the rewrite wrote between cutover and rollback is gone — `pg_restore`
  resets the DB to the pre-cutover snapshot. For brief rollback windows
  this is acceptable. For longer windows, prefer §R3.A (piecewise SQL)
  and accept its lossy notes.
- **Orphan rows that `00002` deleted.** Per its design, those rows were
  unreachable to the application anyway. The legacy backend works fine
  without them.

---

## Post-rollback: what changed on the host

After §R1 → §R5, the host is byte-for-byte equivalent to its pre-cutover
state EXCEPT for:

- `docker-compose.yml`, `env.backend.example`, `.env.backend`,
  `.gitignore`, `MIGRATION/`, `tgem-backend-rewrite/Dockerfile`,
  `tgem-backend-rewrite/docker-entrypoint.sh`,
  `tgem-backend-rewrite/.dockerignore`,
  `tgem-front-rewrite/Dockerfile`, `tgem-front-rewrite/nginx.conf`,
  `tgem-front-rewrite/.dockerignore` — these files exist on disk but are
  inert (no service references them).
- Docker images (`tgem-backend:latest`, `tgem-frontend:latest`) and named
  volumes if `down -v` was not used — also inert.
- `postgresql.conf` and `pg_hba.conf` may have a docker-bridge entry
  added in §3.1.3. **Leaving these in place is harmless** (nothing is
  connecting from the bridge after the container is down). If you want
  to remove them too, edit the files back to their pre-cutover form and
  `systemctl reload postgresql`.

These artifacts are kept on disk specifically so a *second* cutover
attempt is faster — only the actually-mutated files need to be touched.

Decide on day 7 (per the forward runbook's §3.9) whether to keep them or
remove them after a successful cutover.
