# Phase 6 — Invoice XLSX investigation

## Pre-flight confirmation (2026-05-14)

Run from `/home/reborn20/projects/tgem/` on the dev workstation. Goal: validate every claim the §6 fix shape rests on, before touching `docker-compose.yml` or any handler.

### Step 1 — legacy backend cwd on the host **(UNRESOLVED — must be answered on production)**

This workstation is **not** the production host. Evidence collected here:

- `command -v pm2` → not installed. `~/.pm2` → does not exist. `ps auxf | grep tgem` → no Go process, only browser tabs.
- `docker ps` → no containers running. `docker volume ls | grep -E "tgem|backend"` → no volumes.
- `~/tgem` → does not exist (runbook §3.2 expected the legacy at `~/tgem/_legacy/tgem-backend/`).
- `find /home /var /opt /srv -type d -name excels` returned:
  - `/home/reborn20/projects/tgem/tgem-legacy/tgem-backend/pkg/excels` — fresh git checkout, `templates/` has 25 files, `{input,output,return,writeoff,temp}/` all empty (the runtime-generated files were never checked in to git).
  - `/home/reborn20/snap/code/238/.local/share/Trash/...` — a discarded copy, same shape, same emptiness.
  - **No path on this machine contains real legacy-generated xlsx/pdf files.**

The fix shape uses an env var (`TGEM_LEGACY_EXCELS_PATH`) so the compose file isn't bound to a specific host path. That means the diff below is **path-agnostic and can be reviewed now**, but before `docker compose up` is run on production the operator must:

```sh
# On the production host, with the legacy backend still under pm2:
pm2 show tgem-back | grep -E 'cwd|script path'   # primary
# fallback if pm2 was started detached:
ls -la /proc/$(pgrep -f tgem-back)/cwd 2>/dev/null
# verify the directory layout:
ls "<cwd>/pkg/excels/"
for d in input output return writeoff templates temp; do
    printf '%-10s %s\n' "$d" "$(ls "<cwd>/pkg/excels/$d" 2>/dev/null | wc -l)"
done
```

The first line that resolves to an absolute path containing `pkg/excels/{input,output,return,writeoff}` is the value to set in `TGEM_LEGACY_EXCELS_PATH` (as `<cwd>/pkg/excels`, with the trailing `pkg/excels` segment included). Until that value is known, the bind-mount cannot resolve and `docker compose up` will fail loudly (which is the desired behavior — `${TGEM_LEGACY_EXCELS_PATH:?must be set}` aborts compose if the var is missing rather than silently mounting nothing).

### Step 2 — every code path that reads from disk

Source: `/home/reborn20/projects/tgem/tgem-backend-rewrite/`. All paths are literal strings; none are constructed via a path helper or constant. None have fallback logic.

| Handler | Route | Computed path | What it serves | Fallback today? |
|---|---|---|---|---|
| `invoice_output_handler.go:226` (`GetDocument`) | `GET /output/document/:deliveryCode` | `./storage/import_excel/output/{deliveryCode}{.xlsx\|.pdf}` (extension from `invoiceOutputUsecase.GetDocument`) | Generated invoice file | No |
| `invoice_output_out_of_project_handler.go:284` (`GetDocument`) | `GET /invoice-output-out-of-project/document/:deliveryCode` | `./storage/import_excel/output/{deliveryCode}{.xlsx\|.pdf}` (same dir as output) | Generated invoice file | No |
| `invoice_return_handler.go:215` (`GetDocument`) | `GET /return/document/:deliveryCode` | `./storage/import_excel/return/{deliveryCode}{.xlsx\|.pdf}` | Generated invoice file | No |
| `invoice_input_handler.go:309` (`GetDocument`) | `GET /input/document/:deliveryCode` | `./storage/import_excel/input/{deliveryCode}.pdf` (only PDF — input never has xlsx) | Confirmation PDF | No |
| `invoice_writeoff_handler.go:237` (`GetDocument`) | `GET /invoice-writeoff/document/:deliveryCode` | `./storage/import_excel/writeoff/{deliveryCode}.*` via `filepath.Glob` | Confirmation PDF (or stale xlsx that lingered) | No. **Also has a panic-on-no-match bug at `:244` — `fileGlob[0]` indexed without checking len.** |
| `invoice_writeoff_handler.go:142` (`GetRawDocument`) | (not routed — dead interface method) | `./storage/import_excel/writeoff/{deliveryCode}.xlsx` | n/a | n/a |

