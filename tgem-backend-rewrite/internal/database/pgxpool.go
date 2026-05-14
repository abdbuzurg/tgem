package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
)

// InitPgxPool opens a pgx connection pool against the same database that
// InitDB connects to via GORM. It is the connection seam for sqlc-generated
// code (internal/db). During the GORM → sqlc migration the two coexist:
// not-yet-migrated aggregates use *gorm.DB, migrated aggregates use the
// *db.Queries built from this pool.
func InitPgxPool(ctx context.Context) (*pgxpool.Pool, error) {
	username := viper.GetString("Database.Username")
	password := viper.GetString("Database.Password")
	host := viper.GetString("Database.Host")
	port := viper.GetInt("Database.Port")
	dbname := viper.GetString("Database.DBName")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", username, password, host, port, dbname)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pgxpool.Ping: %w", err)
	}
	return pool, nil
}
