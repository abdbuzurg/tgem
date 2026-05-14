// Command dump_baseline_schema runs gorm AutoMigrate against an empty
// Postgres database so that `pg_dump --schema-only` can capture the
// resulting schema as the phase-5 Goose initial migration.
//
// This binary is one-shot: once the baseline migration is checked in, it
// has no further use. It is committed alongside the migration so the
// process is reproducible (and so future maintainers can regenerate the
// dump if AutoMigrate ever produces something different again).
//
// Usage:
//
//	createdb -h 127.0.0.1 -U postgres tgem_phase5_baseline
//	go run ./cmd/dump_baseline_schema -dsn "host=127.0.0.1 user=postgres password=password dbname=tgem_phase5_baseline port=5432 sslmode=disable" -mode automigrate
//	pg_dump --schema-only --no-owner --no-privileges -h 127.0.0.1 -U postgres -d tgem_phase5_baseline
//	dropdb -h 127.0.0.1 -U postgres tgem_phase5_baseline
//
// The same binary can also exercise the goose path against a fresh DB
// (used to verify the migration baseline applies cleanly):
//
//	go run ./cmd/dump_baseline_schema -dsn ... -mode goose
package main

import (
	"flag"
	"log"

	"backend-v2/internal/database"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	dsn := flag.String("dsn", "", "Postgres DSN")
	mode := flag.String("mode", "automigrate", "one of: automigrate, goose")
	flag.Parse()

	if *dsn == "" {
		log.Fatal("--dsn is required")
	}

	db, err := gorm.Open(postgres.Open(*dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("gorm.Open: %v", err)
	}

	switch *mode {
	case "automigrate":
		if err := database.AutoMigrate(db); err != nil {
			log.Fatalf("AutoMigrate: %v", err)
		}
		log.Println("AutoMigrate complete. Run pg_dump --schema-only on the same database to capture the baseline.")
	case "goose":
		if err := database.MigrateUp(db); err != nil {
			log.Fatalf("MigrateUp: %v", err)
		}
		log.Println("Goose migrations applied to HEAD.")
	default:
		log.Fatalf("unknown --mode %q (expected automigrate or goose)", *mode)
	}
}
