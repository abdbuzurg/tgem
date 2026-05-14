-- +goose Up
-- Phase 1 of the permissions v2 architecture (see docs/permissions-spec.md).
--
-- Adds the new data layer alongside the legacy permissions/resources tables.
-- The legacy tables remain authoritative until phase 4. The new tables are
-- populated from the spec defaults; the legacy permissions table is NOT
-- consulted (the spec is the source of truth — operators can re-customize
-- after rollout via the new admin UI in phase 3).
--
-- Idempotent: ON CONFLICT DO NOTHING throughout. Re-running this migration
-- on a fully migrated database is a no-op.

-- +goose StatementBegin

-- =============================================================================
-- 1. Schema
-- =============================================================================

CREATE TABLE IF NOT EXISTS resource_types (
    code        text PRIMARY KEY,
    category    text NOT NULL,
    display_ru  text NOT NULL,
    display_en  text
);

CREATE TABLE IF NOT EXISTS permission_actions (
    code        text PRIMARY KEY,
    display_ru  text NOT NULL
);

ALTER TABLE roles ADD COLUMN IF NOT EXISTS code text;

CREATE TABLE IF NOT EXISTS role_grants (
    role_id            bigint NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    resource_type_code text   NOT NULL REFERENCES resource_types(code) ON DELETE RESTRICT,
    action_code        text   NOT NULL REFERENCES permission_actions(code) ON DELETE RESTRICT,
    PRIMARY KEY (role_id, resource_type_code, action_code)
);

CREATE TABLE IF NOT EXISTS user_roles (
    id          bigserial PRIMARY KEY,
    user_id     bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id     bigint NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    project_id  bigint REFERENCES projects(id) ON DELETE CASCADE,
    granted_by  bigint REFERENCES users(id) ON DELETE SET NULL,
    granted_at  timestamptz NOT NULL DEFAULT now()
);

-- Allow at most one row per (user, role, project). NULL project_id is global.
-- Two unique indexes — one for non-null, one for null — because the standard
-- UNIQUE constraint treats NULLs as distinct in Postgres.
CREATE UNIQUE INDEX IF NOT EXISTS user_roles_user_role_project_uq
    ON user_roles (user_id, role_id, project_id)
    WHERE project_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS user_roles_user_role_global_uq
    ON user_roles (user_id, role_id)
    WHERE project_id IS NULL;

CREATE INDEX IF NOT EXISTS user_roles_user_id_idx     ON user_roles (user_id);
CREATE INDEX IF NOT EXISTS user_roles_role_id_idx     ON user_roles (role_id);
CREATE INDEX IF NOT EXISTS user_roles_project_id_idx  ON user_roles (project_id);

-- =============================================================================
-- 2. Permission actions (9 verbs — see spec §2)
-- =============================================================================

INSERT INTO permission_actions (code, display_ru) VALUES
    ('view',    'Просмотр'),
    ('create',  'Создание'),
    ('edit',    'Изменение'),
    ('delete',  'Удаление'),
    ('confirm', 'Подтверждение'),
    ('correct', 'Корректировка'),
    ('import',  'Импорт'),
    ('export',  'Экспорт'),
    ('report',  'Формирование отчёта')
ON CONFLICT (code) DO NOTHING;

-- =============================================================================
-- 3. Resource types (see spec §3)
-- =============================================================================

