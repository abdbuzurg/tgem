# Permissions Specification

Design source of truth for the new permission/role architecture. Operations and
developers hand-edit this file; the data migration in phase 1 is generated from
it. Reviewers: confirm taxonomy + grants before phase 1 ships.

Companion to the architecture plan in chat / `PROGRESS.md`. Refresh this doc
whenever a new resource type or role is added.

---

## 1. Concepts

- **Resource type** — a stable, code-keyed domain concept (e.g. `invoice.output`).
  Decoupled from URLs, table names, and Russian display labels.
- **Action** — a verb performed on a resource type (e.g. `confirm`, `import`).
  Closed set; new verbs require a code change + migration.
- **Role** — a named bundle of `(resource_type, action)` grants. Roles do not
  encode project scope.
- **User-role assignment** — `(user, role, project_id)`. `project_id IS NULL`
  means the assignment applies regardless of project membership and is reserved
  for system / admin roles.

All three layers are normalized; the `resources(name, url)` table from the
legacy model is retired in phase 5.

---

## 2. Permission actions

The closed set of verbs. Lock these in before phase 1 — adding a new action
later is cheap (one row + one constant) but renaming one is not.

| `code`    | `display_ru`              | Notes |
|-----------|---------------------------|-------|
| `view`    | Просмотр                  | Read access. |
| `create`  | Создание                  | Insert new rows. |
| `edit`    | Изменение                 | Mutate existing rows (excludes `confirm`/`correct`). |
| `delete`  | Удаление                  | Hard or soft delete. |
| `confirm` | Подтверждение             | Invoice confirmation; moves stock. |
| `correct` | Корректировка             | Issue correction invoice (distinct from `edit` — creates a correction record). |
| `import`  | Импорт                    | Bulk Excel import. |
| `export`  | Экспорт                   | Generate Excel/PDF (covers print/download). |
| `report`  | Формирование отчёта       | Build / run a report. |

`approve`, `reject`, `assign`, `archive` deliberately omitted — see open items.

---

## 3. Resource types

Codes are dot-separated `category.subject[.subtype]`. Stable forever; never
recycled. `display_ru` is the user-visible label and may be re-edited freely.

### 3.1 `invoice` — invoice flavors

| `code`                              | `display_ru`                          | Backing aggregate |
|-------------------------------------|---------------------------------------|-------------------|
| `invoice.input`                     | Накладная приход                      | `invoice_inputs` |
| `invoice.output`                    | Накладная отпуск                      | `invoice_outputs` |
| `invoice.output_out_of_project`     | Накладная отпуск вне проекта          | `invoice_output_out_of_projects` |
| `invoice.return_team`               | Накладная возврат из бригад           | `invoice_returns` (returner_type='team') |
| `invoice.return_object`             | Накладная возврат из объекта          | `invoice_returns` (returner_type='object') |
| `invoice.writeoff`                  | Накладная списание                    | `invoice_write_offs` (covers all warehouse/team/object writeoff + loss UI variants — split later if granular control is needed) |
| `invoice.object`                    | Накладная объект                      | `invoice_objects` |
| `invoice.correction`                | Корректировка оператора               | `invoice_corrections` |

### 3.2 `reference` — reference books

| `code`                                   | `display_ru`                                    |
|------------------------------------------|-------------------------------------------------|
| `reference.material`                     | Справочник материалов                           |
| `reference.material_cost`                | Ценники материалов                              |
| `reference.material_defect`              | Бракованные материалы                           |
| `reference.material_location`            | Местоположение материала                        |
| `reference.serial_number`                | Серийные номера                                 |
| `reference.worker`                       | Сотрудники                                      |
| `reference.team`                         | Бригады                                         |
| `reference.district`                     | Районы                                          |
| `reference.operation`                    | Сервисы / операции                              |
| `reference.project`                      | Проекты                                         |
| `reference.object.kl04kv`                | Объект КЛ-04 кВ                                 |
| `reference.object.mjd`                   | Объект МЖД                                      |
| `reference.object.sip`                   | Объект СИП                                      |
| `reference.object.stvt`                  | Объект СТВТ                                     |
| `reference.object.tp`                    | Объект ТП                                       |
| `reference.object.substation`            | Объект подстанция                               |
| `reference.object.substation_cell`       | Объект ячейка подстанции                        |

