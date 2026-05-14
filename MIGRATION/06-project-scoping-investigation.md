# Phase 6 — Cross-project invoice visibility investigation

## Root cause in one sentence

The paginated invoice listings on the rewrite are **all** correctly project-scoped at the SQL layer (every `/paginated` query has `WHERE invoice_*.project_id = $1` and the `$1` is sourced from the JWT-derived `c.GetUint("projectID")` — never from request body, query string, or URL), so a persistent "lists show another project's invoices" symptom is not explained by the rewrite source alone; the two real source-side weaknesses are (a) react-query cache survives logout — stale rows from the previous project's session flash in for the first paint after a project switch — and (b) a handful of "by ID / by delivery code / by invoice id" endpoints (most prominently `GET /invoice-object/:id`, `GET /<flavor>/document/:deliveryCode`, and the `/.../materials/...` lookups) don't filter on `project_id`, so a user holding any invoice's id or delivery code can read it from a different project.

## Fix scope: small / **medium** / large

- **Backend (security hardening):** small — add project_id to the WHERE clauses of ~10 `:one`/`:many` queries that currently key only on id/delivery_code, and assert in the calling handlers that they pass the JWT project. Mechanical change, sqlc regen.
- **Frontend (cache hygiene at logout):** small — one helper that calls `queryClient.clear()` (or removes all queries) in the logout handlers and on LoginPage mount.
- **Diagnostic step to pin "real" symptom:** small but needs server access — see open questions §8.

---

## 1. How the rewrite scopes by project

One mechanism, used uniformly:

1. The JWT (`pkg/jwt/jwt.go` `Payload`) carries `ProjectID`, set at login from `dto.LoginData.ProjectID` (the project the user chose in the dropdown). The login flow validates that the user is a member of that project via `user_in_projects` unless their role's `code == "superadmin"` (`internal/usecase/user_usecase.go:229-271`).
2. `internal/http/middleware/authenication.go:44` calls `c.Set("projectID", payload.ProjectID)` for every authenticated request. Handlers read it via `c.GetUint("projectID")`.
3. The frontend has **no project switcher**. To change projects, the user logs out (`AppLayout.tsx:11-18`) and logs in again with a different project — yielding a new JWT with a different `ProjectID`. Token lives in `localStorage["token"]`; axios interceptor attaches it as `Authorization: Bearer …` (`src/shared/api/client.ts`).
4. The legacy backend did the same (JWT payload includes `ProjectID`; middleware sets it on the gin context). The rewrite preserves the mechanism verbatim.

There is no header-based, body-based, or URL-based project switcher. **For every routed invoice endpoint I inspected, the projectID value reaching the query layer comes from the JWT, not from anything the client can choose to send.**

---

## 2. Per-flavor handler / query / scoping table

For each invoice flavor's listing-style endpoints. (The `/document/:deliveryCode`, `/:id/materials/...`, and similar by-id endpoints are handled separately in §3.) "Source of project_id" indicates where the value in the SQL `$1` ultimately originates. UNCLEAR is reserved for code paths I couldn't fully resolve from source.

