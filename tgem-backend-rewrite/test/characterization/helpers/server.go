package helpers

import (
	"context"
	httpapp "backend-v2/internal/http"
	"backend-v2/model"
	"backend-v2/internal/config"
	"backend-v2/internal/database"
	"backend-v2/internal/security"
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	TestDBName     = "tgem_test"
	TesterUsername = "tester"
	TesterPassword = "test123"
)

var (
	server     *httptest.Server
	engine     *gin.Engine
	db         *gorm.DB
	pool       *pgxpool.Pool
	tableNames []string
	repoRoot   string
)

func DB() *gorm.DB             { return db }
func Pool() *pgxpool.Pool      { return pool }
func Server() *httptest.Server { return server }
func Engine() *gin.Engine      { return engine }
func BaseURL() string          { return server.URL + "/api" }

// Boot performs the test bring-up: chdir to repo root, configure viper, drop &
// recreate tgem_test, run AutoMigrate + seeds, and wrap httpapp.SetupRouter in
// an httptest server. Idempotent within a process.
func Boot() {
	if server != nil {
		return
	}

	root, err := findRepoRoot()
	if err != nil {
		log.Fatalf("characterization: locating repo root: %v", err)
	}
	repoRoot = root
	if err := os.Chdir(root); err != nil {
		log.Fatalf("characterization: chdir to repo root: %v", err)
	}

	config.GetConfig()
	viper.Set("Database.DBName", TestDBName)

	if err := recreateTestDB(); err != nil {
		log.Fatalf("characterization: recreating test database: %v", err)
	}

	conn, err := database.InitDB()
	if err != nil {
		log.Fatalf("characterization: InitDB: %v", err)
	}
	db = conn

	pgPool, err := database.InitPgxPool(context.Background())
	if err != nil {
		log.Fatalf("characterization: InitPgxPool: %v", err)
	}
	pool = pgPool

	if err := db.Raw(`SELECT tablename FROM pg_tables WHERE schemaname = 'public'`).Scan(&tableNames).Error; err != nil {
		log.Fatalf("characterization: listing tables: %v", err)
	}

	if err := SeedTesterUser(); err != nil {
		log.Fatalf("characterization: seeding tester user: %v", err)
	}

	engine = httpapp.SetupRouter(db, pool)
	server = httptest.NewServer(engine)
}

// Shutdown closes the httptest server and drops the test database.
func Shutdown() {
	if server != nil {
		server.Close()
		server = nil
	}
	if pool != nil {
		pool.Close()
		pool = nil
	}
	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
		db = nil
	}
	if err := dropTestDB(); err != nil {
		log.Printf("characterization: dropping test db (non-fatal): %v", err)
	}
}

// ResetDB truncates every public-schema table, re-runs the SQL seeds + the
// idempotent superadmin migration, and re-inserts the tester user. Call from
// the top of each test that mutates state.
func ResetDB() error {
	if len(tableNames) == 0 {
		return fmt.Errorf("characterization: tableNames not initialized; was Boot() called?")
	}

	quoted := make([]string, len(tableNames))
	for i, t := range tableNames {
		quoted[i] = fmt.Sprintf("%q", t)
	}
	stmt := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", strings.Join(quoted, ", "))
	if err := db.Exec(stmt).Error; err != nil {
		return fmt.Errorf("truncate: %w", err)
	}

	database.InitialMigration(db) // idempotent: re-runs project_dev/resource/superadmin SQL + permission upsert

	if err := SeedTesterUser(); err != nil {
		return fmt.Errorf("seed tester: %w", err)
	}

	if err := seedPermissionsV2(); err != nil {
		return fmt.Errorf("seed permissions v2: %w", err)
	}

	return nil
}

// RegularUserUsername / RegularUserPassword are the credentials of a
// non-superadmin test user, seeded by SeedRegularUser. Useful for tests
// that need to verify the project-membership access check fires (the
// default tester user has the Суперадмин role and bypasses the check).
const (
	RegularUserUsername = "regular"
	RegularUserPassword = "regular123"
)

