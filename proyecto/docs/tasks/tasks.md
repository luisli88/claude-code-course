# Task Breakdown: Password Reset Flow

**Basado en:** plan.md
**Autor:** Luis Ricardo Ruiz
**Fecha:** 2026-05-01
**Total de tareas:** 17

---

## Índice

| ID | Tarea | Fase | Estimación | Depende de |
|----|-------|------|------------|------------|
| T01 | Entidad `PasswordResetToken` | F1 | 30min | — | ✅ |
| T02 | Interfaz `PasswordResetTokenRepository` + sentinel errors | F1 | 45min | T01 | ✅ |
| T03 | Mock `PasswordResetTokenRepository` | F1 | 30min | T02 | ✅ |
| T04 | Migración SQL `004_add_password_reset_tokens.sql` | F1 | 15min | — | ✅* |
| T05 | `PostgresPasswordResetTokenRepo` — implementación | F2 | 60min | T02, T04 |
| T06 | `PostgresPasswordResetTokenRepo` — integration tests | F2 | 60min | T05 |
| T07 | Interfaz `EmailSender` + mock | F3 | 30min | — |
| T08 | `ConsoleEmailSender` | F3 | 20min | T07 |
| T09 | `UpdatePassword` en dominio + Postgres | F5a | 45min | — |
| T10 | Caso de uso `RequestPasswordReset` | F4 | 60min | T02, T03, T07 |
| T11 | Tests unitarios `RequestPasswordReset` | F4 | 50min | T10 |
| T12 | Caso de uso `VerifyResetToken` + tests | F5 | 60min | T02, T03 |
| T13 | Caso de uso `ResetPassword` + tests | F5 | 70min | T09, T12 |
| T14 | DTOs de password reset | F6 | 20min | — |
| T15 | `PasswordResetHandler` | F6 | 60min | T10, T12, T13, T14 |
| T16 | Registrar rutas en router | F6 | 15min | T15 |
| T17 | Composition root + smoke tests E2E | F7 | 90min | T05, T15, T16 |

---

## T01 — Entidad `PasswordResetToken`

**Estimación:** 30 min
**Depende de:** —

**Descripción:**
Crear la entidad de dominio que representa un token de reinicio de contraseña.

**Archivos:**
- `proyecto/internal/domain/entity/password_reset_token.go` *(crear)*

**Acceptance criteria:**
- [ ] El struct exporta todos los campos: `ID`, `UserID`, `Token`, `ExpiresAt`, `Used`, `CreatedAt`
- [ ] Tiene un método `IsExpired() bool` que compara `ExpiresAt` con `time.Now()`
- [ ] Tiene un método `IsValid() bool` que retorna `!t.Used && !t.IsExpired()`
- [ ] Está documentado (doc comment en la struct y en cada método)

**Comando de validación:**
```bash
cd proyecto && go build ./internal/domain/entity/...
```

---

## T02 — Interfaz `PasswordResetTokenRepository` + sentinel errors

**Estimación:** 45 min
**Depende de:** T01

**Descripción:**
Definir el contrato del repositorio y los errores de dominio exportados.

**Archivos:**
- `proyecto/internal/domain/repository/password_reset_token_repository.go` *(crear)*

**Acceptance criteria:**
- [ ] La interfaz declara exactamente 5 métodos:
  - `Create(token entity.PasswordResetToken) (*entity.PasswordResetToken, error)`
  - `FindByToken(token string) (*entity.PasswordResetToken, error)`
  - `MarkAsUsed(id int) error`
  - `CountRecentByEmail(email string, since time.Time) (int, error)`
  - `InvalidatePreviousByUserID(userID int) error`
- [ ] El archivo exporta 3 sentinel errors: `ErrTokenNotFound`, `ErrTokenExpired`, `ErrTokenAlreadyUsed`
- [ ] Cada método de la interfaz tiene doc comment

**Comando de validación:**
```bash
cd proyecto && go build ./internal/domain/repository/...
```

---

## T03 — Mock `PasswordResetTokenRepository`

**Estimación:** 30 min
**Depende de:** T02

**Descripción:**
Crear el mock manual para pruebas unitarias, co-ubicado con la interfaz.