| Flavor | Handler (file:line) | Query | Project_id in WHERE? | Source of project_id | Verdict |
|--------|---------------------|-------|----------------------|----------------------|---------|
| input  | `internal/http/handlers/invoice_input_handler.go:68 GetPaginated` | `ListInvoiceInputsPaginatedFiltered` / `ListInvoiceInputsPaginatedByMaterials` | YES (`invoice_inputs.project_id = $1`) | `c.GetUint("projectID")` → `filter.ProjectID` → SQL `$1` | **CORRECT** |
| input  | `:198 Create`, `:221 Update` | `CreateInvoiceInput` / `UpdateInvoiceInput` | n/a (write) | handler overrides `createData.Details.ProjectID = c.GetUint("projectID")` (`:212` for Create, similarly Update) | CORRECT (body value ignored) |
| output | `invoice_output_handler.go:62 GetPaginated` | `ListInvoiceOutputsPaginatedFiltered` | YES | JWT | **CORRECT** |
| output | `:133 Create`, `:375 Update` | `CreateInvoiceOutput` / `UpdateInvoiceOutput` | n/a (write) | JWT (handler overrides body, `:143-144`, `:385-386`) | CORRECT |
| output-out-of-project | `invoice_output_out_of_project_handler.go:43 GetPaginated` | `ListInvoiceOutputOutOfProjectsPaginated` | YES | JWT | **CORRECT** |
| output-out-of-project | `:88 Create`, `:163 Update` | sqlc | n/a (write) | JWT (handler overrides body, `:98-99`, `:173-174`) | CORRECT |
| return | `invoice_return_handler.go:42 GetPaginated` (routes to `GetPaginatedTeam` or `GetPaginatedObject` based on `returnType` query param) | `ListInvoiceReturnsPaginatedTeam` / `ListInvoiceReturnsPaginatedObject` | YES (both: `invoice_returns.project_id = $1`) | JWT (`:80`) | **CORRECT** |
| return | `:108 Create`, `:126 Update` | sqlc | n/a (write) | JWT (`:115-116`, `:133-134`) | CORRECT |
| writeoff | `invoice_writeoff_handler.go:42 GetPaginated` | `ListInvoiceWriteOffsPaginated` | YES (`invoice_write_offs.project_id = $1`) | JWT (`:64-67`) | **CORRECT** |
| writeoff | `:85 Create`, `:104 Update` | sqlc | n/a (write) | JWT (`:92`, `:111`) | CORRECT |
| correction | `invoice_correction_handler.go:41 GetPaginated` | `ListInvoiceCorrectionsPaginated` | YES (`io.project_id = $1`) | JWT (`:70`) | **CORRECT** |
| correction | `:90 GetAll` (route `GET /invoice-correction/`) | `ListInvoiceObjectsForCorrection` | YES (`invoice_objects.project_id = $1`) | JWT (`:92-94`) | **CORRECT** |
| correction | `:170 Create` | sqlc | n/a (write) | JWT inside `createData.Details.ProjectID = c.GetUint("projectID")` at handler — verify | needs spot-check |
| object | `invoice_object_handler.go:54 GetPaginated` | `ListInvoiceObjectsPaginated` | YES (`invoice_objects.project_id = $1`) | JWT (`:69`) | **CORRECT** |
| object | `:150 Create` | sqlc | n/a (write) | JWT (handler overrides body) | CORRECT |

**Bottom line for the paginated lists:** every routed invoice listing is properly scoped at the SQL layer and the project_id flows from the JWT. There is no "missing WHERE clause", no "client-trusted body field", no "URL param feeds project_id". If the user is seeing whole-list cross-project bleed, the cause is not in the listing handlers as written.

Two unused-but-dangerous queries are present in the rewrite source but **not routed**:

- `ListInvoiceOutputs` (no project filter) used by `invoiceOutputUsecase.GetAll()`, which is on the interface but not bound to any route. Same shape exists for input and return. **Dead code at the HTTP level** today, but if anyone wires a `/all` route to those usecase methods in the future, they would leak. Worth deleting or fixing on cleanup pass.

---

## 3. By-id / by-code endpoints that DO leak across projects

These are routed and do return data without checking that the requested row belongs to the user's project. Most are pre-existing — the legacy GORM repo had the same shape — so they're not regressions, but they are real information-disclosure paths that an authenticated user (any project) can hit if they know or guess an invoice id or delivery code.

