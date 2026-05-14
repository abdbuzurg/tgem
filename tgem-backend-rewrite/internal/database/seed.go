package database

import (
	"fmt"
	"os"

	"gorm.io/gorm"
)

// Function for running SEED scripts
// filepath should related to the file main.go
func ExecSeedFile(db *gorm.DB, filepath string) error {

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