Template reads (all go to `./internal/templates/...` — baked into the image via Dockerfile line 68):

- `worker_handler.go:173`, `team_handler.go:196`, `kl04kv_object_handler.go:154`, `stvt_object_handler.go:148`, `tp_object_handler.go:160`, `worker_attendance_handler.go:77`, `:82`, `substation_cell_object_handler.go:161`, `substation_object_handler.go:160`, `mjd_object_handler.go:155`, `operation_handler.go:226`, `material_handler.go:193`, `sip_object_handler.go:151` — import-template downloads.
- `invoice_output_usecase.go:755`, `:1157`; `invoice_output_out_of_project_usecase.go:588`, `:645`; `invoice_return_usecase.go:703`, `:1050`; `invoice_input_usecase.go:581`; `invoice_writeoff_usecase.go:556`; `invoice_correction_usecase.go:407`; plus all per-aggregate import flows — read templates for export/import. **All point at `./internal/templates/`, which the Dockerfile COPYs at build time. No fallback needed.**

Temp staging reads (all under `./storage/import_excel/temp/`) — used by Report and Import flows immediately after writing in the same request; `pkg/tempfiles.Track` deletes them via the request-cleanup middleware. **No fallback needed (legacy `pkg/excels/temp/` is documented as ephemeral too — nothing meaningful survives there).**

### Step 3 — every code path that writes to disk

| Handler / usecase | Path written | When |
|---|---|---|
| `invoice_output_usecase.go:1267` (`SaveAs`) | `./storage/import_excel/output/{DeliveryCode}.xlsx` | At `Create` and `Update` (`:228`, `:284`) |
| `invoice_output_out_of_project_usecase.go:722` (`SaveAs`) | `./storage/import_excel/output/{DeliveryCode}.xlsx` | At `Create` and `Update` (`:102`, `:337`) |
| `invoice_return_usecase.go:1190` (`SaveAs`) | `./storage/import_excel/return/{DeliveryCode}.xlsx` | At `Create` and `Update` (`:195`, `:299`) |
| `invoice_output_handler.go:200` (`SaveUploadedFile`) | `./storage/import_excel/output/{DeliveryCode}.pdf` | At `Confirmation` (replaces the xlsx via `os.Remove` at `:207`) |
| `invoice_output_out_of_project_handler.go:214` (`SaveUploadedFile`) | `./storage/import_excel/output/{DeliveryCode}.pdf` | At `Confirmation` (replaces the xlsx at `:221`) |
| `invoice_return_handler.go:190` (`SaveUploadedFile`) | `./storage/import_excel/return/{DeliveryCode}.pdf` | At `Confirmation` (replaces the xlsx at `:197`) |
| `invoice_input_handler.go:292` (`SaveUploadedFile`) | `./storage/import_excel/input/{DeliveryCode}.pdf` | At `Confirmation` (no xlsx to replace) |
| `invoice_writeoff_handler.go:218` (`SaveUploadedFile`) | `./storage/import_excel/writeoff/{DeliveryCode}.pdf` | At `Confirmation` (no xlsx to replace) |
| `invoice_output_handler.go:299`, `:406`; `invoice_output_out_of_project_handler.go:269`; `invoice_return_handler.go:266`; `invoice_input_handler.go:375`, `:438`; `invoice_writeoff_handler.go:266`; `invoice_correction_handler.go:225` | `./storage/import_excel/temp/{name}` | Import staging + Report exports (ephemeral) |

**Every write path is `./storage/import_excel/...`. Nothing writes to a `pkg/excels/`-shaped location. The fix never writes to `/app/legacy_excels/`** — keeping that mount `:ro` is correct.

### Step 4 — xlsx-at-create wiring per flavor

Walked `Create` in each usecase. Match against legacy in `MIGRATION/06-xlsx-investigation.md` §1.