**Archivos:**
- `proyecto/internal/domain/repository/mock_password_reset_token_repository.go` *(crear)*

**Acceptance criteria:**
- [ ] El mock implementa los 5 métodos de la interfaz
- [ ] `Create` almacena el token en un slice interno y lo retorna con un `ID` de prueba
- [ ] `FindByToken` busca en el slice y retorna `ErrTokenNotFound` si no existe
- [ ] `CountRecentByEmail` cuenta tokens creados desde la fecha recibida
- [ ] El archivo tiene una assertion de interfaz: `var _ PasswordResetTokenRepository = (*MockPasswordResetTokenRepository)(nil)`
- [ ] Expone campo `Tokens []entity.PasswordResetToken` para asertar en tests

**Comando de validación:**
```bash
cd proyecto && go build ./internal/domain/repository/... && go vet ./internal/domain/repository/...
```

---

## T04 — Migración SQL `004_add_password_reset_tokens.sql`

**Estimación:** 15 min
**Depende de:** —

**Descripción:**
Crear la tabla `password_reset_tokens` con sus índices.

**Archivos:**
- `proyecto/migrations/004_add_password_reset_tokens.sql` *(crear)*

**Acceptance criteria:**
- [ ] La tabla tiene las columnas: `id`, `user_id`, `token`, `expires_at`, `used`, `created_at`
- [ ] `user_id` tiene FK a `users(id) ON DELETE CASCADE`
- [ ] `token` tiene constraint `UNIQUE`
- [ ] Existen dos índices: `idx_password_reset_tokens_token` e `idx_password_reset_tokens_user_id`

**Comando de validación:**
```bash
cd proyecto && make down-clean && make up && make db-shell
# En psql:
# \d password_reset_tokens
```

---

## T05 — `PostgresPasswordResetTokenRepo` — implementación

**Estimación:** 60 min
**Depende de:** T02, T04

**Descripción:**
Implementar el repositorio Postgres con los 5 métodos de la interfaz.

**Archivos:**
- `proyecto/internal/infrastructure/persistence/postgres_password_reset_token_repo.go` *(crear)*

**Acceptance criteria:**
- [ ] `Create` usa `INSERT … RETURNING id, created_at` — la DB genera ambos campos
- [ ] `FindByToken` mapea `sql.ErrNoRows` → `repository.ErrTokenNotFound`
- [ ] `MarkAsUsed` ejecuta `UPDATE … SET used = TRUE WHERE id = $1`
- [ ] `CountRecentByEmail` hace JOIN con `users` para filtrar por `email` y `created_at >= $2`
- [ ] `InvalidatePreviousByUserID` ejecuta `UPDATE … SET used = TRUE WHERE user_id = $1 AND used = FALSE`
- [ ] El struct tiene assertion de interfaz: `var _ repository.PasswordResetTokenRepository = (*PostgresPasswordResetTokenRepo)(nil)`

**Comando de validación:**
```bash
cd proyecto && go build ./internal/infrastructure/persistence/...
```

---

## T06 — `PostgresPasswordResetTokenRepo` — integration tests

**Estimación:** 60 min
**Depende de:** T05

**Descripción:**
Pruebas de integración que ejercitan el repo contra una DB real.

**Archivos:**
- `proyecto/internal/infrastructure/persistence/postgres_password_reset_token_repo_integration_test.go` *(crear)*

**Acceptance criteria:**
- [ ] Build tag `//go:build integration` en la primera línea
- [ ] `TestCreate_Success` verifica que el token guardado tiene `ID > 0` y `CreatedAt` no nulo
- [ ] `TestFindByToken_NotFound` verifica que retorna `ErrTokenNotFound`
- [ ] `TestMarkAsUsed_Success` verifica que `used = true` después de la llamada
- [ ] `TestCountRecentByEmail_ReturnsCorrectCount` verifica el conteo con tokens en la ventana de tiempo
- [ ] `TestInvalidatePreviousByUserID_MarksAllAsUsed` verifica que tokens previos quedan como `used = true`
- [ ] Cada test limpia su estado usando `testhelper.NewTestDB(t)`

**Comando de validación:**
```bash
cd proyecto && make test-integration 2>&1 | grep -E "PASS|FAIL|ok"
```

---

