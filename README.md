## Docs code update

```bash
"$(go env GOPATH)/bin/swag" init -g main.go --output docs --parseDependency --parseInternal
```

## Migrations (up / down)

Migrations live under:

- `migrations/sqlite`
- `migrations/postgres`

### Install the migrate CLI

```bash
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### SQLite

Notes:
- Use the **sqlite** migration folder: `-path migrations/sqlite`
- For file-based sqlite URLs, golang-migrate expects the `sqlite://` scheme.

```bash
# apply all up migrations
"$(go env GOPATH)/bin/migrate" -path migrations/sqlite -database "sqlite://task_manager.db" up

# rollback 1 migration
"$(go env GOPATH)/bin/migrate" -path migrations/sqlite -database "sqlite://task_manager.db" down 1

# rollback all migrations (DANGEROUS)
"$(go env GOPATH)/bin/migrate" -path migrations/sqlite -database "sqlite://task_manager.db" down
```

### Postgres

Notes:
- Use the **postgres** migration folder: `-path migrations/postgres`

```bash
# apply all up migrations
"$(go env GOPATH)/bin/migrate" -path migrations/postgres -database "postgres://USER:PASS@HOST:5432/DB?sslmode=disable" up

# rollback 1 migration
"$(go env GOPATH)/bin/migrate" -path migrations/postgres -database "postgres://USER:PASS@HOST:5432/DB?sslmode=disable" down 1
```

### App startup behavior

On startup, the app runs migrations **up** automatically (in debug + release) using:
- `DB_DRIVER=sqlite|postgres`
- `DB_DSN=...`

## Environment file (.env)

This project loads environment variables using `godotenv`.

- Copy `env.example` to `.env` and edit values:

```bash
cp env.example .env
```

- The app loads (in this order): `.env`, then `env.example` as fallback.
