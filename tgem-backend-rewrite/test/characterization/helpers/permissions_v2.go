package helpers

import (
	"context"
	"fmt"
)

// seedPermissionsV2 re-seeds the v2 permission tables after a TRUNCATE.
// Mirrors the idempotent seed sections of migration
// 00005_permissions_v2_foundation.sql. Schema (CREATE TABLE / ALTER TABLE) is
// not repeated — only the reference-data INSERTs and the role-grants /
// user-roles rebuild needed by characterization tests.
func seedPermissionsV2() error {
	ctx := context.Background()
	if pool == nil {
		return fmt.Errorf("seedPermissionsV2: pgxpool not initialized")
	}

	stmts := []string{
		// permission_actions (9)
		`INSERT INTO permission_actions (code, display_ru) VALUES
			('view','Просмотр'),('create','Создание'),('edit','Изменение'),
			('delete','Удаление'),('confirm','Подтверждение'),('correct','Корректировка'),
			('import','Импорт'),('export','Экспорт'),('report','Формирование отчёта')
		 ON CONFLICT (code) DO NOTHING;`,

		// resource_types (42)
		`INSERT INTO resource_types (code, category, display_ru) VALUES
			('invoice.input','invoice','Накладная приход'),
			('invoice.output','invoice','Накладная отпуск'),
			('invoice.output_out_of_project','invoice','Накладная отпуск вне проекта'),
			('invoice.return_team','invoice','Накладная возврат из бригад'),
			('invoice.return_object','invoice','Накладная возврат из объекта'),
			('invoice.writeoff','invoice','Накладная списание'),
			('invoice.object','invoice','Накладная объект'),
			('invoice.correction','invoice','Корректировка оператора'),
			('reference.material','reference','Справочник материалов'),
			('reference.material_cost','reference','Ценники материалов'),
			('reference.material_defect','reference','Бракованные материалы'),
			('reference.material_location','reference','Местоположение материала'),
			('reference.serial_number','reference','Серийные номера'),
			('reference.worker','reference','Сотрудники'),
			('reference.team','reference','Бригады'),
			('reference.district','reference','Районы'),
			('reference.operation','reference','Сервисы / операции'),
			('reference.project','reference','Проекты'),
			('reference.object.kl04kv','reference','Объект КЛ-04 кВ'),
			('reference.object.mjd','reference','Объект МЖД'),
			('reference.object.sip','reference','Объект СИП'),
			('reference.object.stvt','reference','Объект СТВТ'),
			('reference.object.tp','reference','Объект ТП'),
			('reference.object.substation','reference','Объект подстанция'),
			('reference.object.substation_cell','reference','Объект ячейка подстанции'),
			('report.balance','report','Отчёт остатков'),
			('report.invoice','report','Отчёт по накладным'),
			('report.attendance','report','Отчёт посещаемости'),
			('report.statistics','report','Статистика'),
			('admin.user','admin','Управление пользователями'),
			('admin.user_action','admin','Журнал действий пользователей'),
			('admin.user_in_project','admin','Доступы пользователей в проекты'),
			('admin.role','admin','Роли'),
			('admin.role_grant','admin','Назначения прав ролям'),
			('admin.resource_type','admin','Управление типами ресурсов'),
			('admin.project','admin','Управление проектами'),
			('auction.bid_public','auction','Публичный аукцион'),
			('auction.bid_private','auction','Закрытый аукцион'),
			('auction.manage','auction','Управление аукционами'),
			('hr.attendance','hr','Посещаемость'),
			('system.import','system','Массовый импорт'),
			('system.material_location_live','system','Текущее местоположение материалов')
		 ON CONFLICT (code) DO NOTHING;`,

		// superadmin wildcard grants (9 × 42 = 378)
		`INSERT INTO role_grants (role_id, resource_type_code, action_code)
		 SELECT r.id, rt.code, pa.code
		 FROM roles r CROSS JOIN resource_types rt CROSS JOIN permission_actions pa
		 WHERE r.code = 'superadmin'
		 ON CONFLICT DO NOTHING;`,

		// user_roles backfill — one global row per user that has a legacy role_id
		`INSERT INTO user_roles (user_id, role_id, project_id)
		 SELECT u.id, u.role_id, NULL
		 FROM users u
		 WHERE u.role_id IS NOT NULL
		   AND NOT EXISTS (
		     SELECT 1 FROM user_roles ur
		     WHERE ur.user_id = u.id AND ur.role_id = u.role_id AND ur.project_id IS NULL
		   );`,
	}

	for _, s := range stmts {
		if _, err := pool.Exec(ctx, s); err != nil {
			return fmt.Errorf("seedPermissionsV2: %w", err)
		}
	}
	return nil
}
