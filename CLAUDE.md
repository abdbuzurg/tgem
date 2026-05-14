# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository layout

Two independent packages, each with its own git repo, `CLAUDE.md`, and `PROGRESS.md`:

- `tgem-backend/` — Go 1.25 API server (Gin + GORM + sqlc/pgx, PostgreSQL). See `tgem-backend/CLAUDE.md` for the layered-rewrite plan, the application-level polymorphism patterns that GORM `Preload` cannot follow, and the seven invoice flavors.
- `tgem-front/` — React 18 + TypeScript SPA (Vite, Feature-Sliced Design, react-query v4, MUI). See `tgem-front/CLAUDE.md` for the FSD layer rules enforced by `eslint-plugin-boundaries`, the routing/permission scheme, and feature-internal patterns (unified mutation modal, config-driven object pages, `useScrollPaginated`).

The repo is **not** a git repo at the root; treat each package as its own working tree. There is no monorepo tooling (no workspaces, no shared `package.json`, no Makefile that spans both).

## Domain (one-paragraph orientation)

Russian-language ERP for electrical-infrastructure construction: warehouse → teams → physical objects on site (KL04KV, MJD, SIP, STVT, TP, Substation, SubstationCell). Almost every reference table and every invoice has Excel template/import/export endpoints — Excel I/O is a first-class feature, not a nice-to-have. Most queries are scoped by `projectID` taken from the auth context.

## Cross-cutting contract (frontend ↔ backend)

These are the conventions that bind the two packages and must not be broken on either side:

- **Response envelope.** Every endpoint returns HTTP 200 with `{data, error, success, permission}`. Errors are signalled by `success: false` + `error` string, **not** by HTTP status. The backend's response helper and the frontend's `IApiResponseFormat<T>` (`@shared/api/envelope.ts`) are the two ends of this contract.
- **Auth.** Token in `localStorage["token"]`, sent as `Authorization: Bearer …` by the axios interceptor in `src/shared/api/client.ts`. Backend handlers behind `middleware.Authentication()` read `userID`, `projectID`, `workerID`, `roleID` from the gin context.
- **API base URL.** Frontend reads `VITE_API_BASE_URL` from `.env` / `.env.production`; default dev value is `http://localhost:5000/api`, matching `App.Port: 5000` in `tgem-backend/configurations/App_dev.yaml`.
- **Permissions.** The frontend matches the **last path segment** of the route against `permission.resourceName`; if it's purely numeric (a runtime id like `/invoice/object/123`), it falls back to the second-to-last. Keep dynamic-route resource names on the segment immediately before `:id` when adding routes.

## Common commands

### Backend (`cd tgem-backend`)

```sh
go run ./cmd/api                                 # start the server on :5000 (needs Postgres at 127.0.0.1:5432, db "tgem")
go test ./test/characterization/...              # HTTP-level characterization tests; harness drops & recreates "tgem_test"
go test ./test/characterization/... -update-golden   # rewrite golden files after an intentional contract change
go run ./cmd/dump_baseline_schema                # one-shot: regenerate the phase-5 Goose baseline from GORM AutoMigrate
sqlc generate                                    # regenerate internal/db/*.sql.go from internal/db/queries/*.sql + migrations
```

Migrations live in `internal/database/migrations/` and are applied by Goose at startup (`MigrateUp` in `internal/database/database.go`). `AutoMigrate` is retained but no longer called at boot — it exists only for `cmd/dump_baseline_schema`.

### Frontend (`cd tgem-front`)

```sh
npm run dev      # Vite dev server
npm run build    # tsc (typecheck) + vite build — fails on type errors
npm run lint     # eslint --max-warnings 0; FSD boundary violations fail this
npm run preview  # serve the production build
```

There is no test runner on the frontend — `npm test` is not defined. Verify behavior manually; do not claim tests pass.

## Locked-in backend quirks worth knowing before debugging

These are documented as deliberate, not accidents (see `tgem-backend/test/characterization/README.md`):

- **Permission middleware is a no-op.** `api/middleware/permission.go` always passes through; gating happens client-side via `RequirePermission`.
- **GORM is in `logger.Info` mode.** Every SQL statement is logged to stdout — keep this in mind when grepping server output.
- **Cron jobs run in `Asia/Dushanbe`.** `internal/jobs/jobs.go` snapshots project progress nightly at `55 23 * * *` local time.

## Where to look for recent context

Each package keeps its own `PROGRESS.md`:

- `tgem-backend/PROGRESS.md` — the seven-phase rewrite (controller→handlers, service→usecase, GORM→sqlc, Goose baseline, dedup pass).
- `tgem-front/PROGRESS.md` — the FSD migration (phases 0–9 + post-migration cleanup).

When in doubt about *why* something is shaped the way it is, those are the primary references before reading commit history.
