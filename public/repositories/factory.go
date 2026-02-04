package repositories

import (
	"fmt"
	dbx "task_manager/public/db"
	"task_manager/public/repositories/sqlite"

	postgres "task_manager/public/repositories/postgres"
)

func NewUserRepositoryWithDBTX(driver string, db dbx.DBTX) (UserRepository, error) {
	switch driver {
	case "sqlite":
		return sqlite.NewUserRepository(db), nil
	case "postgres":
		return postgres.NewUserRepository(db), nil
	default:
		return nil, fmt.Errorf("unsupported db driver: %s", driver)
	}
}

func NewTeamRepositoryWithDBTX(driver string, db dbx.DBTX) (TeamRepository, error) {
	switch driver {
	case "sqlite":
		return sqlite.NewTeamRepository(db), nil
	case "postgres":
		return postgres.NewTeamRepository(db), nil
	default:
		return nil, fmt.Errorf("unsupported db driver: %s", driver)
	}
}
