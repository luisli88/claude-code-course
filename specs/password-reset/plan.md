# Implementation Plan: Password Reset Flow

**Basado en:** spec.md
**Autor:** Luis Ricardo Ruiz
**Fecha:** 2026-05-01
**Estimación total:** ~19 horas en 7 fases

---

## Resumen de fases

| Fase | Nombre | Estimación | Depende de |
|------|--------|------------|------------|
| 1 | Capa de Dominio + Migración | 2h | — |
| 2 | Infraestructura Postgres | 3h | Fase 1 |
| 3 | Servicio de envío de correo | 3h | — |
| 4 | Caso de uso `RequestPasswordReset` | 3h | Fase 1, 3 |
| 5 | Casos de uso `VerifyResetToken` + `ResetPassword` | 3h | Fase 1 |
| 6 | Capa de Presentación | 3h | Fase 4, 5 |
| 7 | Composition root e integración E2E | 2h | Fase 2, 6 |

**Grafo de dependencias:**

```
Fase 1 ──┬─► Fase 2 ──┐
         ├─► Fase 4 ──┤
         └─► Fase 5 ──┤
Fase 3 ──┘            │
                      ├─► Fase 7
Fase 4, 5 ─► Fase 6 ──┘
```

> Las fases 1 y 3 pueden ejecutarse en paralelo. Las fases 4 y 5 dependen de la 1 (y la 4 también de la 3).

---

## Fase 1: Capa de Dominio + Migración (2h)

**Objetivo:** Definir la entidad `PasswordResetToken`, su repositorio y el schema de base de datos.

**Tareas:**

1. Crear entidad `PasswordResetToken` en `internal/domain/entity/`
2. Crear interfaz `PasswordResetTokenRepository` en `internal/domain/repository/` con los métodos:
   - `Create(token entity.PasswordResetToken) (*entity.PasswordResetToken, error)`
   - `FindByToken(token string) (*entity.PasswordResetToken, error)`
   - `MarkAsUsed(id int) error`
   - `CountRecentByEmail(email string, since time.Time) (int, error)` — para rate limiting
   - `InvalidatePreviousByUserID(userID int) error` — invalidar tokens previos
3. Crear mock manual `mock_password_reset_token_repository.go` co-ubicado con la interfaz
4. Escribir migración SQL `004_add_password_reset_tokens.sql`
5. Exportar sentinel errors: `ErrTokenNotFound`, `ErrTokenExpired`, `ErrTokenAlreadyUsed`

**Archivos a crear:**

- `proyecto/internal/domain/entity/password_reset_token.go`
- `proyecto/internal/domain/repository/password_reset_token_repository.go`
- `proyecto/internal/domain/repository/mock_password_reset_token_repository.go`
- `proyecto/migrations/004_add_password_reset_tokens.sql`

**Validation checkpoints:**

- [ ] `go build ./...` compila sin errores
- [ ] `make down-clean && make up` aplica la migración nueva sin errores
- [ ] `make db-shell` muestra `\d password_reset_tokens` con todas las columnas e índices
- [ ] El mock implementa toda la interfaz (verificable con assertion `var _ repository.PasswordResetTokenRepository = (*Mock...)(nil)`)

---

## Fase 2: Infraestructura Postgres (3h)

**Objetivo:** Implementar `PostgresPasswordResetTokenRepo` con pruebas de integración.

**Tareas:**

1. Implementar `PostgresPasswordResetTokenRepo` con los 5 métodos de la interfaz
2. Mapear `sql.ErrNoRows` → `repository.ErrTokenNotFound` en `FindByToken`
3. Usar `INSERT … RETURNING` para `Create` (DB genera `id` y `created_at`)
4. Escribir pruebas de integración con build tag `integration`

**Archivos a crear:**

- `proyecto/internal/infrastructure/persistence/postgres_password_reset_token_repo.go`
- `proyecto/internal/infrastructure/persistence/postgres_password_reset_token_repo_integration_test.go`

**Validation checkpoints:**

- [ ] `make test-integration` pasa todas las pruebas
- [ ] Cobertura de los 5 métodos: success path + error path en `FindByToken` cuando no existe
- [ ] Verificar manualmente con `psql` que los tokens se insertan con `expires_at` correcto

**Dependencias:** Fase 1 (necesita la interfaz y la migración aplicada).

---

## Fase 3: Servicio de envío de correo (3h)

**Objetivo:** Abstraer el envío de correos detrás de una interfaz para poder mockearlo en pruebas.

**Tareas:**

1. Definir interfaz `EmailSender` en `internal/domain/service/` con método:
   - `SendPasswordResetEmail(toEmail, resetLink string) error`
