-- name: MaterialDataForProgressReportInProject :many
SELECT
    materials.id                                                       AS id,
    COALESCE(materials.code, '')::text                                 AS code,
    COALESCE(materials.name, '')::text                                 AS name,
    COALESCE(materials.unit, '')::text                                 AS unit,
    COALESCE(materials.planned_amount_for_project, 0)::float8          AS planned_amount_for_project,
    COALESCE(material_locations.amount, 0)::float8                     AS location_amount,
    COALESCE(material_locations.location_type, '')::text               AS location_type
FROM material_locations
INNER JOIN material_costs ON material_locations.material_cost_id = material_costs.id
INNER JOIN materials ON material_costs.material_id = materials.id
WHERE
    materials.project_id = $1
    AND materials.show_planned_amount_in_report = true
ORDER BY materials.id;

-- name: InvoiceMaterialDataForProgressReport :many
SELECT
    materials.id                            AS material_id,
    COALESCE(invoice_materials.amount, 0)::float8 AS amount,
    COALESCE(invoice_materials.invoice_type, '')::text AS invoice_type,
    material_costs.cost_with_customer       AS cost_with_customer
FROM materials
INNER JOIN material_costs ON material_costs.material_id = materials.id
RIGHT JOIN invoice_materials ON invoice_materials.material_cost_id = material_costs.id
WHERE
    materials.project_id = $1
    AND materials.show_planned_amount_in_report = true
    AND (invoice_materials.invoice_type = 'input' OR invoice_materials.invoice_type = 'object-correction')
ORDER BY materials.id;

-- name: InvoiceOperationDataForProgressReport :many
SELECT
    operations.id                                                AS id,
    COALESCE(operations.code, '')::text                          AS code,
    COALESCE(operations.name, '')::text                          AS name,
    operations.cost_with_customer                                AS cost_with_customer,
    COALESCE(operations.planned_amount_for_project, 0)::float8   AS planned_amount_for_project,
    COALESCE(invoice_operations.amount, 0)::float8               AS amount_in_invoice
FROM invoice_objects
INNER JOIN invoice_operations ON invoice_operations.invoice_id = invoice_objects.id
INNER JOIN operations ON operations.id = invoice_operations.operation_id
WHERE
    invoice_objects.confirmed_by_operator = true
    AND invoice_operations.invoice_type = 'object-correction'
    AND operations.project_id = $1
    AND operations.show_planned_amount_in_report = true
ORDER BY operations.id;

-- name: MaterialDataForProgressReportInProjectInGivenDate :many
SELECT
    materials.id                                                                          AS id,
    COALESCE(materials.code, '')::text                                                    AS code,
    COALESCE(materials.name, '')::text                                                    AS name,
    COALESCE(materials.unit, '')::text                                                    AS unit,
    COALESCE(materials.planned_amount_for_project, 0)::float8                             AS amount_planned_for_project,
    COALESCE(project_progress_materials.received, 0)::float8                              AS amount_received,
    COALESCE(project_progress_materials.installed, 0)::float8                             AS amount_installed,
    COALESCE(project_progress_materials.amount_in_warehouse, 0)::float8                   AS amount_in_warehouse,
    COALESCE(project_progress_materials.amount_in_teams, 0)::float8                       AS amount_in_teams,
    COALESCE(project_progress_materials.amount_in_objects, 0)::float8                     AS amount_in_objects,
    COALESCE(project_progress_materials.amount_write_off, 0)::float8                      AS amount_write_off,
    material_costs.cost_with_customer                                                     AS cost_with_customer
FROM project_progress_materials
INNER JOIN material_costs ON material_costs.id = project_progress_materials.material_cost_id
INNER JOIN materials ON materials.id = material_costs.material_id
WHERE
    project_progress_materials.project_id = $1
    AND $2::timestamptz < project_progress_materials.date
    AND project_progress_materials.date < $3::timestamptz
ORDER BY materials.id;

-- name: InvoiceOperationDataForProgressReportInGivenDate :many
SELECT
    COALESCE(operations.code, '')::text                          AS code,
    COALESCE(operations.name, '')::text                          AS name,
    operations.cost_with_customer                                AS cost_with_customer,
    COALESCE(operations.planned_amount_for_project, 0)::float8   AS amount_planned_for_project,
    COALESCE(project_progress_operations.installed, 0)::float8   AS amount_installed
FROM project_progress_operations
INNER JOIN operations ON project_progress_operations.operation_id = operations.id
WHERE
    project_progress_operations.project_id = $1
    AND $2::timestamptz < project_progress_operations.date
    AND project_progress_operations.date < $3::timestamptz;

-- name: MaterialDataForRemainingMaterialAnalysis :many
SELECT
    materials.id                                                       AS id,
    COALESCE(materials.code, '')::text                                 AS code,
    COALESCE(materials.name, '')::text                                 AS name,
    COALESCE(materials.unit, '')::text                                 AS unit,
    COALESCE(materials.planned_amount_for_project, 0)::float8          AS planned_amount_for_project,
    COALESCE(material_locations.amount, 0)::float8                     AS location_amount,
    COALESCE(material_locations.location_type, '')::text               AS location_type
FROM material_locations
INNER JOIN material_costs ON material_locations.material_cost_id = material_costs.id
INNER JOIN materials ON material_costs.material_id = materials.id
WHERE
    materials.project_id = $1
    AND materials.show_planned_amount_in_report = true
    AND (material_locations.location_type = 'warehouse' OR material_locations.location_type = 'team')
ORDER BY materials.id;

-- name: MaterialsInstalledOnObjectForRemainingMaterialAnalysis :many
SELECT
    materials.id                                                AS id,
    COALESCE(invoice_materials.amount, 0)::float8               AS amount,
    invoice_objects.date_of_correction                          AS date_of_correction
FROM invoice_objects
INNER JOIN invoice_materials ON invoice_objects.id = invoice_materials.invoice_id
INNER JOIN material_costs ON invoice_materials.material_cost_id = material_costs.id
INNER JOIN materials ON material_costs.material_id = materials.id
WHERE
    materials.project_id = $1
    AND materials.show_planned_amount_in_report = true
    AND invoice_objects.confirmed_by_operator = true
    AND invoice_materials.invoice_type = 'object-correction'
ORDER BY materials.id;
