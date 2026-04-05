# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

This is a Go REST API course project (`proyecto/`) demonstrating Clean Architecture with a single `GET /users` endpoint backed by PostgreSQL.

- Module: `myapp`
- Single external dependency: `github.com/lib/pq` (Postgres driver)
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
