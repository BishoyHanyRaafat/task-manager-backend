package db

import (
	"context"
	"database/sql"
)

// DBTX is the minimal interface shared by *sql.DB and *sql.Tx.
// Repositories should depend on this so they can run either inside or outside a transaction.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
