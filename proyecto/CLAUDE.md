# Password Reset Implementation

## Context Documents

1. **Spec:** `docs/specs/spec.md` — qué construir y por qué
2. **Plan:** `docs/plans/plan.md` — fases de implementación
3. **Tasks:** `docs/tasks/tasks.md` — tareas ejecutables con acceptance criteria
4. **Constitution:** `docs/constitution.md` — principios de arquitectura y código

## Agentic Workflow Rules

### Task Execution Protocol

For each task in `docs/tasks/tasks.md`:

1. Read task details (archivos, acceptance criteria, comando de validación)
2. Implement ONLY that task — sin features extra ni refactors no pedidos
3. Run validation commands listed in the task
4. If validation fails → Fix → Re-validate (no avanzar hasta que pase)
5. If validation passes → Commit → Mark task complete
6. Move to next task

### Validation Commands (run after EACH task)

```bash
cd proyecto
go build ./...        # must compile clean
go vet ./...          # no static analysis issues
make test             # unit tests pass
make test-integration # integration tests pass (when DB is running)
make lint             # no new lint warnings
```

### Commit Protocol

```bash
git add <only task files>
git commit -m "feat(password-reset): <task description>

Implements <TASK_ID> from docs/tasks/tasks.md"
```

### Blocker Protocol

If blocked (requisito ambiguo, dependencia faltante, decisión de arquitectura):

1. STOP — no proceder
2. Documentar el bloqueo en `docs/tasks/tasks.md` bajo la tarea afectada
3. Listar las posibles soluciones
4. Esperar decisión humana

### Deviation Protocol

If the spec needs a change (mejor enfoque encontrado, spec incompleto):

1. PAUSE — no implementar la desviación
2. Explicar el problema con el spec
3. Sugerir el cambio al spec
4. Esperar aprobación
5. Actualizar spec primero
6. Luego implementar

## Current Status

**Feature:** Password Reset Flow
**Completed:** 6/17 tasks

## State

### Fase 1 — Capa de Dominio + Migración
- [x] T01 — Entidad `PasswordResetToken`
- [x] T02 — Interfaz `PasswordResetTokenRepository` + sentinel errors
- [x] T03 — Mock `PasswordResetTokenRepository`
- [x] T04 — Migración SQL `004_add_password_reset_tokens.sql` *(pendiente validar con `make down-clean && make up`)*

### Fase 2 — Infraestructura Postgres
- [x] T05 — `PostgresPasswordResetTokenRepo` — implementación
- [x] T06 — `PostgresPasswordResetTokenRepo` — integration tests

### Fase 3 — Servicio de correo
- [ ] T07 — Interfaz `EmailSender` + mock
- [ ] T08 — `ConsoleEmailSender`

### Fase 4 — Use case `RequestPasswordReset`
- [ ] T09 — `UpdatePassword` en dominio + Postgres
- [ ] T10 — Caso de uso `RequestPasswordReset`
- [ ] T11 — Tests unitarios `RequestPasswordReset`

### Fase 5 — Use cases `VerifyResetToken` + `ResetPassword`
- [ ] T12 — Caso de uso `VerifyResetToken` + tests
- [ ] T13 — Caso de uso `ResetPassword` + tests

### Fase 6 — Capa de Presentación
- [ ] T14 — DTOs de password reset
- [ ] T15 — `PasswordResetHandler`
- [ ] T16 — Registrar rutas en router

### Fase 7 — Composition root e integración E2E
- [ ] T17 — Composition root + smoke tests E2E
