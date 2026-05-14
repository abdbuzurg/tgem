# Characterization tests

Black-box HTTP tests that lock down the current server contract (status quo,
including known bugs) so the migration phases that follow — layer rename,
package moves, Goose migration, GORM → sqlc — each have a stable safety net.

## Prerequisites

- A local Postgres reachable at `127.0.0.1:5432` with the credentials in
  `configurations/App_dev.yaml` (default: `postgres` / `password`) and
  `CREATEDB` rights. The harness drops & recreates a database called
  `tgem_test` on every run.
- Go matching the version in `go.mod`.

## Run

```sh
# from the repo root (the tests `chdir` to it themselves, but invoking from the
# root keeps relative paths in test output sensible)
go test ./test/characterization/...

# rewrite golden files (use after intentional contract changes)
go test ./test/characterization/... -update-golden
```

## Why this lives in `test/characterization/` rather than next to the code

These tests are black-box (HTTP only), need a single shared `TestMain` to
bring up Postgres + the Gin router once per run, and seed fixtures that have
no business in production packages. A top-level `test/` directory keeps the
boundary obvious and avoids duplicating `TestMain` per package.

## Locked-in behaviors (bugs we deliberately do not work around)

- **Permission middleware is a no-op.** `api/middleware/permission.go` always
  passes through. Tests don't seed permissions for endpoints that mount it.
- **All responses use HTTP 200.** Errors are signalled via the
  `{success: false, error: "..."}` envelope from `pkg/response`. The HTTP
  helpers fail the test if any endpoint returns non-200 — that would itself
  be a contract change worth catching.

## Contract changes during the migration

If a sqlc rewrite legitimately changes a behavior we want to keep (rare) or
fixes a locked-in bug (rarer in phases 2-6, common in phase 7), update the
goldens with `-update-golden` and explain *why* in the same commit. The diff
on the golden file IS the contract change.
