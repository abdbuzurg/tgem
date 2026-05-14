package database

import (
	"backend-v2/model"
	"errors"
	"fmt"
	"os"

	"gorm.io/gorm"
)

func InitialMigration(db *gorm.DB) {
	err := execSeedFile(db, "./internal/database/seed/project_dev.sql")
	if err != nil {
		panic(err)
	}

	err = execSeedFile(db, "./internal/database/seed/resource.sql")
	if err != nil {
		panic(err)
	}

	if err := execSeedFile(db, "./internal/database/seed/superadmin.sql"); err != nil {
		panic(err)
	}

	if err := initialSuperadminMigration(db); err != nil {
		panic(err)
	}
}

// Function for running SEED scripts
// filepath should related to the file main.go
func execSeedFile(db *gorm.DB, filepath string) error {

	file, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("не удалось найти файл в %v: %v", filepath, err)
	}

	sql := string(file)

	err = db.Exec(sql).Error
	if err != nil {
		return fmt.Errorf("не удалось запустить изначальный скрипт seed для доступов: %v", err)
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
