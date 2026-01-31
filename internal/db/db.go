package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Config struct {
	Driver string // "sqlite" | "postgres"
	DSN    string
}

func Open(cfg Config) (*sql.DB, error) {
	if cfg.Driver == "" {
		return nil, fmt.Errorf("db driver is required")
	}
	if cfg.DSN == "" {
		return nil, fmt.Errorf("db dsn is required")
	}

	db, err := sql.Open(cfg.DriverName(), cfg.DSN)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func (c Config) DriverName() string {
	// database/sql driver names
	switch c.Driver {
	case "sqlite":
		return "sqlite"
	case "postgres":
		return "pgx"
	default:
		return c.Driver
	}
}

