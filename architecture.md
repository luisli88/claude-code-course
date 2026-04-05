# Arquitectura: API REST con Clean Architecture

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
│   ├── 001_create_users.sql     # Schema inicial + datos de prueba
│   └── 002_add_password_hash.sql # Columna password_hash
│
├── cmd/
│   └── api/
│       └── main.go              # Punto de entrada, wiring
│
└── internal/
    ├── domain/                  # Capa de dominio (sin dependencias externas)
    │   ├── entity/
    │   │   └── user.go          # Entidad de negocio
    │   ├── repository/
    │   │   ├── user_repository.go      # Interfaz (port) + ErrNotFound
    │   │   └── mock_user_repository.go # Mock manual reutilizable
    │   └── service/
    │       └── token_service.go        # Interfaz TokenService
    │
    ├── application/             # Casos de uso (orquestacion)
    │   └── usecase/
    │       ├── list_users.go    # Caso de uso: listar usuarios (paginado)
    │       ├── list_users_test.go
    │       ├── login.go         # Caso de uso: autenticar usuario
    │       ├── register.go      # Caso de uso: registrar usuario
    │       └── register_test.go
    │
    ├── infrastructure/          # Implementaciones concretas (adapters)
    │   ├── auth/
    │   │   └── jwt_token_service.go    # Implementacion JWT de TokenService
    │   └── persistence/
    │       ├── postgres_user_repo.go
    │       ├── postgres_user_repo_integration_test.go
    │       └── testhelper/
    │           └── db.go               # Conexion + seed para tests
    │
    └── presentation/            # Capa HTTP (handler + routing)
        ├── handler/
        │   ├── user_handler.go
        │   ├── user_handler_test.go
        │   └── auth_handler.go         # Login + Register
        ├── dto/
        │   ├── user_response.go
        │   ├── user_response_test.go
        │   ├── login_request.go
        │   ├── login_response.go
        │   ├── register_request.go     # { name, email, password }
        │   └── register_response.go   # { id, name, email, created_at }
        └── router/
            └── router.go
```

### Flujo de una request

```
Request -> Router -> Handler -> UseCase -> Repository (interfaz)
                                               |
                                     PostgresUserRepo (implementacion)
```

**Endpoints disponibles**

| Metodo | Ruta        | Handler               | Caso de uso   |
| ------ | ----------- | --------------------- | ------------- |
| GET    | /users      | UserHandler.List      | ListUsers     |
| POST   | /login      | AuthHandler.Login     | Login         |
| POST   | /register   | AuthHandler.Register  | Register      |

---

## Capas y responsabilidades

### 1. Domain - Entidad y contrato

```go
// internal/domain/entity/user.go
package entity

import "time"

type User struct {
    ID           string    `json:"id"`
    Name         string    `json:"name"`
    Email        string    `json:"email"`
    CreatedAt    time.Time `json:"created_at"`
    PasswordHash string    `json:"-"`
}
```

```go
// internal/domain/repository/user_repository.go
package repository

import (
    "errors"
    "myapp/internal/domain/entity"
)

var ErrNotFound = errors.New("not found")

type PaginationParams struct {
    Limit  int
    Offset int
}

type UserRepository interface {
    FindAll(params PaginationParams) ([]entity.User, error)
    Count() (int, error)
    FindByEmail(email string) (*entity.User, error)  // devuelve ErrNotFound si no existe
    Create(user entity.User) (*entity.User, error)   // DB genera id y created_at
}
```

El sentinel `ErrNotFound` permite que los casos de uso distingan "no encontrado" de otros errores de base de datos, sin acoplarse a `database/sql`.

### 2. Application - Casos de uso

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

**Caso de uso: Register** — registrar un nuevo usuario

```go
// internal/application/usecase/register.go
package usecase

var (
    ErrEmailAlreadyTaken = errors.New("email already taken")
    ErrInvalidInput      = errors.New("invalid input")
)

type RegisterInput struct {
    Name     string
    Email    string
    Password string
}

type RegisterOutput struct {
    User entity.User
}

type Register struct {
    repo repository.UserRepository
}

func NewRegister(repo repository.UserRepository) *Register {
    return &Register{repo: repo}
}

