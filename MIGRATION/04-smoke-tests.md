# Phase 4 — Smoke-test checklist

Hit these flows in a browser against the production URL right after
finishing `03-runbook.md`. The list is ordered: highest-leverage first
(an early failure tells you the cutover is broken; a late failure tells
you a specific feature is broken).

Mark each line as you go. Anything that fails: capture the URL, the
response payload (browser devtools → Network), and the backend logs
(`docker compose logs --tail=200 backend`). Decide fix-forward vs.
rollback per `03-rollback.md`.

Use at least two accounts so role-gated paths are covered:

- a **superadmin** (sees everything; uses `/admin/*` paths)
- a **non-admin** role with field access (e.g. `warehouse_keeper`,
  `pto`, or `supervisor`)

---

## §A — Sanity (must pass before testing anything else)

- [ ] **Home page loads.** `https://<domain>/` renders the React SPA with
      the login screen. View-source shows hashed asset filenames; Network
      tab shows them as 200 with `cache-control: public, max-age=31536000,
      immutable`.
- [ ] **API health via direct curl.** From the production host:
      ```sh
      curl -i http://127.0.0.1:5000/api/sign-in -X POST \
           -H 'Content-Type: application/json' -d '{}'
      ```
      Returns HTTP 200 with a `{"data":...,"error":"...","success":false}`
      envelope. (The rewrite signals errors via the envelope, not via
      HTTP status — so a 500 here is a real failure.)
- [ ] **API via the public origin.** Same curl against
      `https://<domain>/api/sign-in` — same result.
- [ ] **Backend container is healthy.** `docker compose ps backend` shows
      `Up` (not `Restarting`). `docker compose logs --tail=50 backend`
      shows the migration banner (six Goose lines) and no panics.
- [ ] **Frontend container is healthy.** `docker compose ps frontend`
      shows `Up`. `docker compose logs --tail=20 frontend` shows nginx
      worker process started, no errors.

If any of these fail, **STOP smoke-testing and triage.**

---

## §B — Auth and permissions

Critical because a JWT_SECRET mismatch breaks every user instantly.

- [ ] **Existing logged-in session survives.** If you had a browser tab
      logged in pre-cutover, refresh it. The user should still be logged
      in (no redirect to `/`). This confirms `JWT_SECRET` in
      `.env.backend` matches the production `App_dev.yaml` value.
- [ ] **Login with known-good credentials.** As superadmin and as a
      regular role. Expect: redirect to `/home` (or `/admin/home` for
      admin), `localStorage["token"]` populated.
- [ ] **Logout clears the token.** Click logout, refresh — should land
      back on `/`.
- [ ] **Wrong password is rejected with the envelope error.** The
      backend returns `success:false`, the UI shows a toast.
- [ ] **Permission denied on a forbidden route.** As a non-admin, navigate
      to `/admin/users`. Should land on `/permission-denied`. This
      confirms `AuthProvider.hasPermission` is still reading legacy
      permissions correctly.
- [ ] **Permission enforcement env flag is OFF.** Confirm with:
      ```sh
      docker compose exec backend env | grep AUTH_PERMISSIONS_ENFORCE
      AUTH_PERMISSIONS_ENFORCE=0
      ```
      (We want `0` at cutover — toggle later after verifying all roles
      have the grants they need.)

---

## §C — Invoice flavors (the seven core flows)

Each invoice flavor has its own page, form, table, export, and import.
Hit the *view* path for each at minimum. If you have time, create + open
+ confirm one of each.

- [ ] **Input** — `/invoice/input` loads, lists existing invoices,
      pagination works, open one to see line items.
- [ ] **Output (in-project)** — `/invoice/output-in-project` loads.
- [ ] **Output (out-of-project)** — `/invoice/output-out-of-project`
      loads. If you create a new one, the assigned **delivery code starts
      with "ОВ"** (post-`00004` migration prefix) and the counter row in
      `invoice_counts` for `('output-out-of-project')` increments — not
      the regular output counter. Verify with:
      ```sh
      psql -c "SELECT count, invoice_type FROM invoice_counts
               WHERE project_id=<your_project> ORDER BY invoice_type;"
      ```
- [ ] **Return from team** — `/invoice/return-team` loads.
- [ ] **Return from object** — `/invoice/return-object` loads.
- [ ] **Correction (operator)** — `/invoice/correction` loads. **This
      route was broken in production pre-migration** (00003 fix) — if you
      could open it before, the fix is irrelevant; if you couldn't,
      verify it now opens.