| Route | Handler | Query | What leaks |
|-------|---------|-------|------------|
| `GET /invoice-object/:id` | `invoice_object_handler.go:36 GetInvoiceObjectDescriptiveDataByID` | `GetInvoiceObjectDescriptiveDataByID` (`invoice_object.sql:25`) — `WHERE invoice_objects.id = $1` | invoice_object header data + descriptive fields, regardless of project |
| `GET /output/document/:deliveryCode` | `invoice_output_handler.go:219 GetDocument` | `GetInvoiceOutputByDeliveryCode` (`invoice_output.sql:20`) — `WHERE delivery_code = $1` (no project) — and then the on-disk xlsx/pdf for that delivery code is streamed | the xlsx/pdf attachment of another project's output invoice; only confirmation flag is read from DB, file content from disk |
| `GET /invoice-output-out-of-project/document/:deliveryCode` | `invoice_output_out_of_project_handler.go:275 GetDocument` | `GetInvoiceOutputOutOfProjectByDeliveryCode` (`invoice_output_out_of_project.sql:7`) — no project filter | same as above for OOP flavor |
| `GET /return/document/:deliveryCode` | `invoice_return_handler.go:208 GetDocument` | `GetInvoiceReturnByDeliveryCode` (`invoice_return.sql:15`) — no project filter | same for return |
| `GET /input/document/:deliveryCode` | `invoice_input_handler.go:307 GetDocument` | no query — file system lookup by deliveryCode | any project's input PDF |
| `GET /invoice-writeoff/document/:deliveryCode` | `invoice_writeoff_handler.go:233 GetDocument` | filesystem glob | any project's writeoff PDF/xlsx |
| `GET /output/:id/materials/without-serial-number` (and `…/with-serial-number`) | `invoice_output_handler.go:97/115` | `ListInvoiceMaterialsWithoutSerialNumbers` / `…WithSerialNumbers` — `WHERE invoice_type = $1 AND invoice_id = $2` (no project) | the line items of any output invoice id |
| Same shape for `input`, `return`, `writeoff`, `output-out-of-project` `/:id/materials/...` | their handlers | same shared queries in `invoice_materials.sql` | same |
| `GET /output/invoice-materials/:id` | `invoice_output_handler.go:?` (GetMaterialsForEdit) | `ListInvoiceOutputMaterialsForEdit` — `WHERE invoice_type = 'output' AND invoice_id = $1` | edit-form line items of any output invoice |
| Same for `input`, `return`, `writeoff`, `output-out-of-project` `/invoice-materials/:id` | their handlers | per-flavor "MaterialsForEdit" queries — none include `project_id` | edit-form line items of any flavor's invoice |
| `GET /invoice-object/object/:objectID` | `invoice_object_handler.go:202 GetTeamsFromObjectID` | `ListInvoiceObjectTeamsByObjectID` — `WHERE object_teams.object_id = $1` | teams attached to any project's object |

The Confirmation handlers (e.g. `invoice_output_handler.go:172 Confirmation`, `invoice_writeoff_handler.go:186 Confirmation`) call `usecase.GetByID(uint(id))` and then accept a PDF upload — without verifying the invoice belongs to the user's project. This means a user from project A can call `POST /output/confirm/<another-project-id>` with a PDF and overwrite the confirmation flag + file. **Privilege escalation pre-existing in legacy too.**

The Update handlers override `details.ProjectID` from the JWT and then issue `UPDATE invoice_outputs SET project_id = $2, … WHERE id = $1` — meaning a user from project A who PATCHes with project B's `id` will silently re-parent that invoice into project A. **Pre-existing bug** (legacy code path is the same), still latent today.

None of the above explains "lists show another project's invoices."

---

## 4. Pattern checklist (the four bug shapes from the prompt)