| Flavor | Generates xlsx at Create on rewrite? | Generates xlsx at Create on legacy? | Verdict |
|---|---|---|---|
| `output` | **YES** — `Create` (`invoice_output_usecase.go:228`) calls `GenerateExcelFile` → `SaveAs` at `:1267` | YES | parity |
| `output-out-of-project` | **YES** — `Create` (`invoice_output_out_of_project_usecase.go:102`) calls `GenerateExcelFile` → `SaveAs` at `:722` | YES | parity |
| `return` | **YES** — `Create` (`invoice_return_usecase.go:195`) calls `GenerateExcel` → `SaveAs` at `:1190` | YES | parity |
| `input` | **NO** — `Create` (`invoice_input_usecase.go:232`) writes DB rows only; xlsx never generated at any phase. Confirmation uploads a PDF. | NO (legacy is also PDF-only via Confirmation) | parity, **not a bug** |
| `writeoff` | **NO** — `Create` (`invoice_writeoff_usecase.go:135`) writes DB rows only. Confirmation uploads a PDF. | NO (legacy is also PDF-only via Confirmation) | parity, **not a bug** |
| `correction` | NO at Create. Report-only via `:459` (temp). | NO | parity |
| `object` | NO at Create. Handler has no `GetDocument`, no `FileAttachment`, no `excelize` import. | NO | parity, no download endpoint exists |

**Conclusion for step 4:** no missing xlsx-at-create flows on the rewrite. The "input/writeoff don't generate xlsx" behavior is intentional and matches legacy. The Part 1 legacy-mount fix is sufficient on its own; no Part 3 follow-up needed.

### Step 5 — runtime / docker confirmation **(SOURCE-LEVEL ONLY — runtime checks must run on production)**

From `docker-compose.yml` and `tgem-backend-rewrite/Dockerfile`:

- Mount at `/app/storage`: `backend_storage` **named volume** (`docker-compose.yml:56`, declared at `:104`). No bind-mount to any host directory anywhere in the file.
- Container user: `app`, UID 10001 (Dockerfile `:58`, `USER app` at `:92`). `chown -R app:app /app` at `:90` covers the pre-seeded `/app/storage/import_excel/{input,output,return,writeoff,temp}` subdirs (created at `:82-87`) which Docker copies into the named volume on first mount.
- Volume persistence: named volumes survive `docker compose down` (and rebuilds, restarts). Only `docker compose down -v` deletes them. This is explicitly documented in the compose header (`docker-compose.yml:25-26`).

Runtime checks that must still happen on production after Part 1 lands (because none of these can be verified from the dev workstation):

```sh
docker compose exec backend stat -c '%U:%G %a' \
    /app/storage \
    /app/storage/import_excel \
    /app/storage/import_excel/{input,output,return,writeoff,temp} \
    /app/legacy_excels                              # post-Part 1
docker compose exec backend ls /app/legacy_excels/  # must show input/ output/ return/ writeoff/
docker compose exec backend touch /app/legacy_excels/output/.write_probe  # must fail (ro mount)
docker compose exec backend touch /app/storage/import_excel/output/.write_probe && \
    docker compose exec backend rm /app/storage/import_excel/output/.write_probe
```

If the existing `backend_storage` volume was created against an older image (pre-`chown -R app:app`), the named-volume root may still be `root:root` and writes fail with `permission denied` — see §6.A of the prior investigation for the one-shot `chown` fix.

### Summary — what this means for the fix

- Part 1 (mount) and Part 2 (fallback helper) are both safe to apply now. The diff is path-agnostic; the host path is supplied via env var at deploy time.
- Part 3 is **not needed**. Every invoice type that generated xlsx on legacy still generates xlsx on the rewrite; types that never generated xlsx still don't.
- The writeoff `GetDocument` glob has a latent panic on no-match (`fileGlob[0]` without a `len(fileGlob)>0` check). The Part 2 helper for the glob case fixes this incidentally — calling that out so it's not surprising in the diff.
- No template fallback is needed (`./internal/templates/` is image-baked).
- No temp-staging fallback is needed (ephemeral by design on both sides).

---

## Root cause in one sentence