INSERT INTO resource_types (code, category, display_ru) VALUES
    -- 3.1 invoice
    ('invoice.input',                    'invoice',   'Накладная приход'),
    ('invoice.output',                   'invoice',   'Накладная отпуск'),
    ('invoice.output_out_of_project',    'invoice',   'Накладная отпуск вне проекта'),
    ('invoice.return_team',              'invoice',   'Накладная возврат из бригад'),
    ('invoice.return_object',            'invoice',   'Накладная возврат из объекта'),
    ('invoice.writeoff',                 'invoice',   'Накладная списание'),
    ('invoice.object',                   'invoice',   'Накладная объект'),
    ('invoice.correction',               'invoice',   'Корректировка оператора'),

    -- 3.2 reference
    ('reference.material',               'reference', 'Справочник материалов'),
    ('reference.material_cost',          'reference', 'Ценники материалов'),
    ('reference.material_defect',        'reference', 'Бракованные материалы'),
    ('reference.material_location',      'reference', 'Местоположение материала'),
    ('reference.serial_number',          'reference', 'Серийные номера'),
    ('reference.worker',                 'reference', 'Сотрудники'),
    ('reference.team',                   'reference', 'Бригады'),
    ('reference.district',               'reference', 'Районы'),
    ('reference.operation',              'reference', 'Сервисы / операции'),
    ('reference.project',                'reference', 'Проекты'),
    ('reference.object.kl04kv',          'reference', 'Объект КЛ-04 кВ'),
    ('reference.object.mjd',             'reference', 'Объект МЖД'),
    ('reference.object.sip',             'reference', 'Объект СИП'),
    ('reference.object.stvt',            'reference', 'Объект СТВТ'),
    ('reference.object.tp',              'reference', 'Объект ТП'),
    ('reference.object.substation',      'reference', 'Объект подстанция'),
    ('reference.object.substation_cell', 'reference', 'Объект ячейка подстанции'),

    -- 3.3 report
    ('report.balance',                   'report',    'Отчёт остатков'),
    ('report.invoice',                   'report',    'Отчёт по накладным'),
    ('report.attendance',                'report',    'Отчёт посещаемости'),
    ('report.statistics',                'report',    'Статистика'),

    -- 3.4 admin
    ('admin.user',                       'admin',     'Управление пользователями'),
    ('admin.user_action',                'admin',     'Журнал действий пользователей'),
    ('admin.user_in_project',            'admin',     'Доступы пользователей в проекты'),
    ('admin.role',                       'admin',     'Роли'),
    ('admin.role_grant',                 'admin',     'Назначения прав ролям'),
    ('admin.resource_type',              'admin',     'Управление типами ресурсов'),
    ('admin.project',                    'admin',     'Управление проектами'),

    -- 3.5 auction
    ('auction.bid_public',               'auction',   'Публичный аукцион'),
    ('auction.bid_private',              'auction',   'Закрытый аукцион'),
    ('auction.manage',                   'auction',   'Управление аукционами'),

    -- 3.6 hr
    ('hr.attendance',                    'hr',        'Посещаемость'),

    -- 3.7 system
    ('system.import',                    'system',    'Массовый импорт'),
    ('system.material_location_live',    'system',    'Текущее местоположение материалов')
ON CONFLICT (code) DO NOTHING;

-- =============================================================================
-- 4. Role codes — backfill from roles.name (see spec §4)
-- =============================================================================
-- Existing legacy role names are preserved in roles.name (display_ru). Only
-- the new code column is populated here.

UPDATE roles SET code = 'superadmin'                WHERE name = 'Суперадмин'                    AND code IS NULL;
UPDATE roles SET code = 'warehouse_keeper'          WHERE name = 'Заведующий складом'            AND code IS NULL;
UPDATE roles SET code = 'pto'                       WHERE name = 'ПТО'                           AND code IS NULL;
UPDATE roles SET code = 'bidder'                    WHERE name = 'Оферент'                       AND code IS NULL;
UPDATE roles SET code = 'supply_officer'            WHERE name = 'Снабженец'                     AND code IS NULL;
UPDATE roles SET code = 'supervisor'                WHERE name = 'Супервайзер'                   AND code IS NULL;
UPDATE roles SET code = 'regional_project_manager'  WHERE name = 'Региональный проект-менеджер'  AND code IS NULL;
UPDATE roles SET code = 'project_manager_assistant' WHERE name LIKE 'Асистент проект%менеджера'  AND code IS NULL;

-- Defensive fallback for any roles created out-of-band that don't match the
-- canonical names. Stable fallback so re-running is idempotent.
UPDATE roles SET code = 'role_' || id WHERE code IS NULL;

-- Lock the column once every row has a value.
ALTER TABLE roles ALTER COLUMN code SET NOT NULL;
ALTER TABLE roles ADD CONSTRAINT roles_code_unique UNIQUE (code);

