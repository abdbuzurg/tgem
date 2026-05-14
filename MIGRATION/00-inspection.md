# Phase 0 — Inspection report

Goal: capture everything about the legacy stack and the rewrite stack that the migration plan will depend on. Nothing here changes code or production state.

The migration prompt and the project root `CLAUDE.md` use directory names that don't match the working tree. Actual layout is:

| What the prompt calls it          | Actual path                                  |
| --------------------------------- | -------------------------------------------- |
| `_legacy/tgem-back`               | `_legacy/tgem-backend`                       |
| `_legacy/tgem-front`              | `_legacy/tgem-front`                         |
| `tgem-backend-rewrite`            | `tgem-backend-rewrite`                       |
| `tgem-frontend-rewrite`           | `tgem-front-rewrite`                         |
| (root CLAUDE.md) `tgem-backend/`  | `tgem-backend-rewrite/`                      |
| (root CLAUDE.md) `tgem-front/`    | `tgem-front-rewrite/`                        |

All artifacts in this migration use the actual paths.

The repo root has no `.git` directory; each of the four trees has its own. No monorepo tooling. This will affect the `docker compose` build context paths (they reference `./tgem-backend-rewrite` and `./tgem-front-rewrite`).

---

## A. Legacy backend — `_legacy/tgem-backend`

This is **what is currently running in production under pm2.**

| Item              | Value                                                                       |
| ----------------- | --------------------------------------------------------------------------- |
| Go version        | 1.20 (`_legacy/tgem-backend/go.mod:3`)                                      |
| Entry point       | `_legacy/tgem-backend/main.go` (top-level `main`)                           |
| HTTP port         | `127.0.0.1:5000`, via `viper.GetInt("App.Port")`                            |
| Bind              | localhost only, so host nginx is already the only public ingress            |
| Route prefix      | `/api`, mounted at `_legacy/tgem-backend/api/router.go:32`                  |
| Total routes      | ~259 (`grep -cE '(GET\|POST\|PUT\|DELETE\|PATCH)\(' router.go`)             |
| Healthcheck       | **None**                                                                    |
| Process manager   | pm2 (no `ecosystem.config.js` in tree — process name must be read from server) |

### Config

Viper loads `App_dev.yaml` from `./configurations` (hardcoded path in `_legacy/tgem-backend/pkg/config/config.go`). No env-var override is wired in. The committed dev config:

```yaml
App:
  Port: 5000
Database:
  Host: "127.0.0.1"
  Port: 5432
  Username: "postgres"
  Password: "password"
  DBName: "tgem"
Files:
  Path: "./files"
Jwt:
  Secret: "q1w2e3r4t5y6"
```

`configurations/` contains only `App_dev.yaml` and `permissions.json`. There is no committed prod config. **The production server must have a hand-edited `App_dev.yaml` with the real DB password and JWT secret.** That file is the authoritative source of two values the migration depends on (DB password and JWT secret) — we'll read them from disk in the runbook, not from this repo.

### DB driver and schema

- Driver: `gorm.io/driver/postgres`, single GORM connection.
- DSN: `host=%s user=%s password=%s dbname=%s port=%d sslmode=disable` (`_legacy/tgem-backend/pkg/database/database.go:20`).
- **Schema source of truth: GORM `AutoMigrate` on 40+ model structs, called at every startup** (`pkg/database/database.go:39`). No migration files in the tree.
- After AutoMigrate, three SQL seed files run if empty (`pkg/database/seed/{project_dev,resource,superadmin}.sql`).

Implication: **the schema currently present in production is not knowable from this repo alone.** GORM's `AutoMigrate` adds columns/tables but never drops them, so if any models were removed or fields renamed in the legacy's own git history, the live DB still has those columns. The only authoritative source is `pg_dump --schema-only` against the live database. Phase 1 must start there.

### Auth