The rewrite changed the on-disk path scheme from `./pkg/excels/{output,return,…}/` to `./storage/import_excel/{output,return,…}/` and the new container has no bind-mount to the host directory where the legacy backend wrote files, so legacy-created xlsx/pdf files are unreachable; new-invoice files **should** generate correctly into the `backend_storage` named volume but several edge cases (permissions on the named-volume mountpoint, the named-volume retaining stale ownership across rebuilds, and the rewrite's `output-out-of-project` reusing `./storage/import_excel/output/` together with `output`) deserve verification before declaring this fixed.

## Fix scope: small / medium / **medium**

- **(i) New invoices going forward:** small — a docker-compose volume tweak plus a defensive `MkdirAll` in the rewrite to make file generation tolerant of missing dirs after volume reset.
- **(ii) Old (legacy) invoices:** small in code, **operational** in effort — add a read-only bind-mount of the legacy `pkg/excels/` host directory into the container at `/app/legacy_excels/`, and teach the rewrite's `GetDocument` handlers to fall back to that directory when the primary path misses. Filenames are derived (not stored in DB), so no DB migration is needed.

---

## 1. Inventory — which legacy paths generate xlsx at create?

For each invoice flavor (`input`, `output`, `output-out-of-project`, `return`, `writeoff`, `correction`, `object`), the legacy code does the following:

| Flavor | Auto-generates xlsx at create? | Persistent path | Filename | Source |
|--------|--------------------------------|-----------------|----------|--------|
| **output** (warehouse → team) | **YES** | `./pkg/excels/output/` | `{DeliveryCode}.xlsx` | `internal/service/invoice_output_service.go:387` (Create), `:415` (Delete), `:963` (GenerateExcelFile SaveAs) |
| **output-out-of-project** | **YES** | `./pkg/excels/output/` (same dir as output) | `{DeliveryCode}.xlsx` | `internal/service/invoice_output_out_of_project_service.go:234`, `:511` |
| **return** | **YES** | `./pkg/excels/return/` | `{DeliveryCode}.xlsx` | `internal/service/invoice_return_service.go:264`, `:878` |
| **input** | NO (xlsx) — accepts a **PDF upload** at Confirmation | `./pkg/excels/input/` | `{DeliveryCode}.pdf` | `internal/controller/invoice_input_controller.go:289` (SaveUploadedFile) |
| **writeoff** | NO (xlsx) — accepts a **PDF upload** at Confirmation | `./pkg/excels/writeoff/` | `{DeliveryCode}.pdf` | `internal/controller/invoice_writeoff_controller.go:216` |
| **correction** | NO — only on-demand exports via `Report` to `./pkg/excels/temp/` | — | — | `internal/service/invoice_correction_service.go:262` (temp only) |
| **object** | NO — no xlsx-at-create at all | — | — | `internal/controller/invoice_object_controller.go`, `…/invoice_object_service.go` (no excelize calls) |

Filename is **derived** from `DeliveryCode` (a per-project counter like `О-12-25-0001`); nothing about the file path is stored in any DB column. There is no `files` table.

Confirmation upload (PDF) replaces the auto-generated xlsx for output/return/output-out-of-project: each service's `GetDocument` returns `.pdf` if `Confirmation = true`, else `.xlsx` (`invoice_output_service.go:977-981`, `invoice_return_service.go:891-896`, `invoice_output_out_of_project_service.go:548-552`). Download endpoint joins that extension with the same per-flavor directory.

Download endpoints (legacy):
- `/api/invoice-output/document/:deliveryCode` → `./pkg/excels/output/{deliveryCode}{.xlsx|.pdf}` (`invoice_output_controller.go:225`)
- `/api/invoice-output-out-of-project/document/:deliveryCode` → `./pkg/excels/output/{deliveryCode}{.xlsx|.pdf}` (`invoice_output_out_of_project_controller.go:282`)
- `/api/invoice-return/document/:deliveryCode` → `./pkg/excels/return/{deliveryCode}{.xlsx|.pdf}` (`invoice_return_controller.go:213`)
- `/api/invoice-input/document/:deliveryCode` → `./pkg/excels/input/{deliveryCode}.pdf` (`invoice_input_controller.go:308`)
- `/api/invoice-writeoff/document/:deliveryCode` → glob match in `./pkg/excels/writeoff/{deliveryCode}.*` (`invoice_writeoff_controller.go:237-244`)

The dead-code `GetRawDocument` (only writeoff has it, returning `./pkg/excels/writeoff/{code}.xlsx`) is declared on the interface in both legacy and rewrite but never routed — ignore.

---

## 2. Inventory — what the rewrite does

Same flavors, same code shape (the legacy code was copied largely verbatim), but **two structural differences**:

1. **Path moved.** `./pkg/excels/{output,return,writeoff,input,temp}/` → `./storage/import_excel/{output,return,writeoff,input,temp}/`.
2. **Templates moved.** `./pkg/excels/templates/` → `./internal/templates/`.

Generation-at-create (rewrite):

| Flavor | Generates xlsx at create? | Path | Source |
|--------|---------------------------|------|--------|
| output | YES | `./storage/import_excel/output/{DeliveryCode}.xlsx` | `internal/usecase/invoice_output_usecase.go:228` calls `GenerateExcelFile`, which writes via `f.SaveAs` at `:1266-1268` |
| output-out-of-project | YES | `./storage/import_excel/output/{DeliveryCode}.xlsx` | `internal/usecase/invoice_output_out_of_project_usecase.go:102` calls `GenerateExcelFile`, writes at `:721` |
| return | YES | `./storage/import_excel/return/{DeliveryCode}.xlsx` | `internal/usecase/invoice_return_usecase.go:195` calls `GenerateExcel`, writes at `:1189-1190` |
| input | NO (xlsx) — accepts PDF at Confirmation | `./storage/import_excel/input/` | `internal/http/handlers/invoice_input_handler.go:290` |
| writeoff | NO (xlsx) — accepts PDF at Confirmation | `./storage/import_excel/writeoff/` | `internal/http/handlers/invoice_writeoff_handler.go:216` |
| correction | NO | (temp only) | `internal/usecase/invoice_correction_usecase.go:458` |
| object | NO | — | no excelize calls |

Download endpoints (rewrite) — paths match the create paths above:
- output: `./storage/import_excel/output/{deliveryCode}{.xlsx|.pdf}` (`invoice_output_handler.go:226-227`)
- output-out-of-project: `./storage/import_excel/output/…` (`invoice_output_out_of_project_handler.go:284-285`)
- return: `./storage/import_excel/return/…` (`invoice_return_handler.go:215-216`)
- input: `./storage/import_excel/input/{deliveryCode}.pdf` (`invoice_input_handler.go:309-314`)
- writeoff: glob `./storage/import_excel/writeoff/{deliveryCode}.*` (`invoice_writeoff_handler.go:237-248`)

So the rewrite **does not understand** any legacy path scheme. There is no fallback to `./pkg/excels/…` anywhere in the rewrite.

---

## 3. Deployment shape

`docker-compose.yml`:

```yaml
backend:
  volumes:
    - backend_storage:/app/storage
    - backend_files:/app/files
```

`backend_storage` is a Docker **named volume**. Its initial contents are seeded from the image's `/app/storage/...` directory (created by the Dockerfile via `mkdir -p /app/storage/import_excel/{input,output,return,writeoff,temp} && chown -R app:app /app`). The container process runs as user `app` (UID 10001). New files written into `/app/storage/import_excel/output/...` end up inside the `backend_storage` volume on the host (typically `/var/lib/docker/volumes/tgem_backend_storage/_data/`), and survive container restarts and rebuilds (until `docker compose down -v`).

`backend_files` covers `Files.Path` (`./files`) — referenced by config, used by no invoice flavor.

**There is no bind-mount to any host directory.** In particular:
- The legacy `_legacy/tgem-backend/pkg/excels/output/`, `…/pkg/excels/return/`, `…/pkg/excels/input/`, `…/pkg/excels/writeoff/` on the production host are **not visible inside the backend container** at any path.

The legacy pm2 process's cwd determined where `./pkg/excels/...` resolved. From `MIGRATION/03-runbook.md` the legacy lives at `${TGEM_REPO}/_legacy/tgem-backend/` (= `~/tgem/_legacy/tgem-backend/` on the production host); under pm2 the cwd is typically the directory passed to `pm2 start`. Assuming standard layout, legacy files are at:

```
~/tgem/_legacy/tgem-backend/pkg/excels/output/{DeliveryCode}.xlsx
~/tgem/_legacy/tgem-backend/pkg/excels/output/{DeliveryCode}.pdf   # (output, output-oop after Confirmation)
~/tgem/_legacy/tgem-backend/pkg/excels/return/{DeliveryCode}.xlsx
~/tgem/_legacy/tgem-backend/pkg/excels/return/{DeliveryCode}.pdf
~/tgem/_legacy/tgem-backend/pkg/excels/input/{DeliveryCode}.pdf
~/tgem/_legacy/tgem-backend/pkg/excels/writeoff/{DeliveryCode}.pdf
```

**Question for you (cannot answer from source):**

> Q1. What was the absolute pm2 cwd for the legacy backend on the production host? Run `pm2 describe ${TGEM_PM2_PROCESS} | grep -E "cwd|script path|exec cwd"`, or `ls -la /proc/$(pgrep -f tgem-back)/cwd` while it was running. The contents of `pkg/excels/output/`, `pkg/excels/return/`, etc. live in `<that-cwd>/pkg/excels/{output,return,input,writeoff}/`.

If by mistake the legacy was started from a different directory (e.g. somebody ran `pm2 start ./main` from `~/`), then the files are at `~/pkg/excels/...` instead. The runbook §3.2 backup did **not** capture this directory, so the on-disk state of the host is the only authoritative source.

---

## 4. Sub-issues (a)–(d): verdicts

### (a) Rewrite doesn't generate xlsx at create

**Verdict: NOT PRESENT in source.** For `output`, `output-out-of-project`, and `return`, the rewrite calls the same generation function the legacy did, on the same control-flow path, before the create transaction commits.

Caveats that mean we should still verify on the server:

- `excelize.SaveAs` does not auto-create parent directories. The Dockerfile pre-creates `/app/storage/import_excel/{output,return,writeoff,input,temp}` before the named volume is mounted, and Docker copies that pre-created tree into the empty named volume on first mount. **But** if the named volume was created *before* the Dockerfile change that added a particular subdir (e.g. `temp/` was added later), or if someone manually `docker volume rm`d and now re-creates without rebuilding, a directory could be missing and `SaveAs` would fail at the first new invoice. **Easy verification:** `docker compose exec backend ls -la /app/storage/import_excel/`. If any of `output/`, `return/`, `writeoff/`, `input/`, `temp/` is missing or not writable by `app` (UID 10001), that's the root cause.
- The named volume's root is owned by whatever the image's `/app/storage` was at first mount. With `chown -R app:app /app` in the Dockerfile this should be `app:app`. If a stale volume from a pre-chown build is still in place, the named-volume root is `root:root` and writes fail with `permission denied`. **Verification:** `docker compose exec backend stat -c '%U:%G %a' /app/storage /app/storage/import_excel /app/storage/import_excel/output`.
- The Create flow calls `GenerateExcelFile(data)` *before* the DB transaction. If generation fails, the entire Create returns an error — the frontend would not see a row at all, so symptom "row exists but file missing" rules this out for genuinely new rows.

### (b) Rewrite generates to an ephemeral path that vanishes on rebuild

**Verdict: NOT PRESENT in source.** `/app/storage` is backed by the `backend_storage` named volume, which persists across container restarts, image rebuilds, and `docker compose down` (but not `docker compose down -v`). `docker-compose.yml:55` confirms.

### (c) Create path ≠ Download path

**Verdict: NOT PRESENT in source.** For every flavor that generates, create writes to and download reads from the same directory and the same filename. Cross-checked in §2 above. No regression here.

### (d) Legacy files on host unreachable from container

**Verdict: CONFIRMED.** There is no bind-mount of the legacy `pkg/excels/` host directory into the container. The rewrite's download handlers compute paths that exist only inside the named volume (which was empty at cutover). Any invoice whose document was written by the legacy backend before cutover is now unreachable. This explains "older invoices created under the legacy backend" being unavailable.

Note this affects **all** flavors that ever served a document:
- output / output-oop / return: the auto-generated xlsx, AND any PDF uploaded via Confirmation under legacy.
- input / writeoff: the confirmation PDFs uploaded under legacy.

---

## 5. Where the legacy files actually live (host paths to mount)

Pending Q1 above, the canonical guess based on the runbook's layout (`${TGEM_REPO}=~/tgem`, `_legacy/tgem-backend/`):

```
~/tgem/_legacy/tgem-backend/pkg/excels/output/         # output + output-out-of-project files
~/tgem/_legacy/tgem-backend/pkg/excels/return/         # return files
~/tgem/_legacy/tgem-backend/pkg/excels/input/          # input confirmation PDFs
~/tgem/_legacy/tgem-backend/pkg/excels/writeoff/       # writeoff confirmation PDFs
```

Don't bind-mount `pkg/excels/temp/` or `pkg/excels/templates/` — the first is ephemeral export staging, the second is read-only template assets baked into the rewrite image.

---

## 6. Proposed fix

### (i) Make new invoices work going forward

A. **Verify directory + permissions inside the container** (zero code change first):

```sh
docker compose exec backend ls -la /app/storage/import_excel/
docker compose exec backend stat -c '%U:%G %a' \
    /app/storage \
    /app/storage/import_excel \
    /app/storage/import_excel/output \
    /app/storage/import_excel/return \
    /app/storage/import_excel/writeoff \
    /app/storage/import_excel/input \
    /app/storage/import_excel/temp
```

Expected ownership `app:app` with mode `0755` (or `0775`). If any are missing or owned by `root`, the named volume was created against an older image. Fix without losing data:

```sh
docker compose run --rm --user 0:0 backend sh -c '\
    mkdir -p /app/storage/import_excel/{output,return,writeoff,input,temp} && \
    chown -R 10001:10001 /app/storage'
```

(Runs one-shot as root inside the container to fix the volume contents, then exits.)

B. **Defensive `MkdirAll` in the rewrite.** Make every place that calls `f.SaveAs` ensure the parent directory exists. Add a one-line `os.MkdirAll(filepath.Dir(path), 0o755)` before each of:

- `internal/usecase/invoice_output_usecase.go:1266` — output generate
- `internal/usecase/invoice_output_out_of_project_usecase.go:721` — out-of-project generate
- `internal/usecase/invoice_return_usecase.go:1189` — return generate

This way the rewrite recovers automatically from a freshly-recreated named volume that wasn't pre-populated.

C. **No docker-compose change needed for forward generation.** `backend_storage:/app/storage` is already correct.

### (ii) Make legacy files accessible

D. **Mount the legacy host directory read-only into the container.** Edit `docker-compose.yml`:

```yaml
backend:
  volumes:
    - backend_storage:/app/storage
    - backend_files:/app/files
    - /home/<user>/tgem/_legacy/tgem-backend/pkg/excels:/app/legacy_excels:ro     # NEW
```

Use the absolute host path; bind mounts can't be relative. The runbook expects the legacy at `~/tgem/_legacy/tgem-backend/`, so `${HOME}/tgem/_legacy/tgem-backend/pkg/excels` is the most likely value — confirm via Q1 first.

`:ro` ensures the container can never mutate the legacy files (so a buggy Delete on a legacy invoice can't remove the legacy artifact).

E. **Teach the rewrite's `GetDocument` handlers to fall back.** Smallest-blast-radius change: when `os.Stat(primary)` fails with `ErrNotExist`, retry against the legacy mount before returning an error.

Concretely, add a small helper in `internal/http/handlers/`:

```go
// resolveInvoiceDoc returns the first existing path among the candidates,
// or "" if none exist. Keeps download handlers branch-light.
func resolveInvoiceDoc(deliveryCode, ext, primaryDir, legacyDir string) string {
    primary := filepath.Join(primaryDir, deliveryCode+ext)
    if _, err := os.Stat(primary); err == nil {
        return primary
    }
    legacy := filepath.Join(legacyDir, deliveryCode+ext)
    if _, err := os.Stat(legacy); err == nil {
        return legacy
    }
    return ""
}
```

Then in each handler's `GetDocument`, replace `filePath := filepath.Join(primaryDir, ...)` with a call to the helper. Affected handlers:

- `internal/http/handlers/invoice_output_handler.go:218` — fallback dir `/app/legacy_excels/output`
- `internal/http/handlers/invoice_output_out_of_project_handler.go:273` — fallback dir `/app/legacy_excels/output`
- `internal/http/handlers/invoice_return_handler.go:208` — fallback dir `/app/legacy_excels/return`
- `internal/http/handlers/invoice_input_handler.go:307` — fallback dir `/app/legacy_excels/input`
- `internal/http/handlers/invoice_writeoff_handler.go:233` — uses `filepath.Glob`; extend the glob to search both directories or fall through.

For the writeoff glob, the simplest pattern is:

```go
for _, dir := range []string{primaryDir, legacyDir} {
    if matches, _ := filepath.Glob(filepath.Join(dir, deliveryCode+".*")); len(matches) > 0 {
        // serve matches[0]
    }
}
```

F. **Make the legacy mount point optional.** If the host path is wrong or the legacy directory is later archived, the bind mount fails the whole compose. To avoid that surprise, you can put the bind-mount in an override file (`docker-compose.override.yml`) so it's only active when explicitly required:

```yaml
# docker-compose.override.yml
services:
  backend:
    volumes:
      - /home/<user>/tgem/_legacy/tgem-backend/pkg/excels:/app/legacy_excels:ro
```

Or use the long-form mount with `bind.create_host_path: false` so compose errors clearly when the host dir is missing instead of silently creating an empty one.

---

## 7. Migration concerns

- **Are filenames in the DB?** No. Both legacy and rewrite derive the path from `DeliveryCode` (a string column on the invoice row) plus a hardcoded directory. No DB migration is needed.
- **Do legacy filenames contain paths that need rewriting?** No — `DeliveryCode` is just the counter string like `О-12-25-0001`. Basenames work regardless of which directory the file lives in.
- **Output and output-out-of-project share `./pkg/excels/output/` in legacy and `./storage/import_excel/output/` in rewrite.** This is intentional (and matches legacy). If two invoices of these different types happened to receive the same `DeliveryCode`, one would overwrite the other. The `DeliveryCode` counters are per-`InvoiceType` (see `invoice_counts` table), so collisions only happen if `output` and `output-out-of-project` ever produce identical strings — which they currently can, since `UniqueCodeGeneration` uses a fixed prefix `"О"` for output and migration `00004_split_output_out_of_project_counter.sql` was added precisely to give out-of-project its own counter. **Check whether output-out-of-project's `UniqueCodeGeneration` prefix is the same as output's** before fix (ii) is shipped; if so, the legacy fallback can return a wrong file. Verification: grep `UniqueCodeGeneration\(` in both usecases and check the first argument.
- **PDFs uploaded under legacy and the equivalent rewrite Confirmation upload coexist?** Yes. They share the same `{DeliveryCode}.pdf` filename, but they live in different directories — legacy in `/app/legacy_excels/<flavor>/`, rewrite in `/app/storage/import_excel/<flavor>/`. The rewrite's primary path wins; the legacy fallback only fires when the rewrite path misses. Once a legacy invoice is re-Confirmed in the rewrite (replacing the PDF), the rewrite path takes over.

---

## 8. Open questions for you

1. **Q1 (above): legacy backend's pm2 cwd.** Determines the host path to bind-mount in fix (ii).
2. **Q2: is "new invoices don't work" actually true?** Source review says generation at create is wired correctly. Before changing code, please test: log in to the rewrite, create a new output invoice, and click download. If the download works, the bug is purely (d) (legacy files only). If it doesn't, run §6.A to inspect volume permissions and report what `stat` shows.
3. **Q3: do we want the legacy mount as a permanent feature, or only until you've decided to migrate the files into the named volume?** A one-time copy `cp -a ~/tgem/_legacy/tgem-backend/pkg/excels/{output,return,input,writeoff}/* …/_data/import_excel/<corresponding-dir>/` would eliminate the fallback entirely. Trade-off: simpler runtime code, but harder rollback (the legacy directory becomes dispensable).
