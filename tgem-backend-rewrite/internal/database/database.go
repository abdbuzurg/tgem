package database

import (
	"backend-v2/model"
	"fmt"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB() (*gorm.DB, error) {
	username := viper.GetString("Database.Username")
	password := viper.GetString("Database.Password")
	host := viper.GetString("Database.Host")
	port := viper.GetInt("Database.Port")
	dbname := viper.GetString("Database.DBName")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", host, username, password, dbname, port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, err
	}

	test, err := db.DB()
	if err != nil {
		return nil, err
	}

	if err := test.Ping(); err != nil {
		return nil, err
	}

	// Phase 5: schema is owned by Goose migrations under
	// internal/database/migrations/. AutoMigrate is no longer called
	// from InitDB; it is kept exported only for the one-shot
	// cmd/dump_baseline_schema tool that produced the initial migration
	// (and for any future regeneration if the model definitions and the
	// baseline ever need to be re-aligned).
	if err := MigrateUp(db); err != nil {
		return nil, err
	}
	InitialMigration(db)

	return db, nil
}

// AutoMigrate is retained from phase 4 but is no longer invoked at
// startup. It is used only by cmd/dump_baseline_schema to regenerate
// the phase-5 baseline migration (run AutoMigrate on an empty DB,
// pg_dump the result, replace 00001_initial_schema.sql).
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		model.Role{},
		model.Resource{},
		model.Project{},
		model.District{},
		model.Worker{},
		model.User{},
		model.UserAction{},
		model.UserInProject{},
		model.Material{},
		model.MaterialCost{},
		model.MaterialLocation{},
		model.MaterialDefect{},
		model.Object{},
		model.ObjectTeams{},
		model.ObjectSupervisors{},
		model.Operation{},
		model.OperationMaterial{},
		model.Permission{},
		model.SerialNumber{},
		model.SerialNumberLocation{},
		model.SerialNumberMovement{},
		model.Team{},
		model.TeamLeaders{},
		model.InvoiceMaterials{},
		model.InvoiceCount{},
		model.InvoiceInput{},
		model.InvoiceOutput{},
		model.InvoiceOutputOutOfProject{},
		model.InvoiceReturn{},
		model.InvoiceObject{},
		model.InvoiceOperations{},
		model.InvoiceObjectOperator{},
		model.InvoiceWriteOff{},
		model.OperatorErrorFound{},
		model.KL04KV_Object{},
		model.MJD_Object{},
		model.SIP_Object{},
		model.STVT_Object{},
		model.Substation_Object{},
		model.SubstationCellObject{},
		model.TP_Object{},
		model.TPNourashesObjects{},
		model.SubstationCellNourashesSubstationObject{},
		model.WorkerAttendance{},
		model.ProjectProgressMaterials{},
		model.ProjectProgressOperations{},
		model.Auction{},
		model.AuctionPackage{},
		model.AuctionItem{},
		model.AuctionParticipantPrice{},
	)
}