1. **Missing WHERE on a list** — Searched all `invoice_*.sql` files. Every paginated/list/count query routed at HTTP includes `WHERE … project_id = $1`. The unfiltered `ListInvoiceOutputs` / `ListInvoiceInputs` / `ListInvoiceReturns` queries exist in the SQL files but their usecase callers are not bound to routes. **Not present** on listing endpoints.
2. **Wrong source for project_id** — Audited every handler's body-binding flow. Every Create/Update/Report/Paginated handler overrides any body-supplied `ProjectID` with `c.GetUint("projectID")` (the JWT-derived value) before calling the usecase. **Not present.**
3. **Race condition on active project switch** — In the rewrite there is no in-app project switcher; switching requires logout + login + new JWT. There's no race window from the *backend*. There IS a frontend cache leak (§5) that produces a visual symptom that looks similar — but it's a stale-while-revalidate display, not a true race.
4. **Wrong join (filtering on the wrong table)** — Reviewed the paginated joins. The output paginated query (`invoice_output.sql:56`) joins districts/teams/workers and filters `invoice_outputs.project_id = $1` — correct. Return team/object joins (`invoice_return.sql:60-105`) filter `invoice_returns.project_id = $1` — correct. Correction (`invoice_correction.sql:1`) filters `io.project_id = $1` against the underlying `invoice_objects` — correct. **Not present.**

---

## 5. Frontend cache hygiene (the "soft" cause)

The QueryClient is constructed once in `src/app/providers/ReactQueryProvider.tsx` and never replaced. Defaults: `staleTime: 0`, `cacheTime: 300000` (react-query v4 default).

The logout handlers in `src/app/layouts/AppLayout.tsx:11`, `src/app/layouts/AdminLayout.tsx:10`, and `src/features/auction/private/AuctionPrivatePage.tsx:22` do:

```ts
localStorage.removeItem("token")
localStorage.removeItem("username")
navigate(LOGIN)
```

They **do not** call `queryClient.clear()` or `queryClient.removeQueries()`. The cache survives the logout.

The LoginPage mounts and calls `authContext.clearContext()` which clears in-memory permissions but does not touch the QueryClient.

Query keys do not include the projectID:

- `["invoice-object-user"]` (`features/invoice/object/InvoiceObjectUserPage.tsx:14`)
- `["invoice-input"]` (`features/invoice/object/InvoiceObjectPaginatedPage.tsx:16`)
- `["invoice-return-team"]` (`features/invoice/return-team/InvoiceReturnTeamPage.tsx:21`)
- `["invoice-return-object"]` (`features/invoice/return-object/InvoiceReturnObjectPage.tsx:21`)
- `["invoice-correction", searchParameters]` (`features/invoice/correction/InvoiceCorrectionPage.tsx:22`)
- `["invoice-correction-search-parameters-data"]`, `["invoice-return-codes"]`, `["all-teams-for-select"]`, `["all-districts"]`, `["available-materials"]`, etc.

Consequence: user logs in as project A → opens `/invoice/input` → react-query caches `["invoice-input"]` with project-A rows. User logs out → token removed. User logs back in as project B → opens `/invoice/input`. React-query sees a fresh cache entry under `["invoice-input"]`, immediately renders it (stale-while-revalidate), and fires a network request in the background. **For the first paint, the user sees project A's invoices while logged in as project B.** When the refetch completes (typically a few hundred ms later), the data refreshes to project B's rows. If the user clicks away before the refetch lands, they may genuinely act on project A's rows believing they are project B's.

The legacy frontend's logout has the same code shape (verified at `tgem-legacy/tgem-front/src/components/layout.tsx:43`). So if this is the only cause, it would not be a regression. But the user phrasing — "Users report seeing invoices from other projects when viewing a single project" — fits this symptom much better than any backend leak I can identify.

This is the most plausible source of the observed bug **assuming the user's report is about transient flashes, not persistent appearance**. Confirm with the open question §8.

---

## 6. Per-flavor proposed fix

The fix splits into one shared frontend change and a small backend hardening pass.

### (i) Frontend (single root cause, highest signal/noise ratio)

A. **Clear the cache on logout.** Lift the logout helper into `src/shared/lib/auth/logout.ts`:

```ts
import { useQueryClient } from "@tanstack/react-query"

export function useLogout() {
  const queryClient = useQueryClient()
  return () => {
    queryClient.clear()              // critical: drops cross-project cached data
    localStorage.removeItem("token")
    localStorage.removeItem("username")
  }
}
```

