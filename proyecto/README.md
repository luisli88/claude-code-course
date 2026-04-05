# myapp — API REST en Go con Clean Architecture

API REST con un endpoint `GET /users` y autenticación JWT, construida sobre PostgreSQL usando únicamente la biblioteca estándar de Go.

## Contenido del repositorio

```
.
├── cmd/api/main.go                          # Punto de entrada y composición de capas
├── internal/
│   ├── domain/                              # Entidades e interfaces (sin dependencias externas)
│   │   ├── entity/user.go
│   │   ├── repository/user_repository.go   # Interfaz UserRepository
│   │   └── repository/mock_user_repository.go
│   ├── application/usecase/                 # Casos de uso: ListUsers, Login
│   ├── infrastructure/
│   │   ├── persistence/                     # PostgresUserRepo (implementa la interfaz)
│   │   └── auth/                            # JWTTokenService
│   └── presentation/                        # Handlers HTTP, DTOs, Router
├── migrations/                              # SQL ejecutado automáticamente al iniciar Postgres
├── api/requests.http                        # Peticiones de ejemplo (compatible con REST Client)
├── Dockerfile                               # Build multi-stage (golang → alpine)
├── docker-compose.yml                       # Servicios: db (PostgreSQL) + api
└── Makefile                                 # Comandos de desarrollo
```

## Prerrequisitos

| Herramienta | Versión mínima | Para qué se necesita |
|-------------|---------------|----------------------|
| Go | 1.25 | Compilar y ejecutar tests |
| Docker + Docker Compose | Docker 24 | Levantar DB y API en contenedores |
| `jq` | cualquiera | `make curl-users` (formatear respuesta JSON) |
| `golangci-lint` | 1.57 | `make lint` |

## Comandos y sus prerrequisitos

### Sin prerrequisitos (solo Go instalado)

```bash
make build          # Compila el binario en bin/api
make test           # Tests unitarios (sin DB, sin Docker)
```

### Requieren Docker corriendo

```bash
# Levantar toda la infraestructura (DB + API) en un solo paso:
make up             # Construye la imagen y arranca los contenedores
make down           # Detiene los contenedores (preserva los datos)
make down-clean     # Detiene los contenedores y borra los volúmenes (reset total)
make logs           # Muestra logs de todos los servicios en tiempo real
make logs-api       # Muestra logs solo de la API
```

### Requieren `make up` ejecutado primero

```bash
make curl-users     # curl localhost:8080/users | jq  (también requiere jq)
make db-shell       # Sesión psql interactiva dentro del contenedor
```

### Requieren solo la DB corriendo (no la API completa)

```bash
make dev            # Levanta la DB en Docker y corre la API localmente con `go run`
make test-integration  # Levanta la DB y ejecuta tests con el tag `integration`
```

### Requiere golangci-lint instalado

```bash
make lint
```

## Flujos de uso típicos

**Todo con Docker (recomendado para probar rápido):**

```bash
make up
make curl-users
```

**Desarrollo local (API fuera de Docker, DB en Docker):**

```bash
make dev            # en una terminal
# en otra terminal:
curl -s http://localhost:8080/users | jq .
```

**Reset completo de la base de datos:**

```bash
make down-clean && make up
```

**Solo tests (sin levantar nada):**

```bash
make test
```

**Tests de integración:**

```bash
make test-integration   # levanta la DB automáticamente antes de correr los tests
```

## Endpoints

| Método | Ruta | Auth | Descripción |
|--------|------|------|-------------|
| `POST` | `/login` | No | Devuelve un JWT |
| `GET` | `/users` | No | Lista usuarios con paginación |
| `GET` | `/users?page=2&page_size=10` | No | Paginación explícita |

**Credenciales de prueba** (creadas por las migraciones):

```
email: alice@example.com
password: password
```

Los archivos `api/requests.http` contienen ejemplos listos para usar con la extensión REST Client de VS Code.

## Arquitectura

El proyecto sigue Clean Architecture con la regla de dependencia estricta:

```
Presentation → Application → Domain ← Infrastructure
```

Cada capa se comunica con la siguiente únicamente a través de interfaces definidas en `domain/`. Ver [`architecture.md`](../architecture.md) para la documentación completa.

## Variables de entorno

| Variable | Valor por defecto | Descripción |
|----------|------------------|-------------|
| `DATABASE_URL` | `postgres://appuser:apppass@localhost:5432/myapp?sslmode=disable` | Cadena de conexión a PostgreSQL |
| `JWT_SECRET` | `change-me-in-production` | Clave para firmar tokens JWT |