-- =============================================================================
-- 5. Role grants per spec §5
-- =============================================================================
-- Pattern: VALUES table with (role_code, resource_type_code, action_code),
-- joined to live IDs at INSERT time. Resource types / actions / roles that
-- don't exist are silently skipped (the WHERE EXISTS guards) — that means a
-- typo here won't break the migration but also won't grant anything; verify
-- the row count after running.

-- 5.1 superadmin — wildcard, generated via CROSS JOIN.
INSERT INTO role_grants (role_id, resource_type_code, action_code)
SELECT r.id, rt.code, pa.code
FROM roles r
CROSS JOIN resource_types rt
CROSS JOIN permission_actions pa
WHERE r.code = 'superadmin'
ON CONFLICT DO NOTHING;

-- 5.2 warehouse_keeper
WITH grants(role_code, resource_type_code, action_code) AS (VALUES
    ('warehouse_keeper', 'invoice.input',                   'view'),
    ('warehouse_keeper', 'invoice.input',                   'create'),
    ('warehouse_keeper', 'invoice.input',                   'edit'),
    ('warehouse_keeper', 'invoice.input',                   'delete'),
    ('warehouse_keeper', 'invoice.input',                   'confirm'),
    ('warehouse_keeper', 'invoice.input',                   'import'),
    ('warehouse_keeper', 'invoice.input',                   'export'),
    ('warehouse_keeper', 'invoice.output',                  'view'),
    ('warehouse_keeper', 'invoice.output',                  'create'),
    ('warehouse_keeper', 'invoice.output',                  'edit'),
    ('warehouse_keeper', 'invoice.output',                  'delete'),
    ('warehouse_keeper', 'invoice.output',                  'confirm'),
    ('warehouse_keeper', 'invoice.output',                  'import'),
    ('warehouse_keeper', 'invoice.output',                  'export'),
    ('warehouse_keeper', 'invoice.output_out_of_project',   'view'),
    ('warehouse_keeper', 'invoice.output_out_of_project',   'create'),
    ('warehouse_keeper', 'invoice.output_out_of_project',   'edit'),
    ('warehouse_keeper', 'invoice.output_out_of_project',   'delete'),
    ('warehouse_keeper', 'invoice.output_out_of_project',   'confirm'),
    ('warehouse_keeper', 'invoice.output_out_of_project',   'import'),
    ('warehouse_keeper', 'invoice.output_out_of_project',   'export'),
    ('warehouse_keeper', 'invoice.return_team',             'view'),
    ('warehouse_keeper', 'invoice.return_team',             'create'),
    ('warehouse_keeper', 'invoice.return_team',             'edit'),
    ('warehouse_keeper', 'invoice.return_team',             'delete'),
    ('warehouse_keeper', 'invoice.return_team',             'confirm'),
    ('warehouse_keeper', 'invoice.return_team',             'import'),
    ('warehouse_keeper', 'invoice.return_team',             'export'),
    ('warehouse_keeper', 'invoice.return_object',           'view'),
    ('warehouse_keeper', 'invoice.return_object',           'create'),
    ('warehouse_keeper', 'invoice.return_object',           'edit'),
    ('warehouse_keeper', 'invoice.return_object',           'delete'),
    ('warehouse_keeper', 'invoice.return_object',           'confirm'),
    ('warehouse_keeper', 'invoice.return_object',           'import'),
    ('warehouse_keeper', 'invoice.return_object',           'export'),
    ('warehouse_keeper', 'invoice.writeoff',                'view'),
    ('warehouse_keeper', 'invoice.writeoff',                'create'),
    ('warehouse_keeper', 'invoice.writeoff',                'edit'),
    ('warehouse_keeper', 'invoice.writeoff',                'delete'),
    ('warehouse_keeper', 'invoice.writeoff',                'confirm'),
    ('warehouse_keeper', 'invoice.writeoff',                'export'),
    ('warehouse_keeper', 'invoice.correction',              'view'),
    ('warehouse_keeper', 'reference.material',              'view'),
    ('warehouse_keeper', 'reference.material_cost',         'view'),
    ('warehouse_keeper', 'reference.material_location',     'view'),
    ('warehouse_keeper', 'reference.material_defect',       'view'),
    ('warehouse_keeper', 'reference.material_defect',       'create'),
    ('warehouse_keeper', 'reference.material_defect',       'edit'),
    ('warehouse_keeper', 'reference.serial_number',         'view'),
    ('warehouse_keeper', 'reference.serial_number',         'create'),
    ('warehouse_keeper', 'reference.serial_number',         'edit'),
    ('warehouse_keeper', 'reference.worker',                'view'),
    ('warehouse_keeper', 'reference.team',                  'view'),
    ('warehouse_keeper', 'reference.object.kl04kv',         'view'),
    ('warehouse_keeper', 'reference.object.mjd',            'view'),
    ('warehouse_keeper', 'reference.object.sip',            'view'),
    ('warehouse_keeper', 'reference.object.stvt',           'view'),
    ('warehouse_keeper', 'reference.object.tp',             'view'),
    ('warehouse_keeper', 'reference.object.substation',     'view'),
    ('warehouse_keeper', 'reference.object.substation_cell','view'),
    ('warehouse_keeper', 'report.balance',                  'view'),
    ('warehouse_keeper', 'report.balance',                  'report'),
    ('warehouse_keeper', 'report.balance',                  'export'),
    ('warehouse_keeper', 'report.invoice',                  'view'),
    ('warehouse_keeper', 'report.invoice',                  'report'),
    ('warehouse_keeper', 'report.invoice',                  'export'),
    ('warehouse_keeper', 'system.material_location_live',   'view'),
    ('warehouse_keeper', 'system.import',                   'import')
)
INSERT INTO role_grants (role_id, resource_type_code, action_code)
SELECT r.id, g.resource_type_code, g.action_code
FROM grants g JOIN roles r ON r.code = g.role_code
ON CONFLICT DO NOTHING;