Replace the inline logouts in `AppLayout.tsx`, `AdminLayout.tsx`, and `AuctionPrivatePage.tsx` with `const logout = useLogout()`.

B. **Belt and braces — clear the cache on LoginPage mount too.** Defends against the case where the previous session left a token in place but the user navigates straight to `/login`:

```ts
// LoginPage.tsx, alongside the existing clearContext useEffect
const queryClient = useQueryClient()
useEffect(() => { queryClient.clear() }, [])  // mount-only, same pattern as clearContext
```

C. **(Optional, sturdier) Include the projectID in cache keys.** Read it from the JWT at login (or stash it in localStorage alongside the token) and prefix every invoice/list query key with `[projectID, ...]`. This protects against in-app project switching if it's ever added, and self-documents the dependency. Heavier change — ~30 query-key sites — only worth doing once the simpler clear() lands and the symptom is confirmed gone.

This single backend-free change is enough if §8 confirms the symptom is "flashes of other-project data right after logout/login."

### (ii) Backend hardening (security)

Add project_id to the WHERE clause of every "by id / by code" query that's routed and currently filters only on id. Each is mechanical: add `AND project_id = $N` (sqlc gives you a new param), then thread `c.GetUint("projectID")` through the usecase. The set:

- `invoice_object.sql:25 GetInvoiceObjectDescriptiveDataByID` — add `AND invoice_objects.project_id = $2`. Update `invoice_object_handler.go:36` to read projectID and pass it.
- `invoice_output.sql:20 GetInvoiceOutputByDeliveryCode` — add `AND project_id = $2`. Update `invoice_output_handler.go:219 GetDocument`.
- `invoice_output_out_of_project.sql:7 GetInvoiceOutputOutOfProjectByDeliveryCode` — same fix.
- `invoice_return.sql:15 GetInvoiceReturnByDeliveryCode` — same fix.
- `invoice_input_handler.go:307 GetDocument` — there's no DB query but the filesystem path is `./storage/import_excel/input/{deliveryCode}.pdf` and a user can guess. Validate by joining to invoice_inputs and checking project_id before serving the file.
- `invoice_writeoff_handler.go:233 GetDocument` — same: validate the invoice exists in user's project before glob+serve.
- `invoice_materials.sql:ListInvoiceMaterials{With,Without}SerialNumbers` — add `AND invoice_materials.project_id = sqlc.arg(project_id)::bigint`. Then every `:id/materials/...` route (per flavor) passes `projectID`.
- Per-flavor `MaterialsForEdit` queries (`ListInvoiceOutputMaterialsForEdit`, `ListInvoiceInputMaterialsForEdit`, `ListInvoiceReturnMaterialsForEdit`, `ListInvoiceOutputOutOfProjectMaterialsForEdit`, `ListInvoiceWriteOffMaterialsForEdit`) — same shape, add project filter.
- Confirmation handlers (`<flavor>_handler.go Confirmation`) — after `usecase.GetByID(id)`, assert the loaded row's `ProjectID == c.GetUint("projectID")` and return an error if not. This blocks the "PDF upload to another project's invoice" path.
- Update handlers — same assertion before the UPDATE: load the row, check `existing.ProjectID == c.GetUint("projectID")`. Then issue the update. Blocks the re-parenting bug.
- Delete handlers — same: load, check project, then delete.

Since output and output-out-of-project share most patterns, you can factor a helper:

```go
func ensureProjectOwnership(ctx context.Context, q dbq.Queries, invoiceID uint, invoiceType string, projectID uint) error {
    // single SELECT … FROM invoice_<type> WHERE id = $1 — return row, then assert
}
```

and call it from every Update/Delete/Confirmation/GetDocument. ~40 lines total, replaces ~10 ad-hoc fixes.

### Shared root-cause clusters

Among the backend hardening list, the multi-handler clusters are:

