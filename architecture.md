# Arquitectura: Endpoint GET /users

## Patron: Clean Architecture (capas concentricas)

### Estructura de carpetas

```
.
├── Dockerfile                   # Build multi-stage (builder + alpine)
├── docker-compose.yml           # Orquestacion de servicios (db + api)
├── Makefile                     # Comandos de desarrollo
├── .env                         # Variables de entorno local (excluido de git)
├── .gitignore
├── migrations/
│   └── 001_create_users.sql     # Schema inicial + datos de prueba
│
├── cmd/
│   └── api/
│       └── main.go              # Punto de entrada, wiring
│
└── internal/
    ├── domain/                  # Capa de dominio (sin dependencias externas)
    │   ├── entity/
    │   │   └── user.go          # Entidad de negocio
    │   └── repository/
    │       └── user_repository.go # Interfaz (port)
    │
    ├── application/             # Casos de uso (orquestacion)
    │   └── usecase/
    │       └── list_users.go    # Caso de uso: listar usuarios
    │
    ├── infrastructure/          # Implementaciones concretas (adapters)
    │   └── persistence/
    │       └── postgres_user_repo.go
    │
    └── presentation/            # Capa HTTP (handler + routing)
        ├── handler/
        │   └── user_handler.go
        ├── dto/
        │   └── user_response.go
        └── router/
            └── router.go
```

### Flujo de una request

```
Request -> Router -> Handler -> UseCase -> Repository (interfaz)
                                               |
                                     PostgresUserRepo (implementacion)
```

---

## Capas y responsabilidades

### 1. Domain - Entidad y contrato

```go
// internal/domain/entity/user.go
package entity

import "time"

type User struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}
```

```go
// internal/domain/repository/user_repository.go
package repository

import "myapp/internal/domain/entity"

type PaginationParams struct {
    Limit  int
    Offset int
}

type UserRepository interface {
    FindAll(params PaginationParams) ([]entity.User, error)
    Count() (int, error)
}
```

### 2. Application - Caso de uso

```go
// internal/application/usecase/list_users.go
package usecase

import (
    "myapp/internal/domain/entity"
    "myapp/internal/domain/repository"
)

type ListUsersInput struct {
    Page     int
    PageSize int
}

type ListUsersOutput struct {
    Users    []entity.User `json:"users"`
    Total    int           `json:"total"`
    Page     int           `json:"page"`
    PageSize int           `json:"page_size"`
}

type ListUsers struct {
    repo repository.UserRepository
}

func NewListUsers(repo repository.UserRepository) *ListUsers {
    return &ListUsers{repo: repo}
}

func (uc *ListUsers) Execute(input ListUsersInput) (*ListUsersOutput, error) {
    offset := (input.Page - 1) * input.PageSize

    users, err := uc.repo.FindAll(repository.PaginationParams{
        Limit:  input.PageSize,
        Offset: offset,
    })
    if err != nil {
        return nil, err
    }

    total, err := uc.repo.Count()
    if err != nil {
        return nil, err
    }

    return &ListUsersOutput{
        Users:    users,
        Total:    total,
        Page:     input.Page,
        PageSize: input.PageSize,
    }, nil
}
```

### 3. Infrastructure - Implementacion concreta

```go
// internal/infrastructure/persistence/postgres_user_repo.go
package persistence

import (
    "database/sql"
    "myapp/internal/domain/entity"
    "myapp/internal/domain/repository"
)

type PostgresUserRepo struct {
    db *sql.DB
}

func NewPostgresUserRepo(db *sql.DB) *PostgresUserRepo {
    return &PostgresUserRepo{db: db}
}

func (r *PostgresUserRepo) FindAll(params repository.PaginationParams) ([]entity.User, error) {
    rows, err := r.db.Query(
        "SELECT id, name, email, created_at FROM users LIMIT $1 OFFSET $2",
        params.Limit, params.Offset,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []entity.User
    for rows.Next() {
        var u entity.User
        if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
            return nil, err
        }
        users = append(users, u)
    }
    return users, rows.Err()
}

func (r *PostgresUserRepo) Count() (int, error) {
    var count int
    err := r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
    return count, err
}
```

### 4. Presentation - Handler y rutas

```go
// internal/presentation/dto/user_response.go
package dto

import "myapp/internal/application/usecase"

type ListUsersResponse struct {
    Users    []UserDTO `json:"users"`
    Total    int       `json:"total"`
    Page     int       `json:"page"`
    PageSize int       `json:"page_size"`
}

type UserDTO struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func FromListUsersOutput(output *usecase.ListUsersOutput) ListUsersResponse {
    users := make([]UserDTO, len(output.Users))
    for i, u := range output.Users {
        users[i] = UserDTO{ID: u.ID, Name: u.Name, Email: u.Email}
    }
    return ListUsersResponse{
        Users:    users,
        Total:    output.Total,
        Page:     output.Page,
        PageSize: output.PageSize,
    }
}
```