The seven `reference.object.*` types intentionally mirror the application-level
polymorphism in `objects.type`. Permissions can be granted per-type so e.g. a
KL04KV-only specialist doesn't get edit on substations.

### 3.3 `report` — reports

| `code`                  | `display_ru`               |
|-------------------------|----------------------------|
| `report.balance`        | Отчёт остатков             |
| `report.invoice`        | Отчёт по накладным         |
| `report.attendance`     | Отчёт посещаемости         |
| `report.statistics`     | Статистика                 |

### 3.4 `admin` — administrative

| `code`                       | `display_ru`                                                |
|------------------------------|-------------------------------------------------------------|
| `admin.user`                 | Управление пользователями                                   |
| `admin.user_action`          | Журнал действий пользователей                               |
| `admin.user_in_project`      | Доступы пользователей в проекты                             |
| `admin.role`                 | Роли                                                        |
| `admin.role_grant`           | Назначения прав ролям                                       |
| `admin.resource_type`        | Управление типами ресурсов (внутр.)                         |
| `admin.project`              | Управление проектами                                        |

### 3.5 `auction` — auctions

| `code`                  | `display_ru`               | Notes |
|-------------------------|----------------------------|-------|
| `auction.bid_public`    | Публичный аукцион (ставки) | Project-independent. |
| `auction.bid_private`   | Закрытый аукцион (ставки)  | Project-independent. |
| `auction.manage`        | Управление аукционами      | Admin operations. |

### 3.6 `hr` — human resources

| `code`              | `display_ru`        |
|---------------------|---------------------|
| `hr.attendance`     | Посещаемость        |

### 3.7 `system` — cross-cutting tools

| `code`                            | `display_ru`                        |
|-----------------------------------|-------------------------------------|
| `system.import`                   | Массовый импорт                     |
| `system.material_location_live`   | Текущее местоположение материалов   |

---

## 4. Roles

| `code`                       | `display_ru`                  | Default scope     | Description |
|------------------------------|-------------------------------|-------------------|-------------|
| `superadmin`                 | Суперадмин                    | global (`NULL`)   | Full access. Single user (or very few). |
| `warehouse_keeper`           | Заведующий складом            | per-project       | Receives, issues, returns warehouse stock. |
| `pto`                        | ПТО                           | per-project       | Production-tech office: object data, corrections, reports. |
| `bidder`                     | Оферент                       | global            | Places auction bids; auctions are not project-scoped. |
| `supply_officer`             | Снабженец                     | per-project       | Procurement / incoming materials. |
| `supervisor`                 | Супервайзер                   | per-project       | Site supervision; objects + write-offs + reports. |
| `regional_project_manager`   | Региональный проект-менеджер  | per-project       | Manages project membership; broad read; selective edit. |
| `project_manager_assistant`  | Асистент проект-менеджера     | per-project       | Like RPM but reduced edit. |

Scope rules:
- **Global** (`project_id IS NULL`) — applies in every project. Use only for
  `superadmin` and (probably) `bidder`. Audit-log every assignment.
- **Per-project** — must enroll the user in `user_in_projects` for that project
  too; the role row alone is insufficient. Defense in depth.

---

## 5. Default grants per role

Each list is the **default starter set**. Operations can extend a role for a
specific tenant after rollout via the admin UI without touching this spec.

Notation: `actions = view, create, edit, delete, confirm, correct, import, export, report`. A role's grant per resource lists which subset applies.

### 5.1 `superadmin` (global)

All actions on every resource type. Implemented as a wildcard at the resolver
level (no enumerated rows) for efficiency and so newly-added resource types are
covered automatically. Document explicitly in `permission_usecase`.