1. The five "GetByDeliveryCode" queries (output, output-oop, return, input, writeoff) share the same shape and the same fix.
2. The five per-flavor "GetByID + accept upload" Confirmation handlers share the same fix.
3. The five per-flavor "MaterialsForEdit" queries share the same fix.

So the backend pass is "three identical recipes applied to five flavors each" = 15 mechanical changes plus the GetInvoiceObjectDescriptiveDataByID one-off.

---

## 7. Spot-check on non-invoice routes (per prompt instruction)

The prompt asked to verify whether the same project-scoping mechanism shows up correctly in three non-invoice endpoints.

| Endpoint | Handler | Project_id source | Verdict |
|----------|---------|-------------------|---------|
| `GET /material/paginated` (`/material/all`) | `material_handler.go:41 GetAll`, `:52 GetPaginated` | `c.GetUint("projectID")` flows to `ListMaterialsByProject` / `ListMaterialsPaginatedFiltered`, both with `WHERE project_id = $1` (`material.sql:5`, `:37`) | CORRECT |
| `GET /operation/paginated` | `operation_handler.go GetPaginated` | JWT → `ListOperationsPaginated` (`operation.sql:36`, `WHERE operations.project_id = $1`) | CORRECT |
| `GET /team/paginated`, `GET /team/all`, `GET /team/all/for-select` | `team_handler.go:42, :54, :235` | JWT → `ListTeamsByProject`, `ListTeamsPaginated`, etc. (`team.sql:1, :7`), all `WHERE teams.project_id = $1` | CORRECT |
| `GET /material-location/live` | `material_location_handler.go:262 Live` | JWT → usecase filter (`:280`) | CORRECT |
| `POST /main-reports/project-progress` | `main_report_handler.go ProjectProgress` | JWT → usecase | CORRECT (not re-verified line-by-line, but pattern matches) |

I did not find an endpoint outside the invoice surface that misuses projectID. The same patterns repeat — paginated lists are correctly scoped; by-id endpoints (`GET /team/:id`, `GET /worker/:id`, etc.) usually aren't, but they require knowing the id, which limits exposure to the same "guess an id from another project" attack surface as invoices. The fix shape is identical to the invoice case (add `AND project_id = $N`).

---

## 8. Open questions for you

1. **What does the user actually see — a momentary flash on first load, or persistent display?**
   - If "flash on first load that disappears within ~1 second" → cache hygiene (§6.i) is the fix.
   - If "persistent, the page never reconciles to the right project" → cache hygiene is necessary but not sufficient; need to dig into a specific report. Please grab:
     - The exact route the user is on when they see it.
     - A `curl -H "Authorization: Bearer <their-token>" https://.../api/<that-endpoint>` response, and the user_id and project_id you decode from the token at the same time. If the response includes rows whose `project_id` ≠ token's `project_id`, then there is a real backend bug or data-corruption issue we haven't located in source.

2. **Has anyone changed the JWT signing secret on the live deploy and forgotten to retire old tokens?** A JWT signed with the legacy secret containing a different `ProjectID` might still validate if both secrets exist somewhere. The migration runbook §3.3.1 makes a point of keeping `JWT_SECRET` the same, but worth double-checking against the deployed `.env.backend`.

3. **Any chance an invoice in production has `project_id` set wrong?** E.g. an Import flow that doesn't apply the projectID before insert. I spot-checked `invoice_input.Import` (it does set `data.Details.ProjectID = projectID` before calling repo Import). If you can identify a specific invoice that the user "shouldn't see," I can trace its likely insert path — but that's a server-data investigation, not a source one.

4. **Do superadmins ever appear to see "cross-project" invoices in user reports?** A `superadmin` role bypasses the user-in-project check at login but doesn't get any wildcard treatment at the query layer — they're still bound to their JWT's `ProjectID` and see only that one project. If the user reporting the bug is a superadmin who explicitly logs in to project "Администрирование", they may be viewing aggregated data that looks cross-project but is actually the admin project. Worth confirming the role of the complaining user.