JWT, HMAC-SHA256, secret from `viper.GetString("Jwt.Secret")`. 10-hour TTL. Payload:

```go
type Payload struct {
    jwt.StandardClaims
    UserID    uint
    WorkerID  uint
    RoleID    uint
    ProjectID uint
}
```

Middleware reads `Authorization: Bearer <token>` and sets `userID`, `projectID`, `workerID`, `roleID` in the gin context.

### Cron / background jobs

`internal/jobs/jobs.go` runs in `Asia/Dushanbe`. Per the cross-cutting CLAUDE.md, the daily project-progress snapshot runs at `55 23 * * *` (legacy and rewrite both).

---

## B. Rewrite backend — `tgem-backend-rewrite`

| Item              | Value                                                                       |
| ----------------- | --------------------------------------------------------------------------- |
| Go version        | 1.25.7 (`tgem-backend-rewrite/go.mod:3`)                                    |
| Entry point       | `cmd/api/main.go`                                                           |
| HTTP port         | `127.0.0.1:5000` (`cmd/api/main.go:44`, `viper.GetInt("App.Port")`)         |
| Bind              | localhost only — same as legacy                                             |
| Route prefix      | `/api`, mounted at `internal/http/router.go:43`                             |
| Healthcheck       | **None**                                                                    |
| CGO               | **No.** Pure-Go SQLite (`modernc.org/sqlite`), pgx, GORM. distroless-static is viable. |

### Config

Same `App_dev.yaml`-from-`./configurations` setup, hardcoded in `tgem-backend-rewrite/internal/config/config.go`. The rewrite's own `CLAUDE.md` calls out that env-var override is a *target* for a later phase, **not implemented yet.** So the rewrite container needs the same yaml file with values appropriate to the container environment.

There is one new env var the rewrite reads directly:

- `AUTH_PERMISSIONS_ENFORCE` (`cmd/api/main.go:25`). When set to anything other than `"0"`, permission denials become hard 403s; when unset or `"0"`, denials are logged but the request continues. Default behavior is log-only, matching the legacy's no-op permission middleware.

The committed `configurations/App_dev.yaml` is identical in shape to the legacy's:

```yaml
App:
  Port: 5000
Database:
  Host: "127.0.0.1"
  Port: 5432
  Username: "postgres"
  Password: "password"
  DBName: "tgem"
Files:
  Path: "./files"
Jwt:
  Secret: "q1w2e3r4t5y6"
```

### DB driver and schema

- **Dual connections**: GORM (for migration tooling and a few legacy paths) + pgx/v5 pool (used by sqlc-generated code). DSN format identical to legacy.
- **Schema source of truth: Goose migrations**, embedded into the binary via `//go:embed migrations/*.sql` at `internal/database/migrate.go:11`. `MigrateUp` runs automatically on every startup (`internal/database/database.go:45`).
- `AutoMigrate` is no longer called at boot; it is kept only for `cmd/dump_baseline_schema/main.go`, which is the tool used to regenerate the phase-5 baseline.

Migration files in `tgem-backend-rewrite/internal/database/migrations/`:

```
00001_initial_schema.sql            phase-5 baseline (gorm AutoMigrate snapshot)
00002_phase7_data_cleanup.sql
00003_align_correction_resource_url.sql
00004_split_output_out_of_project_counter.sql
00005_permissions_v2_foundation.sql
00006_user_action_audit.sql
```

`00001_initial_schema.sql` is a full `CREATE TABLE ...` of every table at the time of the rewrite's phase-5 commit. It is **not idempotent**: running it against a database where those tables already exist will fail.

**This is the central migration challenge** (covered in detail in Phase 1):

- Production already has every table in `00001` *plus drift* from AutoMigrate history.
- The rewrite's startup migration code will try to run `00001` because `goose_db_version` doesn't exist yet.
- We need to stamp the Goose version table to say "00001 is applied" before any code path that calls `MigrateUp` runs. Then `MigrateUp` will only consider `00002..00006`, which *are* meant to mutate a real schema.

