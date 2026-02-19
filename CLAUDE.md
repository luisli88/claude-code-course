# CLAUDE.md — Project Rules

## Stack

- Go 1.20
- Stdlib only — no external packages
- HTTP via `net/http`, logging via `log/slog`

## Code conventions

- All HTTP handlers are named functions (not inline anonymous functions)
- Return errors as JSON: `{"error": "message"}`
- Use `log/slog` for all logging — log method, path, status, duration
- Keep `main.go` as wiring only — no business logic

## Test conventions

- All tests are table-driven using `[]struct{ name, ... }{}`
- Test files live alongside the code they test (`handler_test.go`)
- Tests use `net/http/httptest` — no real HTTP server needed
- Coverage target: >80%

## What Claude must never do

- Never modify existing `*_test.go` files — only add new test cases or new test files
- Never add external packages or edit `go.mod`
- Never remove error handling or logging
- Never inline handlers into `main.go`

## Commands

```bash
# Run all tests
go test ./...

# Run a single test
go test -run <TestName> ./...

# Check coverage
go test -cover ./...
```

## Preferred diff style

- Minimal diffs — only change what the spec requires
- Do not refactor code that is not related to the current task
- One spec = one logical change = one commit
