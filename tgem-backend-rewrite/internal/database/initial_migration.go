package database

import (
	"backend-v2/model"
	_ "embed"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

//go:embed seed/project_dev.sql
var projectDevSeedSQL string

//go:embed seed/resource.sql
var resourceSeedSQL string

//go:embed seed/superadmin.sql
var superadminSeedSQL string

func InitialMigration(db *gorm.DB) {
	if err := execSeed(db, "seed/project_dev.sql", projectDevSeedSQL); err != nil {
		panic(err)
	}

	if err := execSeed(db, "seed/resource.sql", resourceSeedSQL); err != nil {
		panic(err)
	}

	if err := execSeed(db, "seed/superadmin.sql", superadminSeedSQL); err != nil {
		panic(err)
	}

	if err := initialSuperadminMigration(db); err != nil {
		panic(err)
	}
}

func execSeed(db *gorm.DB, name, sql string) error {
	if err := db.Exec(sql).Error; err != nil {
		return fmt.Errorf("не удалось запустить seed-скрипт %s: %v", name, err)
	}
	return nil
}

func initialSuperadminMigration(db *gorm.DB) error {

	role := model.Role{}
	err := db.First(&role, "name = 'Суперадмин'").Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		role = model.Role{Name: "Суперадмин", Description: "Суперадмин"}
		if err := db.Create(&role).Error; err != nil {
			return err
		}
	}

	resources := []model.Resource{}
	if err := db.Find(&resources).Error; err != nil {
		return err
	}

	permissionBasedOnAllResources := []model.Permission{}
	for _, resource := range resources {
		permissionBasedOnAllResources = append(permissionBasedOnAllResources, model.Permission{
			RoleID:     role.ID,
			ResourceID: resource.ID,
			R:          true,
			U:          true,
			W:          true,
			D:          true,
		})
	}

	alreadyInDBPermissions := []model.Permission{}
	if err := db.Find(&alreadyInDBPermissions, "role_id = ?", role.ID).Error; err != nil {
		return err
	}

	newSuperAdminPermissions := []model.Permission{}
	for _, newPermission := range permissionBasedOnAllResources {

		exist := false
		for _, oldPermission := range alreadyInDBPermissions {
			if newPermission.ResourceID == oldPermission.ResourceID {
				exist = true
				break
			}
		}

		if exist {
			continue
		}

		newSuperAdminPermissions = append(newSuperAdminPermissions, newPermission)
	}

	if err := db.CreateInBatches(&newSuperAdminPermissions, 10).Error; err != nil {
		return err
	}

	return nil
}
