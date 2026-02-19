# SPEC-001 — Basic Login (POST /login)

## Objective

Allow users to authenticate with email and password and receive a JWT on success.

## Context

- Service: `examples/go-microservice`
- No external packages — use stdlib only
- JWT can be a signed placeholder string for now (e.g., `"token.for.<email>"`)
- Users are stored in-memory as a `map[string]string` (email → hashed password)
- Pre-seed one user: `admin@example.com` / `secret`

## GIVEN / WHEN / THEN

- GIVEN a valid user exists (`admin@example.com` / `secret`)
  WHEN `POST /login` with body `{"email":"admin@example.com","password":"secret"}`
  THEN respond `200 OK` with `{"token":"token.for.admin@example.com"}`

- GIVEN a valid user exists
  WHEN `POST /login` with a wrong password
  THEN respond `401 Unauthorized` with `{"error":"invalid credentials"}`

- GIVEN any state
  WHEN `POST /login` with missing `email` field
  THEN respond `400 Bad Request` with `{"error":"email is required"}`

- GIVEN any state
  WHEN `POST /login` with missing `password` field
  THEN respond `400 Bad Request` with `{"error":"password is required"}`

- GIVEN any state
  WHEN `POST /login` with an empty body
  THEN respond `400 Bad Request`

## Acceptance criteria

- [ ] Returns 200 + `{"token":"..."}` for valid credentials
- [ ] Returns 401 for wrong password
- [ ] Returns 400 for missing `email`
- [ ] Returns 400 for missing `password`
- [ ] Returns 400 for empty body or malformed JSON
- [ ] All scenarios covered by table-driven tests in `handler_test.go`
- [ ] `go test -cover ./...` shows >80%

## Out of scope

- Real JWT signing (RS256, HS256) — use a placeholder token
- Persistent storage — in-memory map only
- Rate limiting
- Password hashing — plain string comparison is fine for this spec