### 5.2 `warehouse_keeper` (per-project)

| Resource type                          | Actions                                                   |
|----------------------------------------|-----------------------------------------------------------|
| `invoice.input`                        | `view, create, edit, delete, confirm, import, export`     |
| `invoice.output`                       | `view, create, edit, delete, confirm, import, export`     |
| `invoice.output_out_of_project`        | `view, create, edit, delete, confirm, import, export`     |
| `invoice.return_team`                  | `view, create, edit, delete, confirm, import, export`     |
| `invoice.return_object`                | `view, create, edit, delete, confirm, import, export`     |
| `invoice.writeoff`                     | `view, create, edit, delete, confirm, export`             |
| `invoice.correction`                   | `view`                                                    |
| `reference.material`                   | `view`                                                    |
| `reference.material_cost`              | `view`                                                    |
| `reference.material_location`          | `view`                                                    |
| `reference.material_defect`            | `view, create, edit`                                      |
| `reference.serial_number`              | `view, create, edit`                                      |
| `reference.worker`                     | `view`                                                    |
| `reference.team`                       | `view`                                                    |
| `reference.object.*` (all 7)           | `view`                                                    |
| `report.balance`                       | `view, report, export`                                    |
| `report.invoice`                       | `view, report, export`                                    |
| `system.material_location_live`        | `view`                                                    |
| `system.import`                        | `import`                                                  |

### 5.3 `pto` (per-project)

| Resource type                          | Actions                                                   |
|----------------------------------------|-----------------------------------------------------------|
| `invoice.input`                        | `view, export, report`                                    |
| `invoice.output`                       | `view, export, report`                                    |
| `invoice.output_out_of_project`        | `view, export, report`                                    |
| `invoice.return_team`                  | `view, export`                                            |
| `invoice.return_object`                | `view, export`                                            |
| `invoice.writeoff`                     | `view, export`                                            |
| `invoice.object`                       | `view, create, edit, confirm, export, report`             |
| `invoice.correction`                   | `view, create, edit, confirm, correct, export`            |
| `reference.material`                   | `view, create, edit`                                      |
| `reference.material_cost`              | `view, create, edit`                                      |
| `reference.material_location`          | `view`                                                    |
| `reference.material_defect`            | `view`                                                    |
| `reference.serial_number`              | `view`                                                    |
| `reference.team`                       | `view, create, edit`                                      |
| `reference.worker`                     | `view, create, edit`                                      |
| `reference.district`                   | `view, create, edit`                                      |
| `reference.operation`                  | `view, create, edit`                                      |
| `reference.project`                    | `view`                                                    |
| `reference.object.*` (all 7)           | `view, create, edit, delete`                              |
| `report.balance`                       | `view, report, export`                                    |
| `report.invoice`                       | `view, report, export`                                    |
| `report.statistics`                    | `view, report`                                            |
| `system.material_location_live`        | `view`                                                    |

### 5.4 `bidder` (global)

| Resource type             | Actions          |
|---------------------------|------------------|
| `auction.bid_public`      | `view, create`   |
| `auction.bid_private`     | `view, create`   |

No project assignments; lives outside the project domain.

### 5.5 `supply_officer` (per-project)

| Resource type                | Actions                              |
|------------------------------|--------------------------------------|
| `invoice.input`              | `view, create, edit, import, export` |
| `reference.material`         | `view, create, edit`                 |
| `reference.material_cost`    | `view, create, edit`                 |
| `reference.material_defect`  | `view`                               |
| `report.balance`             | `view, report, export`               |
| `system.import`              | `import`                             |

### 5.6 `supervisor` (per-project)

| Resource type                  | Actions                                |
|--------------------------------|----------------------------------------|
| `invoice.object`               | `view, create, edit, confirm, export`  |
| `invoice.return_object`        | `view, create, confirm, export`        |
| `invoice.writeoff`             | `view, create, export`                 |
| `invoice.correction`           | `view`                                 |
| `reference.team`               | `view`                                 |
| `reference.worker`             | `view`                                 |
| `reference.object.*` (all 7)   | `view, edit`                           |
| `report.balance`               | `view, report, export`                 |
| `report.invoice`               | `view, report, export`                 |