```go
// internal/presentation/handler/user_handler.go
package handler

import (
    "encoding/json"
    "net/http"
    "strconv"

    "myapp/internal/application/usecase"
    "myapp/internal/presentation/dto"
)

type UserHandler struct {
    listUsers *usecase.ListUsers
}

func NewUserHandler(listUsers *usecase.ListUsers) *UserHandler {
    return &UserHandler{listUsers: listUsers}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    if page < 1 {
        page = 1
    }
    pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
    if pageSize < 1 {
        pageSize = 20
    }

    output, err := h.listUsers.Execute(usecase.ListUsersInput{
        Page:     page,
        PageSize: pageSize,
    })
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(dto.FromListUsersOutput(output))
}
```

### 5. Punto de entrada - Wiring

```go
// cmd/api/main.go
package main

import (
    "database/sql"
    "log"
    "net/http"
    "os"

    _ "github.com/lib/pq"

    "myapp/internal/application/usecase"
    "myapp/internal/infrastructure/persistence"
    "myapp/internal/presentation/handler"
    "myapp/internal/presentation/router"
)

func main() {
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        dsn = "postgres://user:pass@localhost/mydb?sslmode=disable"
    }

    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        log.Fatalf("cannot connect to database: %v", err)
    }

    // Wiring: infrastructure -> application -> presentation
    userRepo := persistence.NewPostgresUserRepo(db)
    listUsers := usecase.NewListUsers(userRepo)
    userHandler := handler.NewUserHandler(listUsers)

    r := router.New(userHandler)

    log.Println("Server running on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}
```

---

## Patrones aplicados

| Patron                    | Donde                                                 | Por que                            |
| ------------------------- | ----------------------------------------------------- | ---------------------------------- |
| **Repository**            | `domain/repository/` -> `infrastructure/persistence/` | Desacopla dominio de base de datos |
| **Dependency Inversion**  | UseCase depende de interfaz, no de implementacion     | Testeable, intercambiable          |
| **DTO**                   | `presentation/dto/`                                   | Controla que se expone en la API   |
| **Use Case / Interactor** | `application/usecase/`                                | Un caso de uso = un struct, SRP    |
| **Constructor Injection** | `New*()` en cada capa                                 | Dependencias explicitas, sin magia |

## Regla de dependencia

```
Presentation -> Application -> Domain <- Infrastructure
```

- **Domain** no importa nada externo.
- **Application** solo importa de Domain.
- **Infrastructure** implementa interfaces de Domain.
- **Presentation** orquesta Application.
- **cmd/api** conecta todo (composition root).

## Convenciones Go aplicadas

- `internal/` impide importaciones desde fuera del modulo.
- Interfaces en el paquete que las consume (domain define el contrato).
- Nombres de archivos en `snake_case`.
- Structs exportados, constructores con prefijo `New`.
- Sin frameworks: `net/http` + `database/sql` de la stdlib.

---

## Infraestructura local con contenedores

### Servicios

| Servicio | Imagen             | Puerto | Proposito                          |
| -------- | ------------------ | ------ | ---------------------------------- |
| **db**   | postgres:16-alpine | 5432   | Base de datos PostgreSQL           |
| **api**  | Build local (Go)   | 8080   | API HTTP con endpoint GET /users   |

### docker-compose.yml

- `db` arranca primero con healthcheck (`pg_isready`).
- `api` espera a que `db` este healthy antes de iniciar.
- Las migraciones en `migrations/` se ejecutan automaticamente al crear el contenedor de Postgres (via `/docker-entrypoint-initdb.d`).
- El volumen `pgdata` persiste los datos entre reinicios.

### Dockerfile (multi-stage)

1. **builder**: compila el binario en `golang:1.23-alpine`.
2. **runtime**: copia el binario a `alpine:3.20` (~8MB de imagen final).

### Migraciones

Los archivos `.sql` en `migrations/` se ejecutan en orden alfabetico al inicializar el contenedor de Postgres. Para resetear datos:

```bash
make down-clean   # elimina volumenes
make up           # recrea todo desde cero
```

---

## Comandos Make

| Comando              | Descripcion                                         |
| -------------------- | --------------------------------------------------- |
| `make help`          | Muestra todos los comandos disponibles              |
| `make up`            | Levanta db + api en contenedores                    |
| `make down`          | Baja los contenedores                               |
| `make down-clean`    | Baja contenedores y elimina datos (reset completo)  |
| `make logs`          | Logs de todos los servicios                         |
| `make logs-api`      | Logs solo de la API                                 |
| `make db-up`         | Levanta solo PostgreSQL                             |
| `make db-shell`      | Abre una sesion psql interactiva                    |
| `make dev`           | Levanta DB + corre la API localmente (sin Docker)   |
| `make build`         | Compila el binario en `bin/api`                     |
| `make test`          | Ejecuta tests unitarios                             |
| `make test-integration` | Tests de integracion contra la DB real           |
| `make lint`          | Ejecuta golangci-lint                               |
| `make curl-users`    | Prueba rapida del endpoint con curl + jq            |

