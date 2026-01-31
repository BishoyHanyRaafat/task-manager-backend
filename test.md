### Testing guide (task_manager)

This project supports two testing styles:

- **Fast tests (unit-ish)** using an in-memory **fake repository**
- **Integration tests** using a real **SQLite database** created per-test and migrated automatically

All tests are written in the standard Go style:

- **Table-driven tests**
- `testing` + `httptest`
- Assertions via `testify/require`

---

### Run tests

From `task_manager/`:

```bash
go test ./...
```

Run a single test:

```bash
go test ./... -run TestSignupLoginMe_FakeRepo
```

Verbose:

```bash
go test ./... -v
```

---

### The unified HTTP test workflow

Most endpoint tests follow this pattern:

1. Build a Gin router (same routes as the app)
2. Send an HTTP request via `httptest`
3. Decode the JSON response into the unified envelope type
4. Assert `success/status_code` + payload fields

Helpers live in:

- `internal/testutil/http.go`
  - `DoJSON(...)` sends JSON requests
  - `DecodeJSON[T](...)` decodes JSON response bodies
- `internal/testutil/router.go`
  - `NewTestRouter(...)` creates a router with `/api/v1/...` routes

Example (already in code): `handlers/auth/auth_integration_test.go`

---

### Option A: Fake repo (“mock DB”, fastest)

Use this when you want fast tests without SQL.

Fake repo:

- `internal/testutil/fakes/user_repo.go`
  - Implements `repositories.UserRepository`
  - Stores users in-memory (map) with a mutex

Example skeleton:

```go
repo := fakes.NewUserRepo()
r := testutil.NewTestRouter(t, repo, "test-secret")

rr := testutil.DoJSON(t, r, http.MethodPost, "/api/v1/auth/signup", dto.SignupRequest{
  FirstName: "A", LastName: "B", Email: "a@b.com", Password: "password123",
}, nil)

env := testutil.DecodeJSON[response.Any](t, rr)
require.True(t, env.Success)
```

---

### Option B: SQLite test DB (real SQL integration tests)

Use this when you want to validate:

- migrations
- SQL queries
- repository behavior end-to-end

Helper:

- `internal/testutil/testdb.go`
  - `NewSQLiteTestDB(t)`:
    - creates a temporary sqlite DB file (`t.TempDir()`)
    - opens it with `modernc.org/sqlite`
    - runs migrations from `migrations/sqlite`

Typical pattern:

```go
sqlDB := testutil.NewSQLiteTestDB(t)
repo, err := repositories.NewUserRepository("sqlite", sqlDB)
require.NoError(t, err)

r := testutil.NewTestRouter(t, repo, "test-secret")
// ... now hit HTTP endpoints and assert behavior ...
```

---

### Response assertions (unified envelope)

All endpoints return:

```json
{
  "success": true,
  "status_code": 200,
  "data": { ... }
}
```

For tests, easiest decode is:

- `response.Any` (`Envelope[any]`)

Then assert with type assertions:

```go
env := testutil.DecodeJSON[response.Any](t, rr)
data := env.Data.(map[string]any)
token := data["access_token"].(string)
```

If you want stronger typing in tests, feel free to decode into the concrete swagger-friendly DTOs:

- `dto.EnvelopeMe`
- `dto.EnvelopeAuthToken`
- `dto.EnvelopeError`

---

### Writing new tests (recommended structure)

- Keep tests close to the code:
  - handler tests in `handlers/<name>/*_test.go`
  - repo tests in `internal/repositories/<driver>/*_test.go`
- Prefer `package x_test` (external tests) when possible.
- Use table-driven tests:

```go
tests := []struct{
  name string
  body any
  wantStatus int
}{
  {"bad email", dto.SignupRequest{FirstName:"A", LastName:"B", Email:"nope", Password:"password123"}, 400},
}
for _, tt := range tests {
  t.Run(tt.name, func(t *testing.T) {
    rr := testutil.DoJSON(t, r, http.MethodPost, "/api/v1/auth/signup", tt.body, nil)
    require.Equal(t, tt.wantStatus, rr.Code)
  })
}
```