// SeedRegularUser inserts a non-superadmin user with its own role and
// links it only to project 1. Idempotent. Used by TestLogin_NoAccessToProject
// and any future test that needs an account without the superadmin
// project-wildcard.
func SeedRegularUser() error {
	worker := model.Worker{
		Name:              "Regular User",
		JobTitleInProject: "Тестировщик",
		JobTitleInCompany: "QA",
		MobileNumber:      "+9920000001",
	}
	if err := db.Where("name = ?", worker.Name).FirstOrCreate(&worker).Error; err != nil {
		return err
	}

	// model.Role lacks the post-00005 `code` column, so we use raw SQL to
	// insert (and look up) the role to satisfy the NOT NULL constraint.
	if err := db.Exec(
		`INSERT INTO roles (name, description, code) VALUES (?, ?, ?)
		 ON CONFLICT DO NOTHING`,
		"Регулярный", "Обычный пользователь для тестов", "regular_test",
	).Error; err != nil {
		return err
	}
	var role model.Role
	if err := db.First(&role, "name = ?", "Регулярный").Error; err != nil {
		return err
	}

	hashed, err := security.Hash(RegularUserPassword)
	if err != nil {
		return err
	}

	user := model.User{
		Username: RegularUserUsername,
		WorkerID: worker.ID,
		RoleID:   role.ID,
		Password: string(hashed),
	}
	if err := db.Where("username = ?", RegularUserUsername).Assign(model.User{
		Password: string(hashed),
		WorkerID: worker.ID,
		RoleID:   role.ID,
	}).FirstOrCreate(&user).Error; err != nil {
		return err
	}

	uip := model.UserInProject{UserID: user.ID, ProjectID: 1}
	if err := db.Where(&uip).FirstOrCreate(&uip).Error; err != nil {
		return err
	}
	return nil
}

// SeedTesterUser inserts the standard test user with a known plaintext
// password. Idempotent — uses GORM upsert semantics. Project IDs 1
// (Test Project) and 2 (Администрирование) are linked.
func SeedTesterUser() error {
	worker := model.Worker{
		Name:              "Test User",
		JobTitleInProject: "Тестировщик",
		JobTitleInCompany: "QA",
		MobileNumber:      "+9920000000",
	}
	if err := db.Where("name = ?", worker.Name).FirstOrCreate(&worker).Error; err != nil {
		return err
	}

	var role model.Role
	if err := db.First(&role, "name = ?", "Суперадмин").Error; err != nil {
		return fmt.Errorf("locate Суперадмин role: %w", err)
	}

	hashed, err := security.Hash(TesterPassword)
	if err != nil {
		return err
	}

	user := model.User{
		Username: TesterUsername,
		WorkerID: worker.ID,
		RoleID:   role.ID,
		Password: string(hashed),
	}
	if err := db.Where("username = ?", TesterUsername).Assign(model.User{
		Password: string(hashed),
		WorkerID: worker.ID,
		RoleID:   role.ID,
	}).FirstOrCreate(&user).Error; err != nil {
		return err
	}

	for _, projectID := range []uint{1, 2} {
		uip := model.UserInProject{UserID: user.ID, ProjectID: projectID}
		if err := db.Where(&uip).FirstOrCreate(&uip).Error; err != nil {
			return err
		}
	}
	return nil
}

func recreateTestDB() error {
	admin, err := openAdminDB()
	if err != nil {
		return err
	}
	defer func() {
		if sqlDB, err := admin.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}()

	if err := admin.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS %q`, TestDBName)).Error; err != nil {
		return fmt.Errorf("drop: %w", err)
	}
	if err := admin.Exec(fmt.Sprintf(`CREATE DATABASE %q`, TestDBName)).Error; err != nil {
		return fmt.Errorf("create: %w", err)
	}
	return nil
}

func dropTestDB() error {
	admin, err := openAdminDB()
	if err != nil {
		return err
	}
	defer func() {
		if sqlDB, err := admin.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}()
	return admin.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS %q`, TestDBName)).Error
}

func openAdminDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=postgres port=%d sslmode=disable",
		viper.GetString("Database.Host"),
		viper.GetString("Database.Username"),
		viper.GetString("Database.Password"),
		viper.GetInt("Database.Port"),
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found above %s", dir)
		}
		dir = parent
	}
}

// RepoPath joins the repo root with the given relative path. Useful for tests
// that need to reach files like internal/templates/...
func RepoPath(parts ...string) string {
	return filepath.Join(append([]string{repoRoot}, parts...)...)
}