### 5.7 `regional_project_manager` (per-project)

| Resource type                       | Actions                          |
|-------------------------------------|----------------------------------|
| `invoice.*` (all flavors)           | `view, export, report`           |
| `reference.material`                | `view`                           |
| `reference.material_cost`           | `view`                           |
| `reference.material_location`       | `view`                           |
| `reference.material_defect`         | `view`                           |
| `reference.serial_number`           | `view`                           |
| `reference.operation`               | `view`                           |
| `reference.team`                    | `view, create, edit`             |
| `reference.worker`                  | `view, create, edit`             |
| `reference.district`                | `view, edit`                     |
| `reference.project`                 | `view, edit`                     |
| `reference.object.*` (all 7)        | `view, create, edit`             |
| `report.balance`                    | `view, report, export`           |
| `report.invoice`                    | `view, report, export`           |
| `report.statistics`                 | `view, report`                   |
| `report.attendance`                 | `view, report`                   |
| `admin.user_in_project`             | `view, edit`                     |
| `admin.user_action`                 | `view`                           |
| `hr.attendance`                     | `view, report`                   |
| `system.material_location_live`     | `view`                           |

### 5.8 `project_manager_assistant` (per-project)

| Resource type                  | Actions                |
|--------------------------------|------------------------|
| `invoice.*` (all flavors)      | `view, export, report` |
| `reference.material`           | `view`                 |
| `reference.material_cost`      | `view`                 |
| `reference.material_location`  | `view`                 |
| `reference.team`               | `view, edit`           |
| `reference.worker`             | `view, edit`           |
| `reference.object.*` (all 7)   | `view`                 |
| `reference.project`            | `view`                 |
| `report.balance`               | `view, report, export` |
| `report.invoice`               | `view, report, export` |
| `report.statistics`            | `view`                 |
| `hr.attendance`                | `view`                 |

---

## 6. Pages without a permission gate

These routes use `RequireAuth` only — they are containers, not gated content.
Items inside them are individually permission-checked.

| Route                | Reason                       |
|----------------------|------------------------------|
| `/home`              | Post-login landing.          |
| `/reference-book`    | Menu of reference books.     |
| `/report`            | Menu of reports.             |
| `/admin/home`        | Admin landing.               |
| `/permission-denied` | Denial page itself.          |
| `/404`               | Not found.                   |

Invariant: every account has at least one `user_roles` row at creation time.
Enforced in `user_usecase.Create` and the seed.

---

## 7. Backend route → required permission map

The middleware factory is `middleware.Require(action, resource_type)`. For every
backend route group, declare the required permission once at the group level
(or per-handler when actions differ). Example excerpt:

```go
inv := router.Group("/invoice-output", middleware.Authentication())
inv.GET("/",                middleware.Require("view",    "invoice.output"), h.GetAll)
inv.GET("/:id",             middleware.Require("view",    "invoice.output"), h.GetByID)
inv.POST("/",               middleware.Require("create",  "invoice.output"), h.Create)
inv.PATCH("/",              middleware.Require("edit",    "invoice.output"), h.Update)
inv.DELETE("/:id",          middleware.Require("delete",  "invoice.output"), h.Delete)
inv.POST("/confirm/:id",    middleware.Require("confirm", "invoice.output"), h.Confirm)
inv.POST("/document/import",middleware.Require("import",  "invoice.output"), h.Import)
inv.GET ("/document/export",middleware.Require("export",  "invoice.output"), h.Export)
```

A full route-to-permission table is generated mechanically during phase 2 and
checked into `internal/http/router_permissions.txt` (similar to the existing
golden `routes.txt` characterization).

---

## 8. Open items (resolve before phase 1)

These are still up for discussion with operations / stakeholders. Each one
blocks generating the migration script.