## T07 — Interfaz `EmailSender` + mock

**Estimación:** 30 min
**Depende de:** —

**Descripción:**
Definir la abstracción del servicio de correo y su mock para pruebas.

**Archivos:**
- `proyecto/internal/domain/service/email_sender.go` *(crear)*
- `proyecto/internal/domain/service/mock_email_sender.go` *(crear)*

**Acceptance criteria:**
- [ ] La interfaz `EmailSender` declara `SendPasswordResetEmail(toEmail, resetLink string) error`
- [ ] El mock implementa la interfaz y almacena la última llamada en campos exportados `LastToEmail string` y `LastResetLink string`
- [ ] El mock tiene campo `Err error` para simular fallos
- [ ] El mock tiene assertion de interfaz
- [ ] Ambos archivos documentados

**Comando de validación:**
```bash
cd proyecto && go build ./internal/domain/service/...
```

---

## T08 — `ConsoleEmailSender`

**Estimación:** 20 min
**Depende de:** T07

**Descripción:**
Implementación de desarrollo que imprime el correo en stdout en lugar de enviarlo.

**Archivos:**
- `proyecto/internal/infrastructure/email/console_email_sender.go` *(crear)*

**Acceptance criteria:**
- [ ] Implementa la interfaz `service.EmailSender`
- [ ] Imprime con `log.Printf` el correo destino y el enlace de reinicio en formato legible
- [ ] Tiene assertion de interfaz
- [ ] Retorna siempre `nil`

**Comando de validación:**
```bash
cd proyecto && go build ./internal/infrastructure/email/...
```

---

## T09 — `UpdatePassword` en dominio + Postgres

**Estimación:** 45 min
**Depende de:** —

**Descripción:**
Extender `UserRepository` con el método para actualizar el hash de contraseña.

**Archivos:**
- `proyecto/internal/domain/repository/user_repository.go` *(modificar)*
- `proyecto/internal/domain/repository/mock_user_repository.go` *(modificar)*
- `proyecto/internal/infrastructure/persistence/postgres_user_repo.go` *(modificar)*

**Acceptance criteria:**
- [ ] La interfaz `UserRepository` agrega `UpdatePassword(userID int, passwordHash string) error`
- [ ] El mock implementa `UpdatePassword` actualizando el campo en el slice interno
- [ ] El mock retorna `ErrNotFound` si el `userID` no existe
- [ ] La implementación Postgres ejecuta `UPDATE users SET password_hash = $1 WHERE id = $2`
- [ ] `go build ./...` compila sin errores (ningún implementador roto)

**Comando de validación:**
```bash
cd proyecto && go build ./... && make test
```

---

## T10 — Caso de uso `RequestPasswordReset`

**Estimación:** 60 min
**Depende de:** T02, T03, T07

**Descripción:**
Implementar la lógica de solicitud de reinicio con validación, rate limiting y envío de correo.

**Archivos:**
- `proyecto/internal/application/usecase/request_password_reset.go` *(crear)*

**Acceptance criteria:**
- [ ] Exporta `ErrInvalidEmail` y `ErrRateLimited`
- [ ] Valida que el correo contenga `@`; retorna `ErrInvalidEmail` si no
- [ ] Llama a `CountRecentByEmail` con ventana de 1 hora; retorna `ErrRateLimited` si count ≥ 3
- [ ] Si el usuario no existe, retorna `nil` sin llamar a `EmailSender` (sin enumeración)
- [ ] Llama a `InvalidatePreviousByUserID` antes de crear el token nuevo
- [ ] Genera el token con `crypto/rand` (32 bytes, codificado en hex)
- [ ] El token creado tiene `ExpiresAt = time.Now().Add(time.Hour)`
- [ ] Llama a `EmailSender.SendPasswordResetEmail` con el correo y el enlace construido

**Comando de validación:**
```bash
cd proyecto && go build ./internal/application/usecase/...
```

---

## T11 — Tests unitarios `RequestPasswordReset`

**Estimación:** 50 min
**Depende de:** T10

**Descripción:**
Pruebas unitarias que cubren todos los caminos del use case.

**Archivos:**
- `proyecto/internal/application/usecase/request_password_reset_test.go` *(crear)*