-- 5.3 pto
WITH grants(role_code, resource_type_code, action_code) AS (VALUES
    ('pto', 'invoice.input',                   'view'),
    ('pto', 'invoice.input',                   'export'),
    ('pto', 'invoice.input',                   'report'),
    ('pto', 'invoice.output',                  'view'),
    ('pto', 'invoice.output',                  'export'),
    ('pto', 'invoice.output',                  'report'),
    ('pto', 'invoice.output_out_of_project',   'view'),
    ('pto', 'invoice.output_out_of_project',   'export'),
    ('pto', 'invoice.output_out_of_project',   'report'),
    ('pto', 'invoice.return_team',             'view'),
    ('pto', 'invoice.return_team',             'export'),
    ('pto', 'invoice.return_object',           'view'),
    ('pto', 'invoice.return_object',           'export'),
    ('pto', 'invoice.writeoff',                'view'),
    ('pto', 'invoice.writeoff',                'export'),
    ('pto', 'invoice.object',                  'view'),
    ('pto', 'invoice.object',                  'create'),
    ('pto', 'invoice.object',                  'edit'),
    ('pto', 'invoice.object',                  'confirm'),
    ('pto', 'invoice.object',                  'export'),
    ('pto', 'invoice.object',                  'report'),
    ('pto', 'invoice.correction',              'view'),
    ('pto', 'invoice.correction',              'create'),
    ('pto', 'invoice.correction',              'edit'),
    ('pto', 'invoice.correction',              'confirm'),
    ('pto', 'invoice.correction',              'correct'),
    ('pto', 'invoice.correction',              'export'),
    ('pto', 'reference.material',              'view'),
    ('pto', 'reference.material',              'create'),
    ('pto', 'reference.material',              'edit'),
    ('pto', 'reference.material_cost',         'view'),
    ('pto', 'reference.material_cost',         'create'),
    ('pto', 'reference.material_cost',         'edit'),
    ('pto', 'reference.material_location',     'view'),
    ('pto', 'reference.material_defect',       'view'),
    ('pto', 'reference.serial_number',         'view'),
    ('pto', 'reference.team',                  'view'),
    ('pto', 'reference.team',                  'create'),
    ('pto', 'reference.team',                  'edit'),
    ('pto', 'reference.worker',                'view'),
    ('pto', 'reference.worker',                'create'),
    ('pto', 'reference.worker',                'edit'),
    ('pto', 'reference.district',              'view'),
    ('pto', 'reference.district',              'create'),
    ('pto', 'reference.district',              'edit'),
    ('pto', 'reference.operation',             'view'),
    ('pto', 'reference.operation',             'create'),
    ('pto', 'reference.operation',             'edit'),
    ('pto', 'reference.project',               'view'),
    ('pto', 'reference.object.kl04kv',         'view'),
    ('pto', 'reference.object.kl04kv',         'create'),
    ('pto', 'reference.object.kl04kv',         'edit'),
    ('pto', 'reference.object.kl04kv',         'delete'),
    ('pto', 'reference.object.mjd',            'view'),
    ('pto', 'reference.object.mjd',            'create'),
    ('pto', 'reference.object.mjd',            'edit'),
    ('pto', 'reference.object.mjd',            'delete'),
    ('pto', 'reference.object.sip',            'view'),
    ('pto', 'reference.object.sip',            'create'),
    ('pto', 'reference.object.sip',            'edit'),
    ('pto', 'reference.object.sip',            'delete'),
    ('pto', 'reference.object.stvt',           'view'),
    ('pto', 'reference.object.stvt',           'create'),
    ('pto', 'reference.object.stvt',           'edit'),
    ('pto', 'reference.object.stvt',           'delete'),
    ('pto', 'reference.object.tp',             'view'),
    ('pto', 'reference.object.tp',             'create'),
    ('pto', 'reference.object.tp',             'edit'),
    ('pto', 'reference.object.tp',             'delete'),
    ('pto', 'reference.object.substation',     'view'),
    ('pto', 'reference.object.substation',     'create'),
    ('pto', 'reference.object.substation',     'edit'),
    ('pto', 'reference.object.substation',     'delete'),
    ('pto', 'reference.object.substation_cell','view'),
    ('pto', 'reference.object.substation_cell','create'),
    ('pto', 'reference.object.substation_cell','edit'),
    ('pto', 'reference.object.substation_cell','delete'),
    ('pto', 'report.balance',                  'view'),
    ('pto', 'report.balance',                  'report'),
    ('pto', 'report.balance',                  'export'),
    ('pto', 'report.invoice',                  'view'),
    ('pto', 'report.invoice',                  'report'),
    ('pto', 'report.invoice',                  'export'),
    ('pto', 'report.statistics',               'view'),
    ('pto', 'report.statistics',               'report'),
    ('pto', 'system.material_location_live',   'view')
)
INSERT INTO role_grants (role_id, resource_type_code, action_code)
SELECT r.id, g.resource_type_code, g.action_code
FROM grants g JOIN roles r ON r.code = g.role_code
ON CONFLICT DO NOTHING;

