# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

This is a Go REST API course project (`proyecto/`) demonstrating Clean Architecture with a `GET /users`, `POST /login`, and `POST /register` endpoints backed by PostgreSQL.

- Module: `myapp`
- External dependencies: `github.com/lib/pq` (Postgres driver), `github.com/golang-jwt/jwt/v5`, `golang.org/x/crypto`
- No frameworks: stdlib `net/http` + `database/sql` only

## Commands

Run from the `proyecto/` directory.

```bash
make up              # Start db + api in Docker
make down            # Stop containers
make down-clean      # Stop containers and wipe volumes (full reset)
make dev             # Start DB in Docker, run API locally (no hot-reload)
make build           # Compile binary to bin/api
make test            # Unit tests (no DB required)
make test-integration  # Integration tests (requires DB running)
make lint            # golangci-lint
make curl-users      # Quick smoke test: curl localhost:8080/users | jq
make db-shell        # psql session into the running DB container
```

## Architecture

> Full architecture documentation: [`architecture.md`](./architecture.md)

Clean Architecture with strict dependency rule:

```
Presentation → Application → Domain ← Infrastructure
```

- **`cmd/api/main.go`** — composition root; wires all layers together
- **`internal/domain/`** — entities and repository interfaces (no external deps)
- **`internal/application/usecase/`** — one struct per use case, depends only on domain interfaces
- **`internal/infrastructure/persistence/`** — concrete `PostgresUserRepo` implementing domain interface
- **`internal/presentation/`** — HTTP handlers, DTOs, router; calls use cases

The mock for `UserRepository` lives in `internal/domain/repository/mock_user_repository.go` — co-located with the interface so any layer can import it without circular deps.

## Testing

- **Unit tests** (`*_test.go`): no DB, no Docker — use the manual mock in `domain/repository/`
- **Integration tests** (`*_integration_test.go`): require DB, gated by `//go:build integration`; `testhelper.NewTestDB(t)` skips if `DATABASE_URL` is unset
- Test packages use `package X_test` (black-box style)

## Migrations

SQL files in `migrations/` run automatically via Postgres's `/docker-entrypoint-initdb.d` on container creation. To re-run: `make down-clean && make up`.

## Implemented features

### POST /register

Full Clean Architecture implementation of user registration. Touch points per layer:

**Domain**
- `user_repository.go`: Added `ErrNotFound` sentinel and `Create(entity.User) (*entity.User, error)` to `UserRepository` interface.
- `mock_user_repository.go`: Implements `Create` (appends with mock ID and timestamp); `FindByEmail` now returns `ErrNotFound` instead of a raw untyped error.

**Application**
- `usecase/register.go`: `Register` use case with input validation, email uniqueness check, bcrypt hashing, and repo delegation.
  - Sentinel errors exported: `ErrInvalidInput`, `ErrEmailAlreadyTaken`.
  - Validation rules: name non-empty, email contains `@`, password ≥ 8 characters.
  - Email uniqueness: calls `FindByEmail`; if `ErrNotFound` → email is free; if `nil` error → email taken.
- `usecase/register_test.go`: Unit tests — success path, duplicate email, all three validation failures, repo error propagation on `Create`.

**Infrastructure**
- `persistence/postgres_user_repo.go`: `Create` uses `INSERT … RETURNING` so the DB generates `id` and `created_at`; `FindByEmail` maps `sql.ErrNoRows` → `repository.ErrNotFound`.
- `persistence/postgres_user_repo_integration_test.go`: Tests for `Create` success, duplicate email error, and `FindByEmail` returning `ErrNotFound`.

**Presentation**
- `dto/register_request.go`: `{ name, email, password }`.
- `dto/register_response.go`: `{ id, name, email, created_at }`.
- `handler/auth_handler.go`: `Register` method — `201 Created` on success, `400` for validation errors, `409 Conflict` for duplicate email, `500` otherwise. `NewAuthHandler` now accepts `*usecase.Register` as second argument.
- `router/router.go`: `POST /register` route added.
- `cmd/api/main.go`: `Register` use case wired at composition root.