2. Crear mock `mock_email_sender.go` co-ubicado con la interfaz
3. Implementación inicial: `ConsoleEmailSender` en `internal/infrastructure/email/` que solo loguea el correo y enlace en stdout (suficiente para desarrollo)
4. Definir variable de entorno `APP_BASE_URL` para construir el enlace de reinicio

**Archivos a crear:**

- `proyecto/internal/domain/service/email_sender.go`
- `proyecto/internal/domain/service/mock_email_sender.go`
- `proyecto/internal/infrastructure/email/console_email_sender.go`

**Validation checkpoints:**

- [ ] `go build ./...` compila sin errores
- [ ] El mock permite asertar el correo destino y el enlace enviado
- [ ] La implementación de consola imprime un mensaje legible al ejecutarse

**Dependencias:** ninguna (puede correr en paralelo con Fase 1 y 2).

---

## Fase 4: Caso de uso `RequestPasswordReset` (3h)

**Objetivo:** Implementar la lógica de solicitud de reinicio con validación y rate limiting.

**Tareas:**

1. Crear `RequestPasswordReset` en `internal/application/usecase/` con dependencias:
   - `UserRepository` (para buscar el usuario por correo)
   - `PasswordResetTokenRepository` (para crear/contar tokens)
   - `EmailSender` (para enviar el correo)
2. Lógica:
   - Validar formato del correo (contiene `@`)
   - Verificar rate limit: `CountRecentByEmail` ≥ 3 en la última hora → error `ErrRateLimited`
   - Buscar usuario por correo: si no existe, retornar éxito sin enviar correo (sin enumeración)
   - Invalidar tokens previos del usuario
   - Generar token con `crypto/rand` (32 bytes hex)
   - Persistir token con `expires_at = now + 1h`
   - Enviar correo con enlace `${APP_BASE_URL}/reset-password?token=...`
3. Exportar sentinel errors: `ErrInvalidEmail`, `ErrRateLimited`
4. Escribir pruebas unitarias con mocks

**Archivos a crear:**

- `proyecto/internal/application/usecase/request_password_reset.go`
- `proyecto/internal/application/usecase/request_password_reset_test.go`

**Validation checkpoints:**

- [ ] `make test` pasa
- [ ] Casos cubiertos: correo válido (envía), correo inexistente (no envía pero retorna éxito), rate limit excedido, formato inválido, fallo al guardar token
- [ ] El mock de `EmailSender` confirma que NO se llama cuando el correo no existe
- [ ] Cobertura ≥ 90% del archivo

**Dependencias:** Fase 1 (entity + repo) y Fase 3 (email sender).

---

## Fase 5: Casos de uso `VerifyResetToken` + `ResetPassword` (3h)

**Objetivo:** Implementar la verificación del token y el cambio efectivo de contraseña.

**Tareas:**

1. `VerifyResetToken` use case:
   - Buscar token por valor → si no existe → `ErrTokenNotFound`
   - Validar `expires_at > now` → si no → `ErrTokenExpired`
   - Validar `used == false` → si no → `ErrTokenAlreadyUsed`
   - Retornar `{ Valid: true, ExpiresAt: ... }`
2. `ResetPassword` use case:
   - Validar nueva contraseña (mínimo 8 caracteres)
   - Reutilizar lógica de verificación del token
   - Hashear nueva contraseña con bcrypt
   - Actualizar `users.password_hash` (extender `UserRepository` con `UpdatePassword(userID int, hash string) error`)
   - Marcar token como usado
3. Pruebas unitarias para ambos casos de uso

**Archivos a crear:**

- `proyecto/internal/application/usecase/verify_reset_token.go`
- `proyecto/internal/application/usecase/verify_reset_token_test.go`
- `proyecto/internal/application/usecase/reset_password.go`
- `proyecto/internal/application/usecase/reset_password_test.go`

**Archivos a modificar:**

- `proyecto/internal/domain/repository/user_repository.go` — agregar `UpdatePassword`
- `proyecto/internal/domain/repository/mock_user_repository.go` — implementar `UpdatePassword`
- `proyecto/internal/infrastructure/persistence/postgres_user_repo.go` — implementar `UpdatePassword`

**Validation checkpoints:**

- [ ] `make test` pasa todas las pruebas unitarias
- [ ] `make test-integration` pasa (incluyendo el nuevo `UpdatePassword`)
- [ ] Casos cubiertos: token válido, expirado, usado, no encontrado, contraseña corta, reinicio exitoso, doble uso del token
- [ ] Cobertura ≥ 90% en ambos archivos

**Dependencias:** Fase 1.

---

## Fase 6: Capa de Presentación (3h)

**Objetivo:** Exponer los tres endpoints HTTP siguiendo el patrón de la capa existente.

**Tareas:**