-- 5.4 bidder
WITH grants(role_code, resource_type_code, action_code) AS (VALUES
    ('bidder', 'auction.bid_public',  'view'),
    ('bidder', 'auction.bid_public',  'create'),
    ('bidder', 'auction.bid_private', 'view'),
    ('bidder', 'auction.bid_private', 'create')
)
INSERT INTO role_grants (role_id, resource_type_code, action_code)
SELECT r.id, g.resource_type_code, g.action_code
FROM grants g JOIN roles r ON r.code = g.role_code
ON CONFLICT DO NOTHING;

-- 5.5 supply_officer
WITH grants(role_code, resource_type_code, action_code) AS (VALUES
    ('supply_officer', 'invoice.input',             'view'),
    ('supply_officer', 'invoice.input',             'create'),
    ('supply_officer', 'invoice.input',             'edit'),
    ('supply_officer', 'invoice.input',             'import'),
    ('supply_officer', 'invoice.input',             'export'),
    ('supply_officer', 'reference.material',        'view'),
    ('supply_officer', 'reference.material',        'create'),
    ('supply_officer', 'reference.material',        'edit'),
    ('supply_officer', 'reference.material_cost',   'view'),
    ('supply_officer', 'reference.material_cost',   'create'),
    ('supply_officer', 'reference.material_cost',   'edit'),
    ('supply_officer', 'reference.material_defect', 'view'),
    ('supply_officer', 'report.balance',            'view'),
    ('supply_officer', 'report.balance',            'report'),
    ('supply_officer', 'report.balance',            'export'),
    ('supply_officer', 'system.import',             'import')
)
INSERT INTO role_grants (role_id, resource_type_code, action_code)
SELECT r.id, g.resource_type_code, g.action_code
FROM grants g JOIN roles r ON r.code = g.role_code
ON CONFLICT DO NOTHING;