### sqlc

`sqlc.yaml`:

```yaml
version: "2"
sql:
  - engine: "postgresql"
    schema: "internal/database/migrations"
    queries: "internal/db/queries"
    gen:
      go:
        package: "db"
        out: "internal/db"
        sql_package: "pgx/v5"
        emit_interface: true
        emit_json_tags: true
        emit_empty_slices: true
        emit_pointers_for_null_types: false
```

No bearing on the migration mechanics — generated code is committed.

### Runtime file dependencies

The binary loads Excel templates from `./internal/templates/` via **relative paths** scattered across handlers and usecases. ~25 `.xlsx` files. Files in `tgem-backend-rewrite/internal/templates/`:

```
Invoice Input Report.xlsx
Invoice Output Out Of Project.xlsx
Invoice Output Report.xlsx
Invoice Return Report.xlsx
Invoice Writeoff Report.xlsx
Object Spenditure Report.xlsx
output out of project.xlsx
output.xlsx
return.xlsx
~$Шаблон для импорта Материалов.xlsx          ← Excel lock file, exclude from image
Анализ Остатка Материалов.xlsx
Отчет Остатка.xlsx
Прогресс Проекта.xlsx
Шаблон для импорт Ячеек Подстанции.xlsx
Шаблон для импорта Бригады.xlsx
Шаблон для импорта КЛ 04 КВ.xlsx
Шаблон для импорта МЖД.xlsx
Шаблон для импорта Материалов.xlsx
Шаблон для импорта Подстанции.xlsx
Шаблон для импорта Рабочего Персонала.xlsx
Шаблон для импорта СИП.xlsx
Шаблон для импорта СТВТ.xlsx
Шаблон для импорта ТП.xlsx
Шаблон для импорта Услуг.xlsx
Шаблон импорта ценников для материалов.xlsx
```

Container implications:

1. WORKDIR must be a fixed path (e.g. `/app`). `./internal/templates/` and `./configurations/` must resolve to `/app/internal/templates/` and `/app/configurations/`.
2. Templates are read-only — `COPY` into the image.
3. `storage/import_excel/` is the writable working directory for imports — needs a Docker volume so files survive a container restart and aren't lost between batches.
4. The `~$...xlsx` lock file should be excluded by `.dockerignore` (Office leaves these behind on Windows hosts).

### tz data

`time.LoadLocation("Asia/Dushanbe")` is called in five places:

```
internal/jobs/jobs.go:11
internal/jobs/progress_report_daily.go:54
internal/http/handlers/main_report_handler.go:38
internal/usecase/main_report_usecase.go:93
internal/usecase/main_report_usecase.go:127
```

A distroless-static or scratch image has no `/usr/share/zoneinfo`. The cron registration in `jobs.go` log-fatals on tz lookup failure, so this is a *hard runtime requirement.* Two fixes in the Dockerfile (no code changes):

- Build with `go build -tags timetzdata`, which embeds Go's bundled tzdata into the binary. Recommended — fully self-contained.
- Or use an alpine runtime and `apk add tzdata`.

I'll use `timetzdata` in Phase 2 — it's smaller and removes one moving part.

### CORS

`internal/http/router.go:33` enables `AllowAllOrigins: true` with `AllowCredentials: true`, `MaxAge: 12h`, broad method/header allowlist. After cutover, nginx serves both frontend and `/api/` on the same origin, so CORS is effectively unused. No friction.

### Permission middleware

The root CLAUDE.md says permission middleware is a no-op. Confirmed but with a caveat: it's *opt-in enforce-mode* via `AUTH_PERMISSIONS_ENFORCE` (default off — same effective behavior as legacy). The migration should keep this off until verified post-cutover.

### Existing container artifacts

None. No `Dockerfile`, no `docker-compose.yml`, no `.dockerignore`.

