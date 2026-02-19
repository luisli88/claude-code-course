# Week 6 — Go Backend: Microservice with >80% Coverage

## Objective

Build a production-quality Go microservice from scratch using Claude and SDD. Understand Go's patterns for HTTP handlers, error handling, testing, and project structure — not just syntax.

---

## 1. Project structure

```
examples/go-microservice/
├── main.go          ← wiring only: register routes, start server
├── handler.go       ← HTTP handlers
├── store.go         ← in-memory data store
├── handler_test.go  ← tests for handlers
├── store_test.go    ← tests for store
└── go.mod
```

Rule: `main.go` is wiring only. No logic there. Each concern gets its own file.

---

## 2. Core Go patterns

### HTTP handler signature

```go
// All handlers follow this signature
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    // parse → validate → process → respond
}

// Mounting handlers
mux := http.NewServeMux()
h := NewHandler(store)
mux.HandleFunc("POST /login", h.Login)
mux.HandleFunc("GET /health", h.Health)
```

### JSON response helpers

```go
func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
    writeJSON(w, status, map[string]string{"error": msg})
}
```

### Input validation pattern

```go
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid JSON")
        return
    }
    if req.Email == "" {
        writeError(w, http.StatusBadRequest, "email is required")
        return
    }
    if req.Password == "" {
        writeError(w, http.StatusBadRequest, "password is required")
        return
    }
    // ... business logic
}
```

### Structured logging with slog

```go
import "log/slog"

// At handler level — log every request/response
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    // ... handle request ...
    slog.Info("login",
        "method", r.Method,
        "path", r.URL.Path,
        "status", status,
        "duration_ms", time.Since(start).Milliseconds(),
        "email", req.Email,
    )
}

// At main.go level — set JSON output for production
slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
```

### In-memory store pattern

```go
type Store struct {
    mu    sync.RWMutex
    users map[string]string  // email → password
}

func NewStore() *Store {
    s := &Store{users: make(map[string]string)}
    s.users["admin@example.com"] = "secret"  // seed data
    return s
}

func (s *Store) Authenticate(email, password string) bool {
    s.mu.RLock()
    defer s.mu.RUnlock()
    stored, ok := s.users[email]
    return ok && stored == password
}
```

---

## 3. Testing patterns

### Table-driven tests (required pattern)

```go
func TestLogin(t *testing.T) {
    store := NewStore()
    h := NewHandler(store)
    mux := http.NewServeMux()
    mux.HandleFunc("POST /login", h.Login)

    tests := []struct {
        name       string
        body       string
        wantStatus int
        wantBody   string
    }{
        {
            name:       "valid credentials",
            body:       `{"email":"admin@example.com","password":"secret"}`,
            wantStatus: http.StatusOK,
            wantBody:   `"token"`,
        },
        {
            name:       "wrong password",
            body:       `{"email":"admin@example.com","password":"wrong"}`,
            wantStatus: http.StatusUnauthorized,
        },
        {
            name:       "missing email",
            body:       `{"password":"secret"}`,
            wantStatus: http.StatusBadRequest,
        },
        {
            name:       "missing password",
            body:       `{"email":"admin@example.com"}`,
            wantStatus: http.StatusBadRequest,
        },
        {
            name:       "empty body",
            body:       `{}`,
            wantStatus: http.StatusBadRequest,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(http.MethodPost, "/login",
                strings.NewReader(tt.body))
            req.Header.Set("Content-Type", "application/json")
            w := httptest.NewRecorder()

            mux.ServeHTTP(w, req)

            if w.Code != tt.wantStatus {
                t.Errorf("status: got %d, want %d\nbody: %s",
                    w.Code, tt.wantStatus, w.Body.String())
            }
            if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
                t.Errorf("body: want to contain %q\ngot: %s", tt.wantBody, w.Body.String())
            }
        })
    }
}
```

### Testing helpers

```go
// Helper to make a test request
func makeRequest(t *testing.T, mux http.Handler, method, path, body string) *httptest.ResponseRecorder {
    t.Helper()
    req := httptest.NewRequest(method, path, strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    mux.ServeHTTP(w, req)
    return w
}
```

---

## 4. Test commands

```bash
# Run all tests
go test ./...

# Run a single test function
go test -run TestLogin ./...

# Run a specific subtest
go test -run "TestLogin/valid_credentials" ./...

# Verbose output
go test -v ./...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out    # per-function breakdown
go tool cover -html=coverage.out    # open in browser

# Coverage percentage only
go test -cover ./...

# Race condition detector
go test -race ./...

# Benchmark
go test -bench=. -benchmem ./...
```

---

## 5. Build and run

```bash
# Run the server
go run .

# Build binary
go build -o server .
./server

# Build with version info
go build -ldflags="-X main.version=1.0.0" -o server .

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o server-linux .
```

---

## 6. Common mistakes and fixes

| Mistake | Fix |
|---------|-----|
| `w.WriteHeader` called after `w.Write` | Always call `WriteHeader` before writing body |
| Forgot to close request body | Use `defer r.Body.Close()` |
| Data race on shared map | Use `sync.RWMutex` in store |
| Test creates real HTTP server | Use `httptest.NewRecorder()` instead |
| Handler does too much | Split into handler (HTTP) + store (data) |
| No error path tested | Add table test rows for every error case |

---

## Exercise 1 — Implement SPEC-001 with Claude

```bash
claude "Implement specs/spec-login.md. Use the project structure: main.go for wiring, handler.go for handlers, store.go for data. Table-driven tests required. Stdlib only."

go test -cover ./...
# Must show >80%
```

## Exercise 2 — Reach 80% coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -v "100.0%"
# Find uncovered functions

claude "These functions have no test coverage: [paste output]. Add test cases for each error path."
go test -cover ./...
```

## Exercise 3 — Add middleware

```bash
claude "Add request logging middleware that wraps all handlers. Log: method, path, status code, duration. Use log/slog. Do not modify existing tests."
go test -race ./...
```

---

## Resources

- [Go testing package](https://pkg.go.dev/testing)
- [net/http/httptest](https://pkg.go.dev/net/http/httptest)
- [log/slog](https://pkg.go.dev/log/slog)
- [Go by Example](https://gobyexample.com)
- [Effective Go](https://go.dev/doc/effective_go)

---

## Pass gate

`go test -cover ./...` shows >80%, `go test -race ./...` is clean, CI is green, and every handler has a corresponding table-driven test covering at least 3 error scenarios.