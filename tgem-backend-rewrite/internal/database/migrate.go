package database

import (
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// MigrateUp applies every migration under migrations/ that hasn't been
// applied yet. It is the replacement for the gorm AutoMigrate call that
// InitDB used to perform; phase 5 onward, schema changes are versioned
// SQL files under internal/database/migrations/ rather than
// auto-generated from struct tags.
func MigrateUp(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("MigrateUp: get *sql.DB: %w", err)
	}

	goose.SetBaseFS(migrationsFS)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("MigrateUp: SetDialect: %w", err)
	}

	if err := goose.Up(sqlDB, "migrations"); err != nil {
		return fmt.Errorf("MigrateUp: goose.Up: %w", err)
	}

	return nil
}