- [ ] **Object** — `/invoice/object` (or `/invoice/object/paginated`)
      loads. Open `/invoice/object/<existing-id>` and verify materials
      and operations render (this hits the application-level polymorphic
      joins).
- [ ] **Writeoff (warehouse)** — `/invoice/writeoff/warehouse` loads.

---

## §D — Reference books (seven physical-object kinds + tables)

- [ ] **Workers** — `/reference-book/worker` lists workers. Open the
      Excel template download (per-entity download button) — should
      stream a `.xlsx` file with a non-zero size. Open it; should not be
      corrupt.
- [ ] **Teams** — `/reference-book/team`.
- [ ] **Materials** — `/reference-book/material`. Try the **Excel import**:
      pick the template, fill in 1–2 rows, submit. Verify rows appear.
      (This exercises `client_max_body_size 25m` in the host nginx and
      the writable storage volume.)
- [ ] **Material costs** — `/reference-book/material-cost`.
- [ ] **Operations** — `/reference-book/operation`.
- [ ] **Districts** — `/reference-book/district`.
- [ ] **Objects → KL04KV** — `/reference-book/object/kl04kv` lists.
- [ ] **Objects → MJD** — `/reference-book/object/mjd`.
- [ ] **Objects → SIP** — `/reference-book/object/sip`.
- [ ] **Objects → STVT** — `/reference-book/object/stvt`.
- [ ] **Objects → TP** — `/reference-book/object/tp`.
- [ ] **Objects → Substation** — `/reference-book/object/substation`.
- [ ] **Objects → Substation Cell** — `/reference-book/object/substation-cell`.

The seven object kinds exercise polymorphic queries via `objects.type` +
`objects.object_detailed_id`. If one loads but another 404s or returns
empty when data exists, it's a polymorphism bug, not a migration bug.

---

## §E — Reporting / statistics

- [ ] **Report menu** — `/report` lists report types.
- [ ] **Material balance report** — generate an Excel; expect a download.
      Large projects trigger long-running queries (this is what the host
      nginx `proxy_read_timeout 300s` protects).
- [ ] **Statistics** — `/statistics` loads charts and tables.

---

## §F — Admin

Run as superadmin.

- [ ] **Users** — `/admin/users` lists users.
- [ ] **Workers** — `/admin/workers`.
- [ ] **Projects** — `/admin/project`.
- [ ] **User actions audit** — `/admin/user-actions`. **NEW IN REWRITE**
      (00006 migration). This page lists the `user_actions` table with the
      new `http_method` and `request_ip` columns filled in for rows
      created post-cutover (NULL for pre-cutover rows). Filter by user
      and by date; both filters use the post-migration indexes.

If `/admin/user-actions` returns 404 — the route guard is rejecting your
role. Verify your role has `view:admin.user_action` in the v2 grants:

```sh
psql -c "SELECT rg.* FROM role_grants rg
         JOIN roles r ON r.id=rg.role_id
         WHERE r.code='superadmin' AND rg.resource_type_code='admin.user_action';"
```

Should return one row. If not, `00005` didn't seed superadmin grants —
investigate before continuing.

---

## §G — HR / attendance / aux

- [ ] **HR attendance** — `/hr/attendance` loads.
- [ ] **Material location live** — `/material-location-live` loads.
- [ ] **Import (bulk)** — `/import` loads and shows the available
      import types.
- [ ] **Auction (private)** — `/auction/private` loads (no layout
      wrapper — direct page).
- [ ] **Auction (public)** — `/auction/public` loads.

---

## §H — Migration-sensitive checks

These verify the specific data transformations from `00002`–`00006`.

- [ ] **No orphan objects.** From psql:
      ```sql
      SELECT type, count(*) AS orphans FROM objects o
      WHERE NOT EXISTS (
          SELECT 1 FROM kl04_kv_objects   WHERE id=o.object_detailed_id AND o.type='kl04kv_objects'
          UNION ALL
          SELECT 1 FROM mjd_objects       WHERE id=o.object_detailed_id AND o.type='mjd_objects'
          UNION ALL
          SELECT 1 FROM s_ip_objects      WHERE id=o.object_detailed_id AND o.type='sip_objects'
          UNION ALL
          SELECT 1 FROM stvt_objects      WHERE id=o.object_detailed_id AND o.type='stvt_objects'
          UNION ALL
          SELECT 1 FROM tp_objects        WHERE id=o.object_detailed_id AND o.type='tp_objects'
          UNION ALL
          SELECT 1 FROM substation_objects        WHERE id=o.object_detailed_id AND o.type='substation_objects'
          UNION ALL
          SELECT 1 FROM substation_cell_objects   WHERE id=o.object_detailed_id AND o.type='substation_cell_objects'
      ) GROUP BY type;
      ```
      Expected: 0 rows returned (or only types not in the LEFT JOIN above,
      which is also fine).