---

## C. Rewrite frontend — `tgem-front-rewrite`

| Item                | Value                                                                |
| ------------------- | -------------------------------------------------------------------- |
| Bundler             | Vite 4.x                                                             |
| Framework           | React 18 + TypeScript + MUI + React Query v4 + React Router v6      |
| Build command       | `npm run build` → `tsc && vite build`                                |
| Output              | `dist/` (default)                                                    |
| Node version        | **Not pinned** (no `engines`, no `.nvmrc`)                           |
| Base path           | `/` (no `base` override)                                             |
| SPA fallback        | Required — non-hash routes, dynamic segments like `/invoice/object/:id` |
| Postbuild           | None                                                                 |

### `package.json` scripts

```json
"dev":     "vite",
"build":   "tsc && vite build",
"lint":    "eslint . --ext ts,tsx --report-unused-disable-directives --max-warnings 0",
"preview": "vite preview"
```

### `vite.config.ts`

```ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@app':      fileURLToPath(new URL('./src/app',      import.meta.url)),
      '@routes':   fileURLToPath(new URL('./src/routes',   import.meta.url)),
      '@features': fileURLToPath(new URL('./src/features', import.meta.url)),
      '@entities': fileURLToPath(new URL('./src/entities', import.meta.url)),
      '@shared':   fileURLToPath(new URL('./src/shared',   import.meta.url)),
    },
  },
})
```

### API base URL

Read once at module-load in `src/shared/api/client.ts:6`:

```ts
axiosClient.defaults.baseURL = import.meta.env.VITE_API_BASE_URL;
```

**Build-time injection only.** No `window.__CONFIG__`, no `config.js`, no runtime templating. The bundle that ships is hard-bound to whatever value of `VITE_API_BASE_URL` was set when `vite build` ran.

`.env` and `.env.production` exist in the tree but are **not readable by me** due to permission scope. The root `CLAUDE.md` says the dev default is `http://localhost:5000/api` and that prod reads from `.env.production`.

**Decision for Phase 2:** build the container with `VITE_API_BASE_URL=/api` (relative, same-origin). The shipped bundle then works under any host nginx (HTTP, HTTPS, any `server_name`) because all API requests resolve to whatever origin served `index.html`. This sidesteps the "build per environment" tax entirely.

This is also the value implied by the current production nginx block: it terminates TLS for `domain.name` and proxies `/api/` to localhost:5000. The frontend in the container doesn't need to know the public hostname.

### Auth

Token in `localStorage["token"]`. Set on login (`src/features/system/login/LoginPage.tsx`), attached to every request via the axios request interceptor (`src/shared/api/client.ts:10-16`).

Because the JWT is verified server-side and the secret is the only thing that matters for compatibility, **frontend cutover does not invalidate sessions** as long as the rewrite backend is given the same `Jwt.Secret` as the legacy. See Section E.

### Routes

Mounted at `/`. React Router v6 BrowserRouter. Full route inventory is in §F (smoke-test surface).

### Existing container artifacts

None.

---

## D. Legacy frontend — `_legacy/tgem-front`

| Item            | Value                                                                  |
| --------------- | ---------------------------------------------------------------------- |
| Bundler         | Vite                                                                   |
| Framework       | React + TS + React Router v6                                           |
| Build / output  | `npm run build` → `dist/` → copied to `/var/www/dist` on server        |
| Node version    | Not pinned                                                             |
| API base URL    | **Hardcoded** in `src/services/api/axiosClient.ts`                     |

The hardcoded URL:

```ts
axiosClient.defaults.baseURL =
  process.env.NODE_ENV === "production"
    ? "http://79.141.74.35/api"
    : "http://localhost:5000/api";
```

That `http://79.141.74.35/api` doesn't match the production nginx block (`server_name domain.name`, presumably HTTPS via certbot). Two ways this is reconciled in reality:

- The on-server build is patched before `npm run build`, or
- The deployed `dist/` was built elsewhere with a different URL, or
- The production frontend genuinely calls a different host (the IP) and works around CORS via the legacy's `AllowAllOrigins: true`.

**This means `_legacy/tgem-front` as committed may not match the bytes actually served from `/var/www/dist`.** The migration assumes that the legacy frontend stays in place via the on-disk `dist/` (we don't rebuild it), so the discrepancy doesn't affect the migration mechanics — but anyone trying to rebuild legacy during rollback should be aware.

Routing: BrowserRouter, mounts at `/`, requires `try_files /index.html`. Same as rewrite.

Auth: `localStorage["token"]`, attached via axios interceptor. The interceptor has a logic bug (`token !== undefined || token !== null` is always true), but it's harmless because requests with a `null` token just get rejected by the backend. Rewrite has fixed it. Not a migration concern.

---

## E. Frontend ↔ backend contract — divergences between legacy and rewrite

The legacy production frontend is going to be replaced *at the same time* as the legacy backend, so the only mixed-state we will ever have is "rewrite backend serving the legacy frontend's `dist/`" — which lasts for the seconds between step 3.6 (start rewrite backend) and step 3.7 (switch nginx to the new frontend). Even that brief mix must function so we have a rollback boundary.

Across the four trees:

| Concern                      | Legacy backend       | Rewrite backend                 | Legacy frontend            | Rewrite frontend             | Compatible at cutover? |
| ---------------------------- | -------------------- | ------------------------------- | -------------------------- | ---------------------------- | ---------------------- |
| API prefix                   | `/api/`              | `/api/`                         | calls `/api/...`           | calls `/api/...`             | ✅                     |
| Response envelope            | `{data,error,success,permission}` always 200 | same                            | reads envelope             | reads envelope               | ✅                     |
| Auth scheme                  | JWT HS256, custom Payload | JWT HS256, same Payload     | sends Bearer               | sends Bearer                 | ✅ if same secret      |
| Auth secret                  | `Jwt.Secret` from yaml | `Jwt.Secret` from yaml       | —                          | —                            | ✅ if same value used  |
| Token TTL                    | 10h                  | 10h (verify in Phase 1)         | —                          | —                            | ✅                     |
| `localStorage` token key     | —                    | —                               | `"token"`                  | `"token"`                    | ✅                     |
| CORS                         | `AllowAllOrigins`    | `AllowAllOrigins`               | —                          | —                            | ✅                     |
| Permission enforcement       | Backend no-op; frontend gates routes via `RequirePermission` | Backend opt-in; default no-op | Legacy permits more routes | Stricter `RouteGate`         | ⚠️ see below          |
| Permissions schema           | Tables created by AutoMigrate from legacy models | New migration `00005_permissions_v2_foundation.sql` reshapes permissions | Reads `permission.resourceName` last segment | Same | ⚠️ see Phase 1     |
| API surface                  | ~259 routes          | Refactored handlers; route set largely preserved per CLAUDE.md phase 3 | — | — | likely ✅, must verify routes that legacy frontend hits   |
| `AUTH_PERMISSIONS_ENFORCE`   | n/a                  | optional env var, default off   | —                          | —                            | ✅ (left default)      |
| Backend port                 | 5000                 | 5000                            | hits port 5000 via nginx   | hits `/api` via nginx        | ✅                     |
| WORKDIR-relative paths       | `./pkg/excels/...`   | `./internal/templates/...`      | —                          | —                            | ✅ (different path, but only matters inside each container) |

**Two amber flags** worth tracking through the rest of the migration:

1. **Permissions schema reshape (migration `00005`).** The rewrite introduces a permissions-v2 schema. Existing rows in the legacy permissions tables need to be transformed, not just left in place. Phase 1 must read `00005_permissions_v2_foundation.sql` carefully and decide whether it's a pure schema change (safe) or whether it expects pre-existing data in a specific shape. If a user logs in immediately after migration, their `RoleID` must still resolve to a permission set the new code understands.

2. **Frontend route guards are stricter in the rewrite.** Routes that legacy users could reach (because legacy's permission middleware was a no-op) may show `/permission-denied` in the rewrite if the user's role doesn't have the matching resource. This is a UX behavior change, not a migration bug — but it should be in the smoke-test checklist (Phase 4).

---

## F. Smoke-test surface (preview for Phase 4)

Inventoried for use in `MIGRATION/04-smoke-tests.md`. Grouped by feature domain, with the permission resource shown for routes behind `RequirePermission`.

**Unauthenticated**
- `GET /` — login page
- `GET /404`, `/permission-denied`
- `GET /auction/public`, `/auction/private` — public auction screens (no layout)

**Home**
- `/home`, `/admin/home`

**Invoice** (the seven flavors plus object detail)
- `/invoice/input` (`view:invoice.input`)
- `/invoice/output-in-project` (`view:invoice.output`)
- `/invoice/output-out-of-project` (`view:invoice.output_out_of_project`)
- `/invoice/return-team` (`view:invoice.return_team`)
- `/invoice/return-object` (`view:invoice.return_object`)
- `/invoice/correction` (`view:invoice.correction`)
- `/invoice/object`, `/invoice/object/paginated`, `/invoice/object/add`, `/invoice/object/:id` (`view:invoice.object`, `create:invoice.object`)

**Writeoff / loss**
- `/invoice/writeoff/warehouse`, `/invoice/writeoff/object` (`view:invoice.writeoff`)
- `/invoice/loss/{warehouse,team,object}` (`view:invoice.writeoff`)

**Reference books** (seven physical-object kinds + reference tables)
- `/reference-book/worker`, `/team`, `/material`, `/operation`, `/material-cost`, `/district`
- `/reference-book/object/{kl04kv,mjd,sip,stvt,tp,substation,substation-cell}`

**Reporting / stats**
- `/report` + sub-routes (`view:report.*`)
- `/statistics` (`view:report.statistics`)

**Admin**
- `/admin/users` (`view:admin.user`)
- `/admin/workers` (`view:reference.worker`)
- `/admin/project` (`view:admin.project`)
- `/admin/user-actions` (`view:admin.user_action`) — **new in rewrite**

**Other**
- `/import` (`import:system.import`)
- `/material-location-live` (`view:system.material_location_live`)
- `/hr/attendance` (`view:hr.attendance`)

The Excel template/import/export endpoints are not full routes — they hang off the entity pages above (each has `/document/template`, `/document/export`, `/document/import`). These will be tested via the entity pages.

---

## G. Divergences worth restating, by impact

### Behavior-affecting (could surprise a user during cutover)
- **`/admin/user-actions` is new** in the rewrite. Existing admins may or may not have the `view:admin.user_action` permission row; if not, they see "permission denied" on the audit page. Out of scope to seed — flag in smoke-tests.
- **Stricter route gating** in the rewrite (described above). A `/permission-denied` page that legacy users never saw is the most likely surprise.

### Build-affecting
- **Frontend's `VITE_API_BASE_URL` is build-time only.** Resolved by building with `VITE_API_BASE_URL=/api`.
- **Backend's config is loaded from a fixed yaml path.** Resolved by `COPY`ing a container-tuned `App_dev.yaml` into the image (with `Database.Host: host.docker.internal` and the production JWT secret + DB password supplied at build/runtime).
- **Backend's tzdata** is required. Resolved by `-tags timetzdata` at build time.
- **CGO is not needed.** Distroless-static is viable.
- **WORKDIR-relative paths.** `./internal/templates/` and `./configurations/` must be present under WORKDIR in the container.

### Migration-mechanics-affecting
- **Live schema is the only authoritative legacy schema.** Phase 1 must start from `pg_dump --schema-only` against production, not from inspecting models.
- **Goose `00001` is not idempotent against an existing schema.** Stamp `goose_db_version` to skip it; let `00002..00006` run normally.
- **Permission migration `00005` may transform existing rows.** Read carefully in Phase 1, design a rollback that preserves legacy rows.

### Operational
- **No healthcheck route on either backend.** Compose can use a TCP-port readiness check (`wget`/`curl` against `localhost:5000/api/...` with a known-public route, e.g. login submission with empty body returning a 200 envelope) — but it's simpler to just have `restart: unless-stopped` and verify via curl in the runbook. Phase 2 will pick one.
- **GORM logs every SQL statement** to stdout. Compose's `json-file` driver needs `max-size` / `max-file` caps or the host disk fills.
- **Cron at 23:55 Asia/Dushanbe** mutates rows nightly. Cutover should not coincide with that window; the runbook will say so.

---

## H. BLOCKERS

Things we cannot proceed past Phase 1 without obtaining from production.

1. **Production `Jwt.Secret`** — the actual value from the live `App_dev.yaml` on the production server. The rewrite container must be configured with this same value, otherwise all logged-in users are kicked out at cutover (10-hour rolling, but the cutover itself happens instantly).

2. **Production `Database.{Username, Password, DBName, Host}`** — almost certainly differs from the committed dev values (`postgres`/`password`/`tgem`/`127.0.0.1`). Needed for the rewrite container's `.env.backend`.

3. **`pg_dump --schema-only` of the live database** — the only authoritative legacy schema. Phase 1 needs it for the diff. The runbook will produce it as the first command of the migration; for *planning*, we can use the legacy AutoMigrate output as a proxy, with the understanding that the real diff may reveal drift columns we can't see now.

4. **pm2 process name** for the legacy backend — we need it for `pm2 stop <name>` in step 3.6. Not committed anywhere in the repo.

5. **Exact path of the live nginx site file** (almost certainly `/etc/nginx/sites-available/<something>`). The runbook needs it for the backup and swap steps.

6. **Whether production's nginx has TLS directives** (`ssl_certificate`, `ssl_certificate_key`, `listen 443 ssl`, etc.) added by certbot. The runbook's nginx swap must preserve them — Phase 2 will produce a `nginx.conf.new` that *only contains the changed `location /` block*, with a comment instructing the operator to splice the new block into the live file rather than overwrite it wholesale.

None of these are blockers for Phase 0 → Phase 1 planning; they are blockers for actually running the runbook. Phase 1 can proceed with the committed dev values as placeholders and the runbook will instruct the operator to substitute the real values when filling `.env.backend`.

---

## I. Open questions for the user

Before I write Phase 1, please confirm or correct:

1. The naming differences in §A (especially `tgem-front-rewrite` vs `tgem-frontend-rewrite`, `_legacy/tgem-backend` vs `_legacy/tgem-back`). I'll use the actual on-disk names in all generated files — let me know if you'd prefer to rename directories first.
2. Do you have shell access to run `pg_dump --schema-only` on production *now*, so I can include the live schema in the Phase 1 diff? Or should I plan the diff against the legacy AutoMigrate-derived schema and verify on the server during cutover?
3. The production `Jwt.Secret` value — will you supply this when filling `.env.backend` on the server (the prompt says `.env.backend` is gitignored, so that's the right place), or do you want me to design a one-shot login-everyone-out cutover that doesn't depend on secret continuity?
4. Are the rewrite directories `tgem-backend-rewrite/` and `tgem-front-rewrite/` themselves the migration source-of-truth, or will they be renamed (e.g. to drop the `-rewrite` suffix) before/after cutover? This affects whether the docker-compose `build.context` paths stay or change.

Once these are confirmed, I'll proceed to Phase 1: schema diff and the Goose-baseline-stamp plan.
