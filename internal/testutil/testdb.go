package testutil

import (
	"database/sql"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"task_manager/internal/db"
)

// NewSQLiteTestDB creates a temporary sqlite database, runs migrations, and returns the opened *sql.DB.
// This is useful for integration tests that want real SQL without external dependencies.
func NewSQLiteTestDB(t *testing.T) *sql.DB {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	dsn := "file:" + dbPath + "?_pragma=foreign_keys(1)"

	sqlDB, err := db.Open(db.Config{Driver: "sqlite", DSN: dsn})
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	// migrations/sqlite relative to repo root
	root := repoRoot(t)
	migrationsDir := filepath.Join(root, "migrations", "sqlite")
	require.NoError(t, db.RunMigrations("sqlite", sqlDB, migrationsDir))

	return sqlDB
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	// .../task_manager/internal/testutil/testdb.go -> go up 3 levels to .../task_manager
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