### Flujos tipicos de desarrollo

**Levantar todo con Docker (1 comando):**

```bash
make up
make curl-users
```

**Desarrollo local (API fuera de Docker, DB en Docker):**

```bash
make dev            # en una terminal
make curl-users     # en otra terminal
```

**Reset completo:**

```bash
make down-clean && make up
```

---

## Pruebas

### Estrategia de testing por capa

```
┌─────────────────────────────────────────────────────────┐
│  Pruebas unitarias (sin I/O, sin Docker, rapidas)       │
│  ├── usecase/list_users_test.go        → logica de uso  │
│  ├── handler/user_handler_test.go      → HTTP parsing   │
│  └── dto/user_response_test.go         → mapeo entity→DTO│
├─────────────────────────────────────────────────────────┤
│  Pruebas de integracion (DB real en Docker, lentas)      │
│  └── persistence/*_integration_test.go → SQL real       │
└─────────────────────────────────────────────────────────┘
```

### Estructura de archivos de test

```
internal/
├── application/usecase/
│   ├── list_users.go
│   └── list_users_test.go              # Unitario: inyecta mock
│
├── domain/repository/
│   ├── user_repository.go              # Interfaz
│   └── mock_user_repository.go         # Mock manual (reutilizable)
│
├── infrastructure/persistence/
│   ├── postgres_user_repo.go
│   ├── postgres_user_repo_integration_test.go  # Build tag: integration
│   └── testhelper/
│       └── db.go                       # Conexion + seed para tests
│
└── presentation/
    ├── handler/
    │   ├── user_handler.go
    │   └── user_handler_test.go        # Unitario: httptest
    └── dto/
        ├── user_response.go
        └── user_response_test.go       # Unitario: mapeo
```

### Pruebas unitarias

No requieren Docker ni base de datos. Se ejecutan con:

```bash
make test
```

**Patron:** el mock de `UserRepository` vive en `domain/repository/mock_user_repository.go`. Al estar junto a la interfaz, cualquier test de cualquier capa puede importarlo sin crear dependencias circulares.

**Que se prueba en cada capa:**

| Archivo                    | Que valida                                           |
| -------------------------- | ---------------------------------------------------- |
| `list_users_test.go`       | Calculo de offset, paginacion, propagacion de errores |
| `user_handler_test.go`     | Defaults de query params, status codes, content-type  |
| `user_response_test.go`    | Mapeo entity→DTO, exclusion de campos internos       |

**Ejemplo — test del caso de uso:**

```go
func TestListUsers_Execute(t *testing.T) {
    mock := &repository.MockUserRepository{Users: users}
    uc := usecase.NewListUsers(mock)

    output, err := uc.Execute(usecase.ListUsersInput{Page: 2, PageSize: 2})
    // ...asserts sobre output.Users, output.Total, etc.
}
```

**Ejemplo — test del handler con `httptest`:**

```go
func TestUserHandler_List(t *testing.T) {
    mock := &repository.MockUserRepository{Users: users}
    uc := usecase.NewListUsers(mock)
    h := handler.NewUserHandler(uc)

    req := httptest.NewRequest(http.MethodGet, "/users?page=1&page_size=10", nil)
    rec := httptest.NewRecorder()

    h.List(rec, req)
    // ...asserts sobre rec.Code, body JSON, headers
}
```

### Pruebas de integracion

Requieren la base de datos corriendo en Docker. Se separan con un **build tag** `integration` para que `make test` no las ejecute accidentalmente.

```bash
make test-integration    # levanta DB + ejecuta tests con tag integration
```

**Build tag** al inicio del archivo:

```go
//go:build integration

package persistence_test
```

**Testhelper** (`testhelper/db.go`):
- `NewTestDB(t)` — abre conexion desde `DATABASE_URL`, hace skip si no esta definida.
- `SeedUsers(t, db, n)` — limpia la tabla, inserta N usuarios, registra cleanup automatico.

**Ejemplo — test de integracion:**

```go
func TestPostgresUserRepo_FindAll(t *testing.T) {
    db := testhelper.NewTestDB(t)
    testhelper.SeedUsers(t, db, 5)
    repo := persistence.NewPostgresUserRepo(db)

    users, err := repo.FindAll(repository.PaginationParams{Limit: 3, Offset: 0})
    // ...asserts sobre len(users), contenido, etc.
}
```

### Convenciones de testing

- Tests unitarios: `*_test.go` junto al archivo que prueban.
- Tests de integracion: `*_integration_test.go` con build tag `//go:build integration`.
- Mock manual en `domain/repository/` — sin dependencias externas (sin testify, sin mockgen).
- `testhelper/` encapsula setup de DB — cada test hace seed + cleanup automatico.
- Paquete `_test` (e.g. `package usecase_test`) para forzar tests de caja negra.