**Acceptance criteria:**
- [ ] `TestRequestPasswordReset_ValidEmail_SendsEmail` — verifica que `EmailSender` se llama con el correo correcto
- [ ] `TestRequestPasswordReset_UnknownEmail_ReturnsSuccessWithoutSendingEmail` — verifica que `EmailSender.SendPasswordResetEmail` NO se llama
- [ ] `TestRequestPasswordReset_RateLimitExceeded_ReturnsErrRateLimited`
- [ ] `TestRequestPasswordReset_InvalidEmailFormat_ReturnsErrInvalidEmail`
- [ ] `TestRequestPasswordReset_RepoCreateFails_ReturnsError`
- [ ] Cobertura ≥ 90%: `go test -cover ./internal/application/usecase/ -run RequestPasswordReset`

**Comando de validación:**
```bash
cd proyecto && go test -v -cover ./internal/application/usecase/ -run RequestPasswordReset
```

---

## T12 — Caso de uso `VerifyResetToken` + tests

**Estimación:** 60 min
**Depende de:** T02, T03

**Descripción:**
Implementar la verificación del token y sus pruebas unitarias.

**Archivos:**
- `proyecto/internal/application/usecase/verify_reset_token.go` *(crear)*
- `proyecto/internal/application/usecase/verify_reset_token_test.go` *(crear)*

**Acceptance criteria:**
- [ ] Retorna un resultado con `Valid bool` y `ExpiresAt time.Time`
- [ ] Token no encontrado → propaga `ErrTokenNotFound`
- [ ] Token expirado → propaga `ErrTokenExpired`
- [ ] Token ya usado → propaga `ErrTokenAlreadyUsed`
- [ ] Token válido → retorna `{ Valid: true, ExpiresAt: token.ExpiresAt }`
- [ ] Tests: token válido, expirado, usado, no encontrado — 4 casos mínimo
- [ ] Cobertura ≥ 90%

**Comando de validación:**
```bash
cd proyecto && go test -v -cover ./internal/application/usecase/ -run VerifyResetToken
```

---

## T13 — Caso de uso `ResetPassword` + tests

**Estimación:** 70 min
**Depende de:** T09, T12

**Descripción:**
Implementar el cambio efectivo de contraseña y sus pruebas unitarias.

**Archivos:**
- `proyecto/internal/application/usecase/reset_password.go` *(crear)*
- `proyecto/internal/application/usecase/reset_password_test.go` *(crear)*

**Acceptance criteria:**
- [ ] Valida que `newPassword` tenga al menos 8 caracteres; retorna `ErrInvalidInput` si no
- [ ] Reutiliza la lógica de `VerifyResetToken` (o llama al repo directamente para la misma validación)
- [ ] Hashea la nueva contraseña con `bcrypt`
- [ ] Llama a `UserRepository.UpdatePassword` con el hash
- [ ] Llama a `PasswordResetTokenRepository.MarkAsUsed` para invalidar el token
- [ ] Tests: contraseña corta, token inválido (expirado/usado/no encontrado), reinicio exitoso
- [ ] Cobertura ≥ 90%

**Comando de validación:**
```bash
cd proyecto && go test -v -cover ./internal/application/usecase/ -run ResetPassword
```

---

## T14 — DTOs de password reset

**Estimación:** 20 min
**Depende de:** —

**Descripción:**
Crear los structs de request/response para los tres endpoints.

**Archivos:**
- `proyecto/internal/presentation/dto/forgot_password_request.go` *(crear)*
- `proyecto/internal/presentation/dto/forgot_password_response.go` *(crear)*
- `proyecto/internal/presentation/dto/verify_reset_token_response.go` *(crear)*
- `proyecto/internal/presentation/dto/reset_password_request.go` *(crear)*
- `proyecto/internal/presentation/dto/reset_password_response.go` *(crear)*

**Acceptance criteria:**
- [ ] `ForgotPasswordRequest` — campo `Email string` con tag `json:"email"`
- [ ] `ForgotPasswordResponse` — campo `Message string` con tag `json:"message"`
- [ ] `VerifyResetTokenResponse` — campos `Valid bool` y `ExpiresAt *time.Time` con tags JSON
- [ ] `ResetPasswordRequest` — campos `Token string` y `NewPassword string` con tags JSON
- [ ] `ResetPasswordResponse` — campo `Message string` con tag `json:"message"`