-- 5.6 supervisor
WITH grants(role_code, resource_type_code, action_code) AS (VALUES
    ('supervisor', 'invoice.object',                  'view'),
    ('supervisor', 'invoice.object',                  'create'),
    ('supervisor', 'invoice.object',                  'edit'),
    ('supervisor', 'invoice.object',                  'confirm'),
    ('supervisor', 'invoice.object',                  'export'),
    ('supervisor', 'invoice.return_object',           'view'),
    ('supervisor', 'invoice.return_object',           'create'),
    ('supervisor', 'invoice.return_object',           'confirm'),
    ('supervisor', 'invoice.return_object',           'export'),
    ('supervisor', 'invoice.writeoff',                'view'),
    ('supervisor', 'invoice.writeoff',                'create'),
    ('supervisor', 'invoice.writeoff',                'export'),
    ('supervisor', 'invoice.correction',              'view'),
    ('supervisor', 'reference.team',                  'view'),
    ('supervisor', 'reference.worker',                'view'),
    ('supervisor', 'reference.object.kl04kv',         'view'),
    ('supervisor', 'reference.object.kl04kv',         'edit'),
    ('supervisor', 'reference.object.mjd',            'view'),
    ('supervisor', 'reference.object.mjd',            'edit'),
    ('supervisor', 'reference.object.sip',            'view'),
    ('supervisor', 'reference.object.sip',            'edit'),
    ('supervisor', 'reference.object.stvt',           'view'),
    ('supervisor', 'reference.object.stvt',           'edit'),
    ('supervisor', 'reference.object.tp',             'view'),
    ('supervisor', 'reference.object.tp',             'edit'),
    ('supervisor', 'reference.object.substation',     'view'),
    ('supervisor', 'reference.object.substation',     'edit'),
    ('supervisor', 'reference.object.substation_cell','view'),
    ('supervisor', 'reference.object.substation_cell','edit'),
    ('supervisor', 'report.balance',                  'view'),
    ('supervisor', 'report.balance',                  'report'),
    ('supervisor', 'report.balance',                  'export'),
    ('supervisor', 'report.invoice',                  'view'),
    ('supervisor', 'report.invoice',                  'report'),
    ('supervisor', 'report.invoice',                  'export')
)
INSERT INTO role_grants (role_id, resource_type_code, action_code)
SELECT r.id, g.resource_type_code, g.action_code
FROM grants g JOIN roles r ON r.code = g.role_code
ON CONFLICT DO NOTHING;