func (uc *Register) Execute(input RegisterInput) (*RegisterOutput, error) {
    // 1. Validacion de entrada
    if err := validateRegisterInput(input); err != nil {
        return nil, err  // ErrInvalidInput wrapeado con mensaje especifico
    }

    // 2. Verificar unicidad del email
    _, err := uc.repo.FindByEmail(input.Email)
    if err == nil {
        return nil, ErrEmailAlreadyTaken
    }
    if !errors.Is(err, repository.ErrNotFound) {
        return nil, err  // error de infraestructura, propagar
    }

    // 3. Hashear contrasena
    hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }

    // 4. Persistir
    created, err := uc.repo.Create(entity.User{
        Name:         input.Name,
        Email:        input.Email,
        PasswordHash: string(hash),
    })
    if err != nil {
        return nil, err
    }

    return &RegisterOutput{User: *created}, nil
}
```

**Reglas de validacion** (funcion privada `validateRegisterInput`):

| Campo    | Regla                          | Error                                          |
| -------- | ------------------------------ | ---------------------------------------------- |
| name     | no vacio (despues de TrimSpace)| `invalid input: name is required`              |
| email    | contiene `@` y no es vacio     | `invalid input: email is invalid`              |
| password | longitud ≥ 8                   | `invalid input: password must be at least 8 characters` |

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

// FindByEmail mapea sql.ErrNoRows -> repository.ErrNotFound
func (r *PostgresUserRepo) FindByEmail(email string) (*entity.User, error) {
    var u entity.User
    err := r.db.QueryRow(
        "SELECT id, name, email, created_at, password_hash FROM users WHERE email = $1",
        email,
    ).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.PasswordHash)
    if errors.Is(err, sql.ErrNoRows) {
        return nil, repository.ErrNotFound
    }
    if err != nil {
        return nil, err
    }
    return &u, nil
}

// Create inserta el usuario y retorna la fila completa con id y created_at generados por Postgres
func (r *PostgresUserRepo) Create(user entity.User) (*entity.User, error) {
    var created entity.User
    err := r.db.QueryRow(
        "INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3) RETURNING id, name, email, created_at",
        user.Name, user.Email, user.PasswordHash,
    ).Scan(&created.ID, &created.Name, &created.Email, &created.CreatedAt)
    if err != nil {
        return nil, err
    }
    return &created, nil
}
```

> `id` y `created_at` son generados por Postgres (`gen_random_uuid()` y `now()`). La capa de aplicacion nunca genera ni asume estos valores.

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

**AuthHandler — Login y Register** en el mismo handler de autenticacion:

```go
// internal/presentation/handler/auth_handler.go
type AuthHandler struct {
    login    *usecase.Login
    register *usecase.Register
}

func NewAuthHandler(login *usecase.Login, register *usecase.Register) *AuthHandler {
    return &AuthHandler{login: login, register: register}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    var req dto.RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }

    output, err := h.register.Execute(usecase.RegisterInput{
        Name: req.Name, Email: req.Email, Password: req.Password,
    })
    if err != nil {
        switch {
        case errors.Is(err, usecase.ErrInvalidInput):
            http.Error(w, err.Error(), http.StatusBadRequest)       // 400
        case errors.Is(err, usecase.ErrEmailAlreadyTaken):
            http.Error(w, "email already taken", http.StatusConflict) // 409
        default:
            http.Error(w, "internal server error", http.StatusInternalServerError) // 500
        }
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated) // 201
    json.NewEncoder(w).Encode(dto.RegisterResponse{
        ID: output.User.ID, Name: output.User.Name,
        Email: output.User.Email, CreatedAt: output.User.CreatedAt,
    })
}
```

**Mapa de codigos HTTP del endpoint `/register`**

| Caso                      | Codigo HTTP |
| ------------------------- | ----------- |
| Registro exitoso          | 201 Created |
| Cuerpo JSON invalido      | 400         |
| Validacion fallida        | 400 (con mensaje especifico) |
| Email ya registrado       | 409 Conflict |
| Error de infraestructura  | 500         |

**DTOs de registro**

```go
// dto/register_request.go
type RegisterRequest struct {
    Name     string `json:"name"`
    Email    string `json:"email"`
    Password string `json:"password"`
}

// dto/register_response.go
type RegisterResponse struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}
```

> `PasswordHash` nunca se incluye en ninguna respuesta — la etiqueta `json:"-"` en la entidad lo excluye automaticamente.

