# Feature Specification: Password Reset Flow

**Status:** Draft
**Author:** Luis Ricardo Ruiz
**Created:** 2026-05-01

## 1. Executive Summary (2-3 líneas)

Servicio usado por el usuario final para la recuperación de contraseña de forma autónoma.
Este servicio debe reducir el contacto con servicio al cliente.

## 2. Problem Statement

### Current State (qué NO funciona hoy)

- Los usuarios que no pueden resetear sus contraseñas deben contactar a un administrador del Sistema
- El tiempo de atención depende de la disponibilidad del administrador del Sistema

### Desired State (cómo DEBERÍA funcionar)

- El usuario debe poder gestionar sus contraseñas
- El administrador se debe desligar de tareas de gestión de usuario

### Success Metrics (cómo sabrás que funciona)

- El administrador debe tener 0 tareas de reseteo de contraseñas
- Al menos el 50% de usuarios que inician el proceso de reseteo pueden terminarlo. Quiere decir que asocian una nueva contraseña a su usuario.

## 3. User Stories

### HU1: Solicitar una nueva contraseña

Como un usuario registrado que olvidó su contraseña
Quiero solicitar el reseteo de la contraseña por correo electrónico
Para que pueda recuperar el acceso a mi cuenta

#### Criterios de aceptación

- [ ] Dado que estoy en la página de login, cuando hago clic en el botón "Olvidé mi contraseña", entonces veo el formulario de reinicio de contraseña
- [ ] Dado que ingreso un correo electrónico válido, cuando hago clic en "Enviar", entonces recibo un correo electrónico con el enlace de reinicio en menos de 2 minutos
- [ ] Dado que ingreso un correo electrónico inválido o inexistente, cuando hago clic en "Enviar", entonces veo un mensaje genérico de éxito (sin enumerar si el correo existe)
- [ ] Dado que solicito el reinicio 3 o más veces en una hora, cuando intento una cuarta vez, entonces el sistema me bloquea temporalmente

### HU2: Restablecer contraseña

Como un usuario con un enlace de reinicio válido
Quiero establecer una nueva contraseña
Para que pueda ingresar a mi cuenta con las nuevas credenciales

#### Criterios de aceptación

- [ ] Dado que tengo un enlace válido (menor a 1 hora), cuando lo abro, entonces veo el formulario para ingresar la nueva contraseña
- [ ] Dado que el enlace expiró (mayor a 1 hora), cuando lo abro, entonces veo un error y la opción de solicitar uno nuevo
- [ ] Dado que establezco una nueva contraseña, cuando confirmo el cambio, entonces la contraseña anterior queda invalidada
- [ ] Dado que ya usé el enlace una vez, cuando intento usarlo de nuevo, entonces el sistema lo rechaza (uso único)

## 4. Requisitos Funcionales

### Obligatorio (Must have)

- Solicitar reinicio de contraseña vía correo electrónico
- Generar token seguro con expiración de 1 hora
- Enviar correo con enlace de reinicio al usuario
- Validar token antes de mostrar el formulario de nueva contraseña
- Invalidar el token después de usarlo (uso único)
- Actualizar la contraseña hasheada en la base de datos
- Mensaje genérico de respuesta que no enumere usuarios existentes

### Recomendado (Should have)

- Rate limiting: máximo 3 solicitudes por correo por hora
- Invalidar tokens anteriores al generar uno nuevo para el mismo usuario

### Opcional (Nice to have)

- Notificación al usuario cuando su contraseña fue cambiada exitosamente

### Fuera de alcance (Out of scope)

- Cambio de contraseña para usuarios autenticados (flujo diferente)
- Autenticación multifactor en el flujo de reinicio
- Soporte para reinicio vía SMS

## 5. Especificación Técnica

### Endpoints

```
POST /api/auth/forgot-password
Body:     { email: string }
Response: { message: "Si el correo existe, recibirás un enlace de reinicio" }
Rate limit: 3 solicitudes por hora por correo

GET /api/auth/verify-reset-token/:token
Response: { valid: boolean, expiresAt?: string }

POST /api/auth/reset-password
Body:     { token: string, newPassword: string }
Response: { message: "Contraseña actualizada" }
```

### Base de datos

```sql
-- migrations/003_add_password_reset_tokens.sql
CREATE TABLE password_reset_tokens (
    id         SERIAL PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_password_reset_tokens_token   ON password_reset_tokens(token);
CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
```

### Modelo de dominio

```go
// internal/domain/entity/password_reset_token.go
type PasswordResetToken struct {
    ID        int
    UserID    int
    Token     string
    ExpiresAt time.Time
    Used      bool
    CreatedAt time.Time
}
```

### Flujo de datos

```
Usuario → Formulario "Olvidé mi contraseña"
  ↓
POST /forgot-password
  ↓
Generar token (32 bytes aleatorio seguro)
  ↓
Guardar en DB (expira en 1 hora)
  ↓
Enviar correo con enlace
  ↓
Usuario hace clic en el enlace
  ↓
GET /verify-reset-token/:token
  ↓
Mostrar formulario de nueva contraseña
  ↓
POST /reset-password
  ↓
Hashear nueva contraseña
  ↓
Actualizar users.password_hash
  ↓
Marcar token como usado
  ↓
Éxito
```

### Seguridad

- Token generado con `crypto/rand` (32 bytes, hexadecimal)
- Expiración: 1 hora desde la creación
- Uso único: el token se marca como `used = true` tras el reinicio exitoso
- Rate limiting: 3 intentos por correo por hora
- Sin enumeración de usuarios: respuesta genérica independiente de si el correo existe
- Requisitos de contraseña: mínimo 8 caracteres

## 6. Estrategia de Pruebas

### Pruebas unitarias (sin DB)

| Caso | Descripción |
|------|-------------|
| Token válido | `VerifyResetToken` retorna `valid: true` para token no expirado y no usado |
| Token expirado | `VerifyResetToken` retorna error cuando `expires_at < now` |
| Token ya usado | `VerifyResetToken` retorna error cuando `used = true` |
| Reinicio exitoso | `ResetPassword` actualiza el hash y marca el token como usado |
| Correo inexistente | `RequestPasswordReset` retorna éxito sin enviar correo (sin enumerar) |
| Rate limit excedido | `RequestPasswordReset` retorna error después de 3 intentos en 1 hora |
| Contraseña inválida | `ResetPassword` retorna error si la nueva contraseña no cumple los requisitos |

### Pruebas de integración (con DB)

- `POST /forgot-password` con correo válido: crea token en DB y responde 200
- `GET /verify-reset-token/:token` con token válido: responde `valid: true`
- `GET /verify-reset-token/:token` con token expirado: responde `valid: false`
- `POST /reset-password` con token válido: actualiza contraseña y marca token como usado
- `POST /reset-password` usando el mismo token dos veces: segundo intento falla

### Lo que NO se testea

- El envío real del correo (se mockea el servicio de email)
- La librería de generación de tokens