1. Crear DTOs:
   - `ForgotPasswordRequest` — `{ email }`
   - `ForgotPasswordResponse` — `{ message }`
   - `VerifyResetTokenResponse` — `{ valid, expiresAt }`
   - `ResetPasswordRequest` — `{ token, newPassword }`
   - `ResetPasswordResponse` — `{ message }`
2. Crear `PasswordResetHandler` en `internal/presentation/handler/` con métodos:
   - `ForgotPassword` — siempre responde 200 con mensaje genérico (o 429 si rate limited)
   - `VerifyResetToken` — extrae token de la URL, llama al use case, responde 200
   - `ResetPassword` — responde 200 en éxito, 400 en validación, 410 si token expirado/usado
3. Registrar las 3 rutas en `internal/presentation/router/router.go`
4. Mapeo de errores → códigos HTTP:
   - `ErrInvalidEmail`, contraseña corta → 400
   - `ErrRateLimited` → 429
   - `ErrTokenNotFound` → 404
   - `ErrTokenExpired`, `ErrTokenAlreadyUsed` → 410
   - cualquier otro → 500

**Archivos a crear:**

- `proyecto/internal/presentation/dto/forgot_password_request.go`
- `proyecto/internal/presentation/dto/forgot_password_response.go`
- `proyecto/internal/presentation/dto/verify_reset_token_response.go`
- `proyecto/internal/presentation/dto/reset_password_request.go`
- `proyecto/internal/presentation/dto/reset_password_response.go`
- `proyecto/internal/presentation/handler/password_reset_handler.go`

**Archivos a modificar:**

- `proyecto/internal/presentation/router/router.go` — registrar `POST /api/auth/forgot-password`, `GET /api/auth/verify-reset-token/{token}`, `POST /api/auth/reset-password`

**Validation checkpoints:**

- [ ] `go build ./...` compila sin errores
- [ ] Cada handler tiene una prueba unitaria que verifica el código HTTP por escenario
- [ ] Verificar que el mensaje de respuesta de `ForgotPassword` es idéntico para correos existentes e inexistentes

**Dependencias:** Fase 4 y 5.

---

## Fase 7: Composition root e integración E2E (2h)

**Objetivo:** Cablear todas las dependencias en `main.go` y validar el flujo completo.

**Tareas:**

1. En `cmd/api/main.go`:
   - Construir `PostgresPasswordResetTokenRepo`
   - Construir `ConsoleEmailSender` (leer `APP_BASE_URL` de entorno)
   - Construir los 3 use cases
   - Construir `PasswordResetHandler` y pasarlo al router
2. Documentar la nueva variable `APP_BASE_URL` en el `Makefile` o `.env.example`
3. Pruebas manuales con `curl`:
   - `POST /api/auth/forgot-password` con correo registrado → revisar log de consola por el enlace
   - `GET /api/auth/verify-reset-token/<token>` → 200 `valid: true`
   - `POST /api/auth/reset-password` → 200 + login funciona con la nueva contraseña
   - Reusar el mismo token → 410
4. Actualizar `CLAUDE.md` con la sección "Implemented features → POST /api/auth/forgot-password, GET /verify-reset-token, POST /reset-password"

**Archivos a modificar:**

- `proyecto/cmd/api/main.go`
- `proyecto/Makefile` (opcional, agregar smoke tests `curl-forgot-password`)
- `CLAUDE.md`

**Validation checkpoints:**

- [ ] `make up` arranca sin errores
- [ ] El flujo E2E completo funciona vía `curl`
- [ ] El login con la contraseña anterior falla después del reinicio
- [ ] El login con la contraseña nueva funciona
- [ ] `make lint` no reporta nuevos warnings
- [ ] `make test && make test-integration` pasan en limpio

**Dependencias:** Fase 2 y 6.

---

## Cumplimiento de la Constitución

- **Clean Architecture:** cada fase respeta la dirección de dependencias (`Presentation → Application → Domain ← Infrastructure`)
- **Documentación:** cada función, struct e interfaz pública lleva doc comment en español
- **Errores tipados:** sentinel errors exportados (`ErrTokenExpired`, `ErrRateLimited`, etc.) — sin strings sueltos
- **Sin secrets en código:** `APP_BASE_URL` y futuras credenciales de SMTP vía variables de entorno
- **Validación en el servidor:** formato de correo, longitud de contraseña y validez del token siempre se validan en el use case
- **Tests antes de merge:** cada fase termina con `make test` (y `make test-integration` cuando aplique) en verde
- **Conventional Commits:** un commit por fase, formato `feat(password-reset): ...`
- **Boy Scout Rule:** si al modificar un archivo existente se detecta código duplicado o naming inconsistente, refactorizar en el mismo commit