1. **Writeoff granularity.** Is `invoice.writeoff` a single resource, or do we
   split into `invoice.writeoff_warehouse`, `invoice.writeoff_team`,
   `invoice.writeoff_object` (mirroring the 5 frontend pages)? Default in this
   spec: single resource. Decision needed if any role should be able to do one
   variant but not another.

2. **Object-type granularity.** Seven `reference.object.*` codes vs one
   `reference.object`. Default: seven (mirrors `objects.type` polymorphism).
   Cost: 7× the grant rows per role. Benefit: per-type specialization. Confirm
   this is wanted before locking in.

3. **Audit-log read access.** `admin.user_action` currently granted only to
   `regional_project_manager`. Should `pto` or `supervisor` see it too?

4. **Auction roles beyond `bidder`.** Who manages auctions (`auction.manage`)?
   Today only `superadmin`. If your business has auction admins separate from
   superadmin, add an `auction_admin` role.

5. **Bidder scope.** Confirmed global, or should bids be tied to a project /
   organization? Today the auction routes `/auction/public` and
   `/auction/private` aren't project-scoped, so global is the default.

6. **Statistics access.** `report.statistics` is a heavy resource. Currently
   granted to `pto`, `regional_project_manager`, `project_manager_assistant`.
   Confirm — particularly whether `warehouse_keeper` and `supply_officer`
   should see project-level KPIs.

7. **HR module ownership.** `hr.attendance` is granted to RPM/PMA only. If
   site supervisors take attendance on the ground, add to `supervisor` too.

8. **Correction action.** Confirm `correct` is the right verb name (vs
   `recalc`, `adjust`). It's a domain-loaded term — defer to the operations
   team.

9. **Role display labels.** The spec uses the exact existing Russian labels
   from prod. Operations may want to standardize (e.g. add "(склад)",
   "(объект)" suffixes). Cosmetic — change `display_ru` freely without
   re-rolling out.

---

## 9. Status & deferred work (post phase 5)

The five-phase rollout described in chat is largely complete. What's still
intentionally deferred:

- **Drop legacy `permissions` and `resources` tables.** The `UserPermissionModal`
  and `AddNewRoleModal` admin pages still call `/permission/role/name/:role`
  and `/resource/`, which read from these tables. Dropping them now would
  break the admin tooling. Drop only after the new admin UI (next bullet)
  is built and switched over.

- **Rebuild the admin UI for (user, project, role) editing.** The new model
  needs a redesigned admin page that edits `user_roles` directly, supporting
  multiple roles per user across projects. Two recommended views:
  user-centric (`/admin/user/:id` showing a per-project roles table) and
  project-centric (`/admin/project/:id/roles` showing who has access).
  Don't build a full users × projects × roles matrix.

- **Drop `users.role_id` column.** Currently still consulted by the JWT
  payload (`user_usecase.go`) and the legacy admin tooling. Once the new
  admin UI lands and the JWT can carry the active project's role(s)
  instead, drop the column.

- **Cache the resolver.** `auth.NewResolver` queries
  `ListEffectivePermissionsForUser` per request. Add a per-(userID) TTL
  cache (~30 s) before this becomes a hot-path concern, with invalidation
  on `user_roles` / `role_grants` mutations.

- **Operations: log audit of grant changes.** `user_roles.granted_by` /
  `granted_at` capture the immediate context. Add a separate
  `permission_audit_log` table if you need full trail (revocations,
  before/after states).

- **Stakeholder confirmation on §8 open items.** None of them block the
  current rollout but each one is cheap to incorporate as soon as
  operations weighs in.

## 10. Change control

When you add a resource type, action, or role:

1. Edit this file.
2. Add a Goose migration that inserts the new row.
3. For new resource types: add the `middleware.Require(...)` calls to the
   relevant routes; the route-permissions golden updates automatically.
4. For new actions: add a constant in `internal/auth/actions.go` and update
   the resolver tests.
5. For new roles: insert a `roles` row + the `role_grants` rows in the same
   migration.

This file is the contract. Code that drifts from it should fail the
characterization tests.
