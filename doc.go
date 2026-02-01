package main

// Developer commands (quick reference)
//
// Generate envelope DTOs for Swagger/OpenAPI (from public/dto/*Data structs):
//go:generate go run ./cmd/swaggen -g main.go -o docs --parseDependency --parseInternal

// Migrations (CLI examples; run manually):
//
// SWAG CLI:
//   swag init -g main.go --output docs --parseDependency --parseInternal
//
// SQLite:
//   migrate -path migrations/sqlite   -database "sqlite://task_manager.db" up
//   migrate -path migrations/sqlite   -database "sqlite://task_manager.db" down 1
//
// Postgres:
//   migrate -path migrations/postgres -database "postgres://USER:PASS@HOST:5432/DB?sslmode=disable" up
//   migrate -path migrations/postgres -database "postgres://USER:PASS@HOST:5432/DB?sslmode=disable" down 1