### 5. Punto de entrada - Wiring

```go
// cmd/api/main.go
func main() {
    // Configuracion desde entorno
    dsn := os.Getenv("DATABASE_URL")      // DATABASE_URL o default local
    jwtSecret := os.Getenv("JWT_SECRET")  // JWT_SECRET o "change-me-in-production"

    db, _ := sql.Open("postgres", dsn)
    db.Ping()

    // Wiring: infrastructure -> application -> presentation
    userRepo := persistence.NewPostgresUserRepo(db)

    listUsers  := usecase.NewListUsers(userRepo)
    login      := usecase.NewLogin(userRepo, auth.NewJWTTokenService(jwtSecret))
    register   := usecase.NewRegister(userRepo)

    userHandler := handler.NewUserHandler(listUsers)
    authHandler := handler.NewAuthHandler(login, register)

    r := router.New(userHandler, authHandler)
    http.ListenAndServe(":8080", r)
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
│  ├── usecase/list_users_test.go        → logica de paginacion    │
│  ├── usecase/register_test.go          → validacion, hash, errores│
│  ├── handler/user_handler_test.go      → HTTP parsing            │
│  └── dto/user_response_test.go         → mapeo entity→DTO        │
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
│   ├── list_users_test.go              # Unitario: paginacion y errores
│   ├── register.go
│   └── register_test.go               # Unitario: validacion, hash, duplicado, error repo
│
├── domain/repository/
│   ├── user_repository.go              # Interfaz + ErrNotFound
│   └── mock_user_repository.go         # Mock manual (FindAll, Count, FindByEmail, Create)
│
├── infrastructure/persistence/
│   ├── postgres_user_repo.go
│   ├── postgres_user_repo_integration_test.go  # FindAll, Count, Create, FindByEmail
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

| Archivo                    | Que valida                                                         |
| -------------------------- | ------------------------------------------------------------------ |
| `list_users_test.go`       | Calculo de offset, paginacion, propagacion de errores              |
| `register_test.go`         | Validacion de campos, hash de contrasena, email duplicado, errores de repo |
| `user_handler_test.go`     | Defaults de query params, status codes, content-type               |
| `user_response_test.go`    | Mapeo entity→DTO, exclusion de campos internos                     |

**Casos cubiertos en `register_test.go`:**

| Test                                  | Verifica                                                  |
| ------------------------------------- | --------------------------------------------------------- |
| creates user successfully             | ID generado, nombre/email correctos, hash ≠ password plano |
| returns error when email already taken | `ErrEmailAlreadyTaken` cuando `FindByEmail` devuelve usuario |
| returns invalid input when name empty | `ErrInvalidInput` wrapeado                                |
| returns invalid input when bad email  | `ErrInvalidInput` wrapeado                                |
| returns invalid input when short pwd  | `ErrInvalidInput` wrapeado                                |
| propagates repository error on Create | error de infra se propaga sin envolver                    |

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

**Tests de integracion en `postgres_user_repo_integration_test.go`:**

| Test                                             | Verifica                                             |
| ------------------------------------------------ | ---------------------------------------------------- |
| `TestPostgresUserRepo_FindAll`                   | Limit, offset, pagina vacia                          |
| `TestPostgresUserRepo_Count`                     | Conteo exacto despues de seed                        |
| `TestPostgresUserRepo_Create/success`            | INSERT devuelve id y created_at generados por Postgres |
| `TestPostgresUserRepo_Create/duplicate email`    | Error al insertar email duplicado (UNIQUE constraint) |
| `TestPostgresUserRepo_Create/FindByEmail 404`    | `ErrNotFound` cuando el email no existe              |

### Convenciones de testing

- Tests unitarios: `*_test.go` junto al archivo que prueban.
- Tests de integracion: `*_integration_test.go` con build tag `//go:build integration`.
- Mock manual en `domain/repository/` — sin dependencias externas (sin testify, sin mockgen).
- `testhelper/` encapsula setup de DB — cada test hace seed + cleanup automatico.
- Paquete `_test` (e.g. `package usecase_test`) para forzar tests de caja negra.
- Mocks con comportamiento selectivo: usar struct auxiliar en el mismo archivo de test cuando `MockUserRepository.Err` afectaria multiples metodos (ver `createErrorMock` en `register_test.go`).