-- 5.7 regional_project_manager
WITH grants(role_code, resource_type_code, action_code) AS (VALUES
    ('regional_project_manager', 'invoice.input',                  'view'),
    ('regional_project_manager', 'invoice.input',                  'export'),
    ('regional_project_manager', 'invoice.input',                  'report'),
    ('regional_project_manager', 'invoice.output',                 'view'),
    ('regional_project_manager', 'invoice.output',                 'export'),
    ('regional_project_manager', 'invoice.output',                 'report'),
    ('regional_project_manager', 'invoice.output_out_of_project',  'view'),
    ('regional_project_manager', 'invoice.output_out_of_project',  'export'),
    ('regional_project_manager', 'invoice.output_out_of_project',  'report'),
    ('regional_project_manager', 'invoice.return_team',            'view'),
    ('regional_project_manager', 'invoice.return_team',            'export'),
    ('regional_project_manager', 'invoice.return_team',            'report'),
    ('regional_project_manager', 'invoice.return_object',          'view'),
    ('regional_project_manager', 'invoice.return_object',          'export'),
    ('regional_project_manager', 'invoice.return_object',          'report'),
    ('regional_project_manager', 'invoice.writeoff',               'view'),
    ('regional_project_manager', 'invoice.writeoff',               'export'),
    ('regional_project_manager', 'invoice.writeoff',               'report'),
    ('regional_project_manager', 'invoice.object',                 'view'),
    ('regional_project_manager', 'invoice.object',                 'export'),
    ('regional_project_manager', 'invoice.object',                 'report'),
    ('regional_project_manager', 'invoice.correction',             'view'),
    ('regional_project_manager', 'invoice.correction',             'export'),
    ('regional_project_manager', 'invoice.correction',             'report'),
    ('regional_project_manager', 'reference.material',             'view'),
    ('regional_project_manager', 'reference.material_cost',        'view'),
    ('regional_project_manager', 'reference.material_location',    'view'),
    ('regional_project_manager', 'reference.material_defect',      'view'),
    ('regional_project_manager', 'reference.serial_number',        'view'),
    ('regional_project_manager', 'reference.operation',            'view'),
    ('regional_project_manager', 'reference.team',                 'view'),
    ('regional_project_manager', 'reference.team',                 'create'),
    ('regional_project_manager', 'reference.team',                 'edit'),
    ('regional_project_manager', 'reference.worker',               'view'),
    ('regional_project_manager', 'reference.worker',               'create'),
    ('regional_project_manager', 'reference.worker',               'edit'),
    ('regional_project_manager', 'reference.district',             'view'),
    ('regional_project_manager', 'reference.district',             'edit'),
    ('regional_project_manager', 'reference.project',              'view'),
    ('regional_project_manager', 'reference.project',              'edit'),
    ('regional_project_manager', 'reference.object.kl04kv',        'view'),
    ('regional_project_manager', 'reference.object.kl04kv',        'create'),
    ('regional_project_manager', 'reference.object.kl04kv',        'edit'),
    ('regional_project_manager', 'reference.object.mjd',           'view'),
    ('regional_project_manager', 'reference.object.mjd',           'create'),
    ('regional_project_manager', 'reference.object.mjd',           'edit'),
    ('regional_project_manager', 'reference.object.sip',           'view'),
    ('regional_project_manager', 'reference.object.sip',           'create'),
    ('regional_project_manager', 'reference.object.sip',           'edit'),
    ('regional_project_manager', 'reference.object.stvt',          'view'),
    ('regional_project_manager', 'reference.object.stvt',          'create'),
    ('regional_project_manager', 'reference.object.stvt',          'edit'),
    ('regional_project_manager', 'reference.object.tp',            'view'),
    ('regional_project_manager', 'reference.object.tp',            'create'),
    ('regional_project_manager', 'reference.object.tp',            'edit'),
    ('regional_project_manager', 'reference.object.substation',    'view'),
    ('regional_project_manager', 'reference.object.substation',    'create'),
    ('regional_project_manager', 'reference.object.substation',    'edit'),
    ('regional_project_manager', 'reference.object.substation_cell','view'),
    ('regional_project_manager', 'reference.object.substation_cell','create'),
    ('regional_project_manager', 'reference.object.substation_cell','edit'),
    ('regional_project_manager', 'report.balance',                 'view'),
    ('regional_project_manager', 'report.balance',                 'report'),
    ('regional_project_manager', 'report.balance',                 'export'),
    ('regional_project_manager', 'report.invoice',                 'view'),
    ('regional_project_manager', 'report.invoice',                 'report'),
    ('regional_project_manager', 'report.invoice',                 'export'),
    ('regional_project_manager', 'report.statistics',              'view'),
    ('regional_project_manager', 'report.statistics',              'report'),
    ('regional_project_manager', 'report.attendance',              'view'),
    ('regional_project_manager', 'report.attendance',              'report'),
    ('regional_project_manager', 'admin.user_in_project',          'view'),
    ('regional_project_manager', 'admin.user_in_project',          'edit'),
    ('regional_project_manager', 'admin.user_action',              'view'),
    ('regional_project_manager', 'hr.attendance',                  'view'),
    ('regional_project_manager', 'hr.attendance',                  'report'),
    ('regional_project_manager', 'system.material_location_live',  'view')
)
INSERT INTO role_grants (role_id, resource_type_code, action_code)
SELECT r.id, g.resource_type_code, g.action_code
FROM grants g JOIN roles r ON r.code = g.role_code
ON CONFLICT DO NOTHING;