**Comando de validación:**
```bash
cd proyecto && go build ./internal/presentation/dto/...
```

---

## T15 — `PasswordResetHandler`

**Estimación:** 60 min
**Depende de:** T10, T12, T13, T14

**Descripción:**
Implementar los tres handlers HTTP con mapeo de errores a códigos de estado.

**Archivos:**
- `proyecto/internal/presentation/handler/password_reset_handler.go` *(crear)*

**Acceptance criteria:**
- [ ] `ForgotPassword`: responde siempre 200 (mensaje genérico), excepto 429 si `ErrRateLimited` y 500 para errores internos
- [ ] `VerifyResetToken`: extrae `token` de la URL con `r.PathValue("token")`; responde 200 con `VerifyResetTokenResponse`, 404 si `ErrTokenNotFound`, 410 si `ErrTokenExpired` o `ErrTokenAlreadyUsed`
- [ ] `ResetPassword`: responde 200 en éxito, 400 si `ErrInvalidInput` o `ErrInvalidEmail`, 410 si token inválido/expirado/usado
- [ ] Mapeo completo de errores (sin `switch` con strings — siempre con `errors.Is`)
- [ ] El constructor `NewPasswordResetHandler` recibe los 3 use cases como interfaces

**Comando de validación:**
```bash
cd proyecto && go build ./internal/presentation/handler/...
```

---

## T16 — Registrar rutas en router

**Estimación:** 15 min
**Depende de:** T15

**Descripción:**
Agregar las tres nuevas rutas al router existente.

**Archivos:**
- `proyecto/internal/presentation/router/router.go` *(modificar)*

**Acceptance criteria:**
- [ ] `POST /api/auth/forgot-password` → `handler.ForgotPassword`
- [ ] `GET /api/auth/verify-reset-token/{token}` → `handler.VerifyResetToken`
- [ ] `POST /api/auth/reset-password` → `handler.ResetPassword`
- [ ] La función `NewRouter` (o equivalente) acepta `*PasswordResetHandler` como parámetro adicional

**Comando de validación:**
```bash
cd proyecto && go build ./internal/presentation/router/...
```

---

## T17 — Composition root + smoke tests E2E

**Estimación:** 90 min
**Depende de:** T05, T15, T16

**Descripción:**
Cablear todas las dependencias en `main.go` y validar el flujo completo.

**Archivos:**
- `proyecto/cmd/api/main.go` *(modificar)*
- `proyecto/Makefile` *(modificar — agregar target `curl-forgot-password`)*

**Acceptance criteria:**
- [ ] `main.go` construye `PostgresPasswordResetTokenRepo`, `ConsoleEmailSender` y los 3 use cases
- [ ] `APP_BASE_URL` se lee de `os.Getenv("APP_BASE_URL")` con fallback `"http://localhost:3000"`
- [ ] `make up` arranca sin errores de compilación ni de runtime
- [ ] Smoke test 1 — solicitar reinicio:
  ```bash
  curl -s -X POST localhost:8080/api/auth/forgot-password \
    -H "Content-Type: application/json" \
    -d '{"email":"<usuario registrado>"}' | jq
  # El log del servidor imprime el enlace de reinicio
  ```
- [ ] Smoke test 2 — verificar token (copiar token del log):
  ```bash
  curl -s localhost:8080/api/auth/verify-reset-token/<TOKEN> | jq
  # Responde {"valid":true,"expiresAt":"..."}
  ```
- [ ] Smoke test 3 — restablecer contraseña:
  ```bash
  curl -s -X POST localhost:8080/api/auth/reset-password \
    -H "Content-Type: application/json" \
    -d '{"token":"<TOKEN>","newPassword":"nuevaClave123"}' | jq
  # Responde {"message":"Contraseña actualizada"}
  ```
- [ ] Smoke test 4 — reusar el token:
  ```bash
  # Mismo comando anterior → HTTP 410
  ```
- [ ] `make test && make test-integration` pasan en limpio
- [ ] `make lint` no reporta nuevos warnings

**Comando de validación:**
```bash
cd proyecto && make test && make test-integration && make lint
```
