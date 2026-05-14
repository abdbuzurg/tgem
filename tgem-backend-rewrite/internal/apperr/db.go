package apperr

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// FromDB maps GORM and pgx error flavors onto apperr codes. Survives the
// Phase 6 GORM→sqlc swap unchanged: pgconn.PgError is the underlying type
// both for GORM (via gorm.io/driver/postgres → pgx/v5) and for sqlc-emitted
// code, and sql.ErrNoRows is what sqlc returns for empty single-row queries.
func FromDB(err error) *Error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows) {
		return NotFound("Запись не найдена", err)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return Conflict("Запись с такими значениями уже существует", err)
		case "23503":
			return Conflict("Связанная запись не существует", err)
		case "23514":
			return InvalidInput("Нарушение ограничения данных", err)
		}
	}

	return Internal("Внутренняя ошибка базы данных", err)
}