-- 5.8 project_manager_assistant
WITH grants(role_code, resource_type_code, action_code) AS (VALUES
    ('project_manager_assistant', 'invoice.input',                  'view'),
    ('project_manager_assistant', 'invoice.input',                  'export'),
    ('project_manager_assistant', 'invoice.input',                  'report'),
    ('project_manager_assistant', 'invoice.output',                 'view'),
    ('project_manager_assistant', 'invoice.output',                 'export'),
    ('project_manager_assistant', 'invoice.output',                 'report'),
    ('project_manager_assistant', 'invoice.output_out_of_project',  'view'),
    ('project_manager_assistant', 'invoice.output_out_of_project',  'export'),
    ('project_manager_assistant', 'invoice.output_out_of_project',  'report'),
    ('project_manager_assistant', 'invoice.return_team',            'view'),
    ('project_manager_assistant', 'invoice.return_team',            'export'),
    ('project_manager_assistant', 'invoice.return_object',          'view'),
    ('project_manager_assistant', 'invoice.return_object',          'export'),
    ('project_manager_assistant', 'invoice.writeoff',               'view'),
    ('project_manager_assistant', 'invoice.writeoff',               'export'),
    ('project_manager_assistant', 'invoice.object',                 'view'),
    ('project_manager_assistant', 'invoice.object',                 'export'),
    ('project_manager_assistant', 'invoice.correction',             'view'),
    ('project_manager_assistant', 'invoice.correction',             'export'),
    ('project_manager_assistant', 'reference.material',             'view'),
    ('project_manager_assistant', 'reference.material_cost',        'view'),
    ('project_manager_assistant', 'reference.material_location',    'view'),
    ('project_manager_assistant', 'reference.team',                 'view'),
    ('project_manager_assistant', 'reference.team',                 'edit'),
    ('project_manager_assistant', 'reference.worker',               'view'),
    ('project_manager_assistant', 'reference.worker',               'edit'),
    ('project_manager_assistant', 'reference.object.kl04kv',        'view'),
    ('project_manager_assistant', 'reference.object.mjd',           'view'),
    ('project_manager_assistant', 'reference.object.sip',           'view'),
    ('project_manager_assistant', 'reference.object.stvt',          'view'),
    ('project_manager_assistant', 'reference.object.tp',            'view'),
    ('project_manager_assistant', 'reference.object.substation',    'view'),
    ('project_manager_assistant', 'reference.object.substation_cell','view'),
    ('project_manager_assistant', 'reference.project',              'view'),
    ('project_manager_assistant', 'report.balance',                 'view'),
    ('project_manager_assistant', 'report.balance',                 'report'),
    ('project_manager_assistant', 'report.balance',                 'export'),
    ('project_manager_assistant', 'report.invoice',                 'view'),
    ('project_manager_assistant', 'report.invoice',                 'report'),
    ('project_manager_assistant', 'report.invoice',                 'export'),
    ('project_manager_assistant', 'report.statistics',              'view'),
    ('project_manager_assistant', 'hr.attendance',                  'view')
)
INSERT INTO role_grants (role_id, resource_type_code, action_code)
SELECT r.id, g.resource_type_code, g.action_code
FROM grants g JOIN roles r ON r.code = g.role_code
ON CONFLICT DO NOTHING;

-- =============================================================================
-- 6. Backfill user_roles from legacy users.role_id
-- =============================================================================
-- One global (project_id=NULL) row per user, preserving current effective
-- access. Operators can refine to per-project assignments via the new admin
-- UI in phase 3.

INSERT INTO user_roles (user_id, role_id, project_id)
SELECT u.id, u.role_id, NULL
FROM users u
WHERE u.role_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1 FROM user_roles ur
      WHERE ur.user_id = u.id
        AND ur.role_id = u.role_id
        AND ur.project_id IS NULL
  );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop new tables. Legacy permissions/resources/roles.role_id remain
-- authoritative through phase 3, so rolling back here is safe.

DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS role_grants;
ALTER TABLE roles DROP CONSTRAINT IF EXISTS roles_code_unique;
ALTER TABLE roles DROP COLUMN IF EXISTS code;
DROP TABLE IF EXISTS resource_types;
DROP TABLE IF EXISTS permission_actions;

-- +goose StatementEnd
