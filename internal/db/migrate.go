package db

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(driver string, sqlDB *sql.DB, migrationsDir string) error {
	if migrationsDir == "" {
		return fmt.Errorf("migrationsDir is required")
	}
	abs, err := filepath.Abs(migrationsDir)
	if err != nil {
		return err
	}

	var dbDriver database.Driver
	var dbName string

	switch driver {
	case "sqlite":
		dbName = "sqlite"
		d, err := sqlite.WithInstance(sqlDB, &sqlite.Config{})
		if err != nil {
			return err
		}
		dbDriver = d
	case "postgres":
		dbName = "postgres"
		d, err := postgres.WithInstance(sqlDB, &postgres.Config{})
		if err != nil {
			return err
		}
		dbDriver = d
	default:
		return fmt.Errorf("unsupported db driver for migrations: %s", driver)
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+abs, dbName, dbDriver)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
