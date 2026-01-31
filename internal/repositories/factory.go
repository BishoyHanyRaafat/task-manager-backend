package repositories

import (
	"database/sql"
	"fmt"
	"task_manager/internal/repositories/sqlite"

	postgres "task_manager/internal/repositories/postgres"
)

func NewUserRepository(driver string, db *sql.DB) (UserRepository, error) {
	switch driver {
	case "sqlite":
		return sqlite.NewUserRepository(db), nil
	case "postgres":
		return postgres.NewUserRepository(db), nil
	default:
		return nil, fmt.Errorf("unsupported db driver: %s", driver)
	}
}