- [ ] **`invoice_counts` covers every (project, type) pair.**
      ```sql
      SELECT count(*) AS missing FROM projects p
      WHERE NOT EXISTS (SELECT 1 FROM invoice_counts ic
                        WHERE ic.project_id=p.id AND ic.invoice_type='input');
      -- repeat with 'output', 'output-out-of-project', 'return', 'writeoff'
      ```
      Each should return `0`.
- [ ] **No duplicate `material_locations` for out-of-project entries.**
      ```sql
      SELECT project_id, material_cost_id, location_id, count(*)
      FROM material_locations
      WHERE location_type='out-of-project'
      GROUP BY project_id, material_cost_id, location_id
      HAVING count(*) > 1;
      ```
      Expected: 0 rows.
- [ ] **`auction_participant_prices` is now unique.**
      ```sql
      SELECT auction_item_id, user_id, count(*)
      FROM auction_participant_prices
      GROUP BY auction_item_id, user_id
      HAVING count(*) > 1;
      ```
      Expected: 0 rows.
- [ ] **Correction resource URL.**
      ```sql
      SELECT name, url FROM resources WHERE name='Корректировка оператора';
      ```
      Expected: `url = '/correction'` (or 0 rows if that resource was
      renamed/removed in production — both are fine).
- [ ] **Every existing user has a `user_roles` row.** Per `00005`'s
      backfill:
      ```sql
      SELECT count(*) FROM users u
      WHERE u.role_id IS NOT NULL
        AND NOT EXISTS (SELECT 1 FROM user_roles ur
                        WHERE ur.user_id=u.id AND ur.project_id IS NULL);
      ```
      Expected: `0`.
- [ ] **Every role has a non-null `code`.**
      ```sql
      SELECT count(*) FROM roles WHERE code IS NULL;
      ```
      Expected: `0`. Roles whose name didn't match the canonical patterns
      got `role_<id>` fallbacks — also fine; verify with:
      ```sql
      SELECT id, name, code FROM roles ORDER BY id;
      ```

---

## §I — Operational (not user-facing)

- [ ] **Compose log volume is sane.** After 30 minutes of normal traffic:
      ```sh
      docker compose logs --no-log-prefix backend | wc -l
      ```
      Expect tens of thousands of lines (GORM logger.Info is verbose) but
      logging caps (`max-size: 10m`, `max-file: 5`) keep disk bounded.
- [ ] **Cron job timezone.** Manually check the cron registration:
      ```sh
      docker compose logs backend 2>&1 | grep -i 'dushanbe\|cron\|location'
      ```
      Expected: at least one log line referencing `Asia/Dushanbe` at
      startup. If you see "unknown time zone Asia/Dushanbe", the
      `-tags timetzdata` build tag didn't take — that's a build bug.
- [ ] **Disk space.** `df -h ~` shows enough headroom for at least a
      week of logs and the named docker volumes.

---

## §J — User-facing UX delta

Items where the rewrite diverges visibly from the legacy. Confirm they
behave as documented, not as users remember from yesterday.

- [ ] **Permission strictness.** Some routes accessible to all roles in
      the legacy now show `/permission-denied` in the rewrite. Confirm
      with a non-admin account: try `/admin/users`, `/report/statistics`.
      These should land on `/permission-denied` (in the legacy they
      might have rendered, then 403'd at the API level).
- [ ] **Page titles and labels are in Russian.** The UI language is
      Russian throughout — confirm a sample of pages render correctly
      (Cyrillic in MUI components, no `???` placeholders).
- [ ] **Auction routes have no layout chrome.** `/auction/public` and
      `/auction/private` render without the main app's sidebar/header.
      This is intentional.

---

## After every box is ticked

- [ ] Add a row to your incident log: cutover date, the
      `${TGEM_BACKUP}` path, observed issues, fixes applied.
- [ ] Set a calendar reminder for day-7 cleanup (`03-runbook.md` §3.9).
- [ ] Consider toggling `AUTH_PERMISSIONS_ENFORCE=1` in `.env.backend`
      after a week of normal traffic without permission-related issues.
      Restart with `docker compose up -d backend` — this is a separate
      runbook entry, not part of cutover.
