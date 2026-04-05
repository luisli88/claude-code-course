# Plan de Trabajo Estratégico: Claude Code SDD Mastery

**Tu perfil:** Usuario avanzado de Claude Code, experiencia con MCPs, actualmente usando Vibe Coding, objetivo de optimización de tokens y transición a Spec-Driven Development.

---

## 🎯 Día 1: Auditoría y Establecimiento de Baseline

### Objetivo

Cuantificar tu uso actual de tokens y establecer métricas base para medir mejoras posteriores.

### Estrategia

Análisis antes de optimización - no puedes mejorar lo que no mides.

### Justificación

Sin baseline cuantitativo, no sabrás si tus optimizaciones funcionan. Necesitas datos concretos de consumo actual vs optimizado.

### Qué Harás

#### 1.1 Análisis de Contexto Actual (90 min)

**Acción concreta:**

```bash
# Abre un proyecto típico tuyo
cd ~/tu-proyecto-actual

# Inicia Claude y observa
claude
/context
```

**Qué observar:**

- Total de tokens en contexto
- Archivos más grandes consumiendo tokens
- MCPs activos (cuáles, cuántos tokens usan)
- Longitud de conversación

**Ejercicio de comprensión:**
Responde estas preguntas:

- ¿Cuántos tokens consume tu sesión típica?
- ¿Qué porcentaje son archivos vs conversación?
- ¿Cuáles archivos están en contexto que NO necesitas?

**Validación de entendimiento:**
Deberías poder decir: "Mi proyecto típico consume X tokens, donde Y% son archivos innecesarios que puedo ignorar."

#### 1.2 Crear .claudeignore Quirúrgico (60 min)

**Por qué esto es crítico:**
Cada archivo que Claude lee = tokens consumidos. Si lees `vendor/` con miles de archivos, estás quemando tokens en dependencias que nunca modificarás.

**Acción:**

```bash
cd ~/tu-proyecto-actual

# Crea archivo
cat > .claudeignore << 'EOF'
# === BUILDS & ARTIFACTS ===
# Justificación: Archivos generados, no fuente
vendor/
bin/
dist/
tmp/

# === BINARIOS COMPILADOS ===
# Justificación: Claude no puede leer binarios
*.exe
*.so
*.dylib
*.test

# === ASSETS BINARIOS ===
# Justificación: Claude no puede leer imágenes/videos
*.jpg
*.png
*.gif
*.mp4
*.mp3

# === LOGS Y TEMPORALES ===
*.log
.DS_Store
*.tmp
coverage.out
coverage.html

# === CÓDIGO GENERADO ===
# Justificación: Generado automáticamente por tooling
*.pb.go
*_mock.go
mock_*.go
EOF
```

**Ejercicio de comprensión:**
Antes de aplicar, lista 3 tipos de archivos en TU proyecto que Claude no debería leer.

**Validación:**

```bash
# Después de crear .claudeignore
claude
/context

# Los tokens deberían reducirse 30-50%
# Si no, revisa qué archivos grandes siguen en contexto
```

**Pregunta de validación:**
"Si mi proyecto tenía 80k tokens de contexto y ahora tiene 45k, ¿por qué bajó tanto?"
_Respuesta esperada:_ "Porque excluí node_modules (20k), assets binarios (10k), y archivos generados (5k)."

#### 1.3 Estrategia de Model Selection (60 min)

**El concepto clave:**
Claude tiene 3 modelos con distintos costos:

- **Haiku 3.5:** Rápido, barato, tareas simples (texto, refactors mecánicos)
- **Sonnet 3.7:** Balanceado, 80% de casos (implementación estándar)
- **Opus 4:** Poderoso, caro, decisiones críticas (arquitectura, debugging complejo)

**Por qué esto importa:**
Opus cuesta 5x más que Sonnet. Si usas Opus para todo, gastas 5x innecesariamente.

**Tu tabla de decisión:**

| Tarea                | Modelo     | Comando             | Por Qué                                          |
| -------------------- | ---------- | ------------------- | ------------------------------------------------ |
| Arquitectura inicial | Opus 4     | `/model opus-4`     | Decisiones críticas, necesito mejor razonamiento |
| Implementar CRUD     | Sonnet 3.7 | `/model sonnet-3.7` | Patrón conocido, no necesito Opus                |
| Escribir docs        | Haiku 3.5  | `/model haiku-3.5`  | Texto simple, Haiku suficiente                   |
| Bug complejo         | Opus 4     | `/model opus-4`     | Debugging requiere razonamiento profundo         |
| Refactor mecánico    | Sonnet 3.7 | `/model sonnet-3.7` | Cambios predecibles                              |

**Ejercicio práctico:**
Implementa un endpoint simple alternando modelos:

```bash
claude
/model opus-4

"Diseña arquitectura para endpoint POST /api/users
- Estructura de carpetas
- Separación de capas
- Patrones a seguir"

# Claude da diseño de alto nivel (Opus para esto)

"Ok, ahora implementa"
/model sonnet-3.7

# Claude implementa con Sonnet (más barato)
```

**Validación de entendimiento:**
¿Cuándo usarías cada modelo? Escribe 2 ejemplos por modelo de tu dominio.

#### 1.4 Session Hygiene - Task Chunking (60 min)

**El problema del Vibe Coding:**

```
Inicio sesión → Feature 1 → Feature 2 → Bug fix → Refactor → ...
                ↓
         Contexto crece y crece
                ↓
         Session de 200k tokens
                ↓
         $$$ desperdiciados
```

**La solución SDD:**

```
Feature 1 → /clear → Feature 2 → /clear → Bug fix
   ↓                    ↓                     ↓
30k tokens          25k tokens            15k tokens
   ↓                    ↓                     ↓
git commit          git commit           git commit
```

**Regla de oro:**
**1 feature = 1 session = 1 commit**

**Práctica:**

```bash
claude
/clear  # Limpia contexto anterior

"Implementa SOLO función de login:
- POST /api/auth/login
- Validación email/password
- Return JWT
- NADA MÁS"

# Implementa
# Valida
git commit -m "feat(auth): implement login endpoint"

exit  # Cierra Claude

# Nueva feature = nueva sesión
claude
/clear

"Ahora implementa SOLO función de registro..."
```

**Por qué funciona:**
Cada sesión empieza limpia. No acumulas contexto de features anteriores.

**Validación:**
Compara:

- **Antes:** Sesión de 4 features = 200k tokens
- **Después:** 4 sesiones de 1 feature = 4 × 40k = 160k tokens (20% ahorro)

**Pregunta de comprensión:**
"¿Por qué `/clear` entre features ahorra tokens?"
_Respuesta esperada:_ "Porque borra el historial de conversación de la feature anterior que ya no necesito."

---

## 🎯 Día 2: Skills System - Automatizar Conocimiento Repetitivo

### Objetivo

Convertir tus patrones de código recurrentes en Skills reutilizables que reduzcan tokens y aumenten consistencia.

### Estrategia

Identificar qué le explicas repetidamente a Claude → Encapsularlo en Skill → Invocar con 1 comando.

### Justificación

Si cada vez que creas un API endpoint explicas "usa esta estructura de 3 capas, con estos imports, esta validación...", estás gastando 500 tokens cada vez. Un Skill lo reduce a 50 tokens (90% ahorro).

### Qué Harás

#### 2.1 Pattern Mining - Identificar Repeticiones (60 min)

**Concepto:**
Tienes patrones que repites en cada proyecto. Identificarlos es el primer paso.

**Acción:**
Revisa tus últimos 5 commits/PRs y responde:

```markdown
## Patterns que SIEMPRE uso:

1. **Backend API Endpoint:**
   - stdlib `net/http` + Clean Architecture
   - ¿Estructura: handler → use case → repository?
   - ¿Validación manual en use case o handler?
   - ¿Tests con `testing` stdlib + `testify`?

2. **Domain Entity:**
   - ¿Struct con campos exportados?
   - ¿Validación en el constructor?
   - ¿Errores sentinel exportados?

3. **Database Access:**
   - ¿`database/sql` directo? ¿`sqlx`? ¿`pgx`?
   - ¿Migrations manuales en SQL?
   - ¿Naming: snake_case en DB, PascalCase en Go?

4. **Git Workflow:**
   - ¿Conventional commits?
   - ¿Branch naming: feature/xxx?
   - ¿Formato de mensaje específico?
```

**Ejercicio de comprensión:**
Lista 3 cosas que le explicas a Claude en CADA proyecto.

**Validación:**
Deberías tener 3-5 patterns identificados con detalles concretos.

#### 2.2 Crear Primer Skill: API Route Generator (90 min)

**Por qué este primero:**
Es el patrón más común en backend development. Cada endpoint sigue la misma estructura.

**Entendiendo la anatomía de un Skill:**

```
~/.claude/skills/api-route/
├── SKILL.md              ← Instrucciones principales
├── references/           ← Docs de referencia (opcional)
│   └── examples.md
└── templates/            ← Templates de código (opcional)
    └── endpoint.ts
```

**Acción:**

```bash
mkdir -p ~/.claude/skills/api-route
```

Crea `~/.claude/skills/api-route/SKILL.md`:

````markdown
---
name: api-route
description: Genera API endpoint con Clean Architecture en Go
disable-model-invocation: false
---

# API Route Generator Skill

## Cuándo usar este skill

Cuando necesites crear un nuevo endpoint REST en Go con Clean Architecture.

## Arquitectura que SIEMPRE uso

```
internal/
├── domain/
│   ├── entity/          ← Entidades (structs puros, sin deps)
│   └── repository/      ← Interfaces de repositorio
├── application/
│   └── usecase/         ← Lógica de negocio (depende solo de domain)
├── infrastructure/
│   └── persistence/     ← Implementaciones concretas de repositorio
└── presentation/
    ├── dto/             ← Request/Response structs
    ├── handler/         ← HTTP handlers
    └── router/          ← Registro de rutas
```

### Capa 1: Domain Entity (internal/domain/entity/\*.go)

**Responsabilidad:** Struct puro sin dependencias externas

```go
package entity

import "time"

type Product struct {
    ID        int
    Name      string
    Price     float64
    Category  string
    CreatedAt time.Time
}
```

### Capa 2: Repository Interface (internal/domain/repository/\*.go)

**Responsabilidad:** Contrato de acceso a datos (solo interfaces)

```go
package repository

import "myapp/internal/domain/entity"

var ErrNotFound = errors.New("not found")

type ProductRepository interface {
    Create(p entity.Product) (*entity.Product, error)
    FindByID(id int) (*entity.Product, error)
}
```

### Capa 3: Use Case (internal/application/usecase/\*.go)

**Responsabilidad:** Lógica de negocio, validación, orquestación

```go
package usecase

import (
    "errors"
    "myapp/internal/domain/entity"
    "myapp/internal/domain/repository"
)

var (
    ErrInvalidInput = errors.New("invalid input")
    ErrNotFound     = errors.New("product not found")
)

type CreateProduct struct {
    repo repository.ProductRepository
}

func NewCreateProduct(repo repository.ProductRepository) *CreateProduct {
    return &CreateProduct{repo: repo}
}

func (uc *CreateProduct) Execute(name string, price float64, category string) (*entity.Product, error) {
    if name == "" {
        return nil, ErrInvalidInput
    }
    if price < 0 {
        return nil, ErrInvalidInput
    }
    p := entity.Product{Name: name, Price: price, Category: category}
    return uc.repo.Create(p)
}
```

### Capa 4: DTO (internal/presentation/dto/\*.go)

**Responsabilidad:** Estructuras de serialización HTTP

```go
package dto

type CreateProductRequest struct {
    Name     string  `json:"name"`
    Price    float64 `json:"price"`
    Category string  `json:"category"`
}

type CreateProductResponse struct {
    ID        int     `json:"id"`
    Name      string  `json:"name"`
    Price     float64 `json:"price"`
    Category  string  `json:"category"`
    CreatedAt string  `json:"created_at"`
}
```

### Capa 5: Handler (internal/presentation/handler/\*.go)

**Responsabilidad:** Manejar request/response HTTP, NO lógica de negocio

```go
package handler

import (
    "encoding/json"
    "net/http"
    "myapp/internal/application/usecase"
    "myapp/internal/presentation/dto"
)

type ProductHandler struct {
    createProduct *usecase.CreateProduct
}

func NewProductHandler(uc *usecase.CreateProduct) *ProductHandler {
    return &ProductHandler{createProduct: uc}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req dto.CreateProductRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }

    product, err := h.createProduct.Execute(req.Name, req.Price, req.Category)
    if err != nil {
        if errors.Is(err, usecase.ErrInvalidInput) {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(dto.CreateProductResponse{
        ID:       product.ID,
        Name:     product.Name,
        Price:    product.Price,
        Category: product.Category,
    })
}
```

### Tests (internal/application/usecase/\*_test.go)

**Responsabilidad:** Validar comportamiento con mock del repositorio

```go
package usecase_test

import (
    "testing"
    "myapp/internal/application/usecase"
    "myapp/internal/domain/repository"
)

func TestCreateProduct_Success(t *testing.T) {
    repo := repository.NewMockProductRepository()
    uc := usecase.NewCreateProduct(repo)

    product, err := uc.Execute("Widget", 9.99, "tools")

    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if product.Name != "Widget" {
        t.Errorf("expected name Widget, got %s", product.Name)
    }
}

func TestCreateProduct_InvalidInput(t *testing.T) {
    repo := repository.NewMockProductRepository()
    uc := usecase.NewCreateProduct(repo)

    _, err := uc.Execute("", 9.99, "tools")

    if !errors.Is(err, usecase.ErrInvalidInput) {
        t.Errorf("expected ErrInvalidInput, got %v", err)
    }
}
```

## Checklist de implementación

Cuando crees un endpoint, DEBES:

- [ ] Entity en `domain/entity/`
- [ ] Interface en `domain/repository/`
- [ ] Use case en `application/usecase/`
- [ ] DTO request/response en `presentation/dto/`
- [ ] Handler en `presentation/handler/`
- [ ] Ruta registrada en `presentation/router/`
- [ ] Use case wired en `cmd/api/main.go`
- [ ] Mínimo 2 tests de use case (happy + error path)

## Errores comunes a EVITAR

❌ Lógica de negocio en handler
❌ Acceso directo a DB desde handler o use case
❌ Importar `infrastructure/` desde `application/` o `domain/`
❌ `interface{}` sin justificación
❌ Commits sin tests
````

**Validación de comprensión:**

**Pregunta 1:** "¿Por qué separamos en estas capas?"
_Respuesta esperada:_ "Clean Architecture con dependency rule estricta: Presentation → Application → Domain ← Infrastructure. Domain y Application no saben nada de HTTP ni de la DB concreta, son 100% testeables sin levantar el servidor."

**Pregunta 2:** "¿Por qué validar en use case y no en handler?"
_Respuesta esperada:_ "El use case es la lógica de negocio, y las reglas de negocio incluyen validación. Si valido en el handler, la validación queda acoplada a HTTP y no puedo reutilizarla desde CLI, jobs, etc."

**Pregunta 3:** "¿Qué pasa si el handler importa el repositorio directamente?"
_Respuesta esperada:_ "Viola la dependency rule: Presentation no debe depender de Infrastructure. El handler se vuelve imposible de testear sin una DB real, y cambiar la DB rompe el handler."

#### 2.3 Usar el Skill (30 min)

**Acción:**

```bash
cd ~/test-project
claude

/api-route

"Crea endpoint POST /api/products para crear producto con:
- name (string, required)
- price (float, min 0)
- category (string, uno de: electronics, clothing, food)"
```

**Qué debería pasar:**
Claude generará automáticamente:

1. `internal/domain/entity/product.go`
2. `internal/domain/repository/product_repository.go`
3. `internal/application/usecase/create_product.go`
4. `internal/application/usecase/create_product_test.go`
5. `internal/presentation/dto/create_product_request.go`
6. `internal/presentation/handler/product_handler.go`

Todo siguiendo tu estructura de Clean Architecture.

**Validación:**
Verifica que los archivos generados:

- [ ] Siguen el flujo domain → usecase → handler
- [ ] No importan `infrastructure/` desde `application/` o `domain/`
- [ ] Validación en use case, no en handler
- [ ] Tests de use case con mock de repositorio
- [ ] Errores sentinel exportados

**Comparación:**

| Métrica                      | Sin Skill | Con Skill  |
| ---------------------------- | --------- | ---------- |
| Tokens explicando estructura | ~800      | ~50        |
| Tiempo                       | 5 min     | 30 seg     |
| Consistencia                 | Variable  | 100%       |
| Ahorro                       | -         | 95% tokens |

#### 2.4 Crear Segundo Skill: Commit Message Generator (60 min)

**Por qué este:**
Generar buenos commit messages consume tiempo y tokens. Automatízalo.

```bash
mkdir -p ~/.claude/skills/commit-pro
```

`SKILL.md`:

````markdown
---
name: commit
description: Genera commit message convencional desde git diff
disable-model-invocation: true
---

# Commit Message Pro

## Proceso

1. Ejecuto: `git diff --cached`
2. Analizo cambios
3. Determino type + scope
4. Genero mensaje formato Conventional Commits

## Types (Conventional Commits)

- **feat:** Nueva feature
- **fix:** Bug fix
- **refactor:** Refactorización (sin cambio funcional)
- **perf:** Mejora de performance
- **test:** Agregar/modificar tests
- **docs:** Documentación
- **style:** Formato (linting, espacios, etc)
- **chore:** Mantenimiento (deps, config)
- **ci:** CI/CD changes

## Scope Inference

Basado en path de archivos modificados:

- `internal/domain/entity/user*` → scope: `user`
- `internal/application/usecase/auth*` → scope: `auth`
- `internal/infrastructure/persistence/*` → scope: `db`
- `internal/presentation/handler/*` → scope: `handler`
- `migrations/*` → scope: `migration`
- `cmd/api/*` → scope: `main`

## Formato Output

```

<type>(<scope>): <description en presente, lowercase, max 50 chars>

[Body opcional explicando POR QUÉ si es complejo]

[Footer: Breaking changes o issues]

```

## Reglas

- Description: presente, lowercase, sin punto final
- Max 50 chars primera línea (GitHub trunca después)
- Body: wrap a 72 chars
- Breaking changes: `BREAKING CHANGE:` en footer

## Ejecución

Después de generar, ejecuto:

```bash
git commit -m "mensaje generado"
```

## Ejemplos

### Example 1: Feature simple

```
git diff --cached:
+ internal/application/usecase/login.go: func Execute()

Output:
feat(auth): implement login use case
```

### Example 2: Bug fix con context

```
git diff --cached:
- internal/infrastructure/persistence/postgres_user_repo.go: missing ErrNotFound mapping

Output:
fix(db): map sql.ErrNoRows to repository.ErrNotFound

FindByEmail was returning raw sql.ErrNoRows instead of
the sentinel error, breaking callers that used errors.Is().
```

### Example 3: Breaking change

```
git diff --cached:
- internal/domain/repository/user_repository.go: FindByID signature changed

Output:
feat(domain): add context to UserRepository interface

BREAKING CHANGE: All repository methods now require context.Context.
Implementations must be updated to pass ctx to sql calls.
```
````

**Usar el skill:**

```bash
# Hacer cambios
git add src/auth/login.ts

claude
/commit

# Claude analiza diff y genera commit message
# Claude ejecuta git commit con el mensaje
```

**Validación de comprensión:**

**Pregunta:** "¿Por qué `disable-model-invocation: true` en este skill?"
_Respuesta esperada:_ "Porque es tarea mecánica (parse diff, aplicar reglas), no necesita invocación extra del modelo. Ahorra tokens."

**Pregunta:** "¿Cuándo usarías body en commit message?"
_Respuesta esperada:_ "Cuando el PORQUÉ del cambio no es obvio desde el código. Por ejemplo, decisión de arquitectura, bug complejo que requiere contexto."

#### 2.5 Crear Tercer Skill: Code Review (60 min)

**Por qué:**
Antes de commit, quieres que Claude revise tu código.

```bash
mkdir -p ~/.claude/skills/code-review
```

`SKILL.md`:

````markdown
---
name: review
description: Code review exhaustivo con checklist
disable-model-invocation: false
---

# Code Review Skill

## Objetivo

Revisar código staged antes de commit para detectar issues.

## Checklist de Revisión

### 🐛 Bugs Potenciales

- [ ] Errores ignorados (err asignado pero no chequeado)
- [ ] Nil pointer dereference
- [ ] Off-by-one en slices
- [ ] Race conditions en goroutines (vars compartidas sin mutex)
- [ ] Goroutine leaks (goroutines que no terminan)
- [ ] Context cancelation no propagada

### 🔒 Security Issues

- [ ] Input validation (SQL injection, path traversal)
- [ ] Autenticación en endpoints sensibles
- [ ] Secrets hardcodeados (tokens, passwords en código)
- [ ] CORS configurado correctamente
- [ ] Rate limiting en endpoints públicos
- [ ] Datos sensibles en logs

### ⚡ Performance

- [ ] Loops O(n²) o peor
- [ ] Queries N+1 (DB)
- [ ] Allocations innecesarias en hot path
- [ ] DB connections no cerradas (defer rows.Close())
- [ ] Large payloads sin pagination

### 🎨 Code Quality

- [ ] Nombres descriptivos (Go idiomático: `err` no `error`, `r` para reader)
- [ ] Funciones < 50 líneas
- [ ] DRY violations
- [ ] Magic numbers (usar constantes)
- [ ] Comentarios solo donde necesario (exported types/funcs documentadas)
- [ ] No `interface{}` sin justificación
- [ ] Dependency rule respetada (no imports cross-layer incorrectos)

### ✅ Testing

- [ ] Happy path covered
- [ ] Error cases covered (cada rama de error)
- [ ] Edge cases (empty string, zero value, nil)
- [ ] Mocks apropiados (sin tocar DB real en unit tests)
- [ ] Subtests con `t.Run` para legibilidad

### 📚 Documentation

- [ ] Comentarios godoc en funciones/types exportados
- [ ] README actualizado si nueva feature
- [ ] Errores sentinel documentados

## Proceso de Review

1. Leo archivos staged: `git diff --cached`
2. Reviso contra checklist
3. Reporto findings por severidad
4. Sugiero fixes específicos

## Output Format

```markdown
## 🔍 Code Review Results

### ❌ Critical Issues (must fix)

- **File:** `src/auth/login.ts:45`
  **Issue:** SQL injection vulnerability
  **Fix:** Use parameterized query: `prisma.user.findFirst({ where: { email } })`

### ⚠️ Warnings (should fix)

- **File:** `src/services/user.ts:120`
  **Issue:** Function too long (85 lines)
  **Fix:** Extract validation logic to separate function

### 💡 Suggestions (nice to have)

- **File:** `src/utils/format.ts:10`
  **Issue:** Magic number `3600`
  **Fix:** Create constant `SECONDS_IN_HOUR = 3600`

### ✅ Strengths

- Excellent error handling in controllers
- Good test coverage (87%)
- Clear naming conventions

## Recommendation

🔴 DO NOT MERGE - Fix critical issues first
🟡 MERGE WITH CAUTION - Address warnings in follow-up
🟢 APPROVED - Ready to merge
```
````

**Usar:**

```bash
# Después de implementar feature
git add .

claude
/review

# Claude revisa y genera reporte
# Corriges issues
# Re-ejecutas /review hasta verde
```

**Validación de comprensión:**

**Ejercicio:** Crea código intencionalmente malo y pásalo por `/review`:

```go
// internal/infrastructure/persistence/bad_example.go
func GetUser(db *sql.DB, id string) *User {
    row := db.QueryRow("SELECT * FROM users WHERE id = " + id)
    var u User
    row.Scan(&u.ID, &u.Name)
    return &u
}
```

**Pregunta:** "¿Qué debería detectar el skill?"
_Respuesta esperada:_

- ❌ Critical: SQL injection (concatenación directa en query)
- ⚠️ Warning: Error de `row.Scan` ignorado
- ⚠️ Warning: Error de `db.QueryRow` no chequeado
- ⚠️ Warning: Retorna `*User` nil si no encuentra (nil pointer en caller)
- 💡 Suggestion: Usar `$1` placeholder + parámetro separado

---

## 🎯 Día 3: Constitution - Tu Sistema de Principios

### Objetivo

Documentar TUS principios de desarrollo en un documento único que Claude siempre consulta, eliminando la necesidad de re-explicar standards.

### Estrategia

"Write once, reference forever" - Define tus reglas una vez, Claude las aplica siempre.

### Justificación

Cada vez que empiezas proyecto nuevo y explicas "uso TypeScript strict, tests con Jest, naming camelCase, etc" gastas 1000+ tokens. Constitution lo reduce a "sigue constitution.md" (20 tokens, 98% ahorro).

### Qué Harás

#### 3.1 Entender la Jerarquía de Configuración (30 min)

**Concepto clave:**
Tienes 3 niveles de configuración con distinta precedencia:

```
┌─────────────────────────────────┐
│  ~/.claude.md (GLOBAL)          │  ← Tus preferencias personales
│  "En TODO proyecto yo..."       │     (Neovim, Arch Linux, español)
└─────────────────────────────────┘
            ↓ overrides
┌─────────────────────────────────┐
│  ./constitution.md (UNIVERSAL)  │  ← Principios técnicos universales
│  "Todo proyecto debe..."        │     (Testing, TypeScript, patterns)
└─────────────────────────────────┘
            ↓ overrides
┌─────────────────────────────────┐
│  ./CLAUDE.md (PROYECTO)         │  ← Contexto específico proyecto
│  "ESTE proyecto usa..."         │     (Stack, APIs, constraints)
└─────────────────────────────────┘
```

**Por qué esta jerarquía:**

- **Global:** No repetir "uso Neovim" en cada proyecto
- **Constitution:** No repetir "tests obligatorios" en cada proyecto
- **Project:** Contexto único de ESTE proyecto

**Analogía:**

```
Global = Tu personalidad
Constitution = Leyes de tu país
Project = Reglas de tu casa
```

**Ejercicio de comprensión:**

Clasifica dónde va cada regla:

1. "Prefiero comentarios en español"
2. "Tests coverage mínimo 80%"
3. "Este proyecto usa PostgreSQL 14"
4. "Uso Arch Linux con Neovim"
5. "TypeScript strict mode obligatorio"
6. "Esta API autentica con JWT"

_Respuestas:_

- Global: 1, 4
- Constitution: 2, 5
- Project: 3, 6

#### 3.2 Crear Constitution Personal (120 min)

**Estructura de Constitution:**

```markdown
# Development Constitution

**Author:** [Tu nombre]
**Version:** 1.0
**Last Updated:** [Fecha]

## 1. Philosophy (Por qué desarrollas así)

## 2. Tech Stack Standards (Qué tecnologías usas)

## 3. Architecture Patterns (Cómo estructuras código)

## 4. Code Quality Standards (Qué calidad esperas)

## 5. Testing Standards (Cómo testeas)

## 6. Security Standards (Cómo proteges)

## 7. Non-Negotiables (Qué NUNCA se rompe)
```

**Acción:** Crea `~/dev-constitution/constitution.md`

Voy a guiarte sección por sección:

**Sección 1: Philosophy**

```markdown
## Philosophy

### Core Values

¿Qué valoras MÁS en código?

Ejemplos:

- Simplicidad > Cleverness
- Explicit > Implicit
- Type Safety > Flexibilidad
- Tests > Documentación
- Performance medida > Performance asumida

### Development Principles

¿Qué principios sigues?

Ejemplos:

- YAGNI (You Aren't Gonna Need It)
- DRY (Don't Repeat Yourself)
- KISS (Keep It Simple)
- Separation of Concerns
- Fail Fast

**Tu turno:** Escribe tus 3-5 valores core y 3-5 principios.
```

**Pregunta de validación:**
"¿Por qué 'Simplicidad > Cleverness'?"
_Respuesta esperada:_ "Código clever es difícil de entender y mantener. Prefiero código obvio que cualquiera pueda modificar."

**Sección 2: Tech Stack Standards**

```markdown
## Tech Stack Standards

### Backend

**Required:**

- Go (stdlib `net/http` + `database/sql`, sin frameworks)
- PostgreSQL con driver `github.com/lib/pq` o `pgx`
- Validación manual en use cases (sin librerías de validación)
- JWT con `github.com/golang-jwt/jwt/v5`
- Passwords con `golang.org/x/crypto/bcrypt`

**Prohibited:**

- Frameworks HTTP (Gin, Echo, Fiber) — stdlib es suficiente
- ORMs (GORM) — SQL directo es más legible y controlable
- `interface{}` sin justificación clara

### Herramientas

**Required:**

- `golangci-lint` para linting
- `go test ./...` para tests (con `-race` para detectar race conditions)
- Docker + `docker-compose` para DB local
- SQL migrations en archivos `.sql` versionados

**Tu turno:** Lista TU stack. Sé específico con versiones si importa.
```

**Ejercicio:** Responde estas preguntas antes de escribir:

1. ¿Qué backend framework uso en 90% de proyectos?
2. ¿Qué ORM/database library prefiero?
3. ¿Qué NUNCA usaría? ¿Por qué?

**Sección 3: Architecture Patterns**

````markdown
## Architecture Patterns

### Backend Structure — Clean Architecture

```
cmd/api/
└── main.go              # Composition root: wiring de todas las capas

internal/
├── domain/
│   ├── entity/          # Structs puros, sin deps externas
│   └── repository/      # Interfaces de repositorio + errores sentinel
├── application/
│   └── usecase/         # Lógica de negocio (depende solo de domain)
├── infrastructure/
│   └── persistence/     # Implementaciones concretas (Postgres, etc.)
└── presentation/
    ├── dto/             # Request/Response structs JSON
    ├── handler/         # HTTP handlers (llaman use cases)
    └── router/          # Registro de rutas

migrations/              # Archivos .sql versionados
```

**Dependency rule:** Presentation → Application → Domain ← Infrastructure
**Tu turno:** Dibuja TU estructura. ¿Usas este layout? ¿Hay capas extra?
````

**Validación de comprensión:**

Si alguien ve tu estructura, debería poder responder:

- ¿Dónde va la validación de input?
- ¿Dónde va la lógica de negocio?
- ¿Dónde van las queries de DB?

Si no puede, tu estructura no es clara. Refina.

**Sección 4: Code Quality Standards**

````markdown
## Code Quality Standards

### Naming Conventions

- Variables/funciones unexported: `camelCase`
- Types/funcs exported: `PascalCase`
- Constantes: `PascalCase` si exported, `camelCase` si no
- Archivos: `snake_case.go` (ej: `user_repository.go`)
- Errores sentinel: `ErrXxx` (ej: `ErrNotFound`)
- Interfaces de 1 método: nombre del método + `-er` (ej: `Reader`, `UserCreator`)

### Function Guidelines

- Max líneas por función: 50
- Max parámetros: 4 (si más, agrupar en struct)
- Retornar siempre `(value, error)` en funciones que pueden fallar
- Preferir funciones pequeñas y enfocadas

### Error Handling

Go no tiene excepciones. El patrón estándar:

```go
// ✅ Siempre verificar errores
result, err := operation()
if err != nil {
    return nil, fmt.Errorf("operation failed: %w", err)
}

// ✅ Errores sentinel para que el caller pueda hacer errors.Is()
var ErrNotFound = errors.New("not found")

// ✅ Wrapping con contexto
return nil, fmt.Errorf("findUser(%d): %w", id, ErrNotFound)
```
````

**Tu turno:** Define TUS standards. Incluye ejemplos de código.

**Ejercicio práctico:**

Escribe una función usando TUS naming conventions:

```go
// ¿Cómo la nombrarías?
func ???(???) (???, error) {
    // ¿Cómo manejas errores?
    // ¿Cómo nombras variables?
    // ¿Usas fmt.Errorf con %w?
}
```

**Sección 5: Non-Negotiables**

```markdown
## Non-Negotiables

Estas reglas NO se negocian NUNCA:

1. ✅ Todos los errores deben verificarse (nunca `_` para errores)
2. ✅ Tests antes de merge (unit tests pasan sin DB)
3. ✅ No secrets en código (usar variables de entorno)
4. ✅ Dependency rule respetada (no imports cross-layer incorrectos)
5. ✅ `golangci-lint` pasa sin warnings
6. ✅ [Tu regla adicional...]

**Tu turno:** ¿Cuáles son tus 5-10 reglas innegociables?
```

**Validación:**
Cada regla debe ser:

- ✅ Verificable objetivamente (no "código bonito")
- ✅ Tiene consecuencia clara si se rompe
- ❌ No opiniones subjetivas

**Ejemplo malo:** "Código debe ser elegante" (¿qué es elegante?)
**Ejemplo bueno:** "Coverage mínimo 80%" (medible con herramienta)

#### 3.3 Aplicar Constitution a Proyecto (60 min)

**Acción:**

```bash
cd ~/proyecto-existente

# Link constitution
ln -s ~/dev-constitution/constitution.md ./constitution.md

# Crear CLAUDE.md que referencia constitution
cat > CLAUDE.md << 'EOF'
# Proyecto: [Nombre]

## Constitution
⚠️ CRITICAL: Este proyecto sigue `./constitution.md`

Antes de cualquier código, verifica:
1. Cumplimiento con constitution
2. No violations
3. Si dudas, pregunta primero

## Project-Specific
[Info única de ESTE proyecto]
- Stack: Next.js 14 + Prisma + PostgreSQL
- Auth: NextAuth.js
- Deployment: Vercel

## Current Focus
[En qué estás trabajando AHORA]
EOF
```

**Test de constitution:**

```bash
claude

"Lee constitution.md.

Ahora implementa endpoint POST /api/users siguiendo constitution."
```

Claude debería:

- Usar tu estructura de carpetas
- Aplicar tus naming conventions
- Incluir tests (si es non-negotiable)
- Manejar errores según tu standard

**Validación:**
Revisa el código generado. ¿Sigue TODO lo de tu constitution? Si no, tu constitution no es suficientemente específica.

**Pregunta de comprensión:**
"¿Por qué constitution.md en root del proyecto y no en cada archivo?"
_Respuesta esperada:_ "Porque es contexto global del proyecto. Claude lo lee una vez al inicio de sesión, no necesita releerlo en cada operación. Ahorra tokens."

---

## 🎯 Día 4: Spec-Driven Development - De Vibe a Estructura

### Objetivo

Dominar el flujo Constitution → Spec → Plan → Tasks para eliminar "vibe coding" y trabajar con dirección clara.

### Estrategia

Separar QUÉ (spec) del CÓMO (plan) del HACER (tasks). Pensar antes de codear.

### Justificación

Vibe coding = codear sin dirección → Cambios constantes → Tokens desperdiciados → Código inconsistente. SDD = Plan primero → Implementación eficiente → Menos refactors → 40% menos tokens.

### Qué Harás

#### 4.1 Entender el Flujo SDD (45 min)

**El problema del Vibe Coding:**

```
"Quiero agregar authentication"
        ↓
    [Claude empieza a codear]
        ↓
    "Ah, también necesito reset password"
        ↓
    [Claude agrega código]
        ↓
    "Espera, debería ser OAuth también"
        ↓
    [Claude refactoriza todo]
        ↓
Resultado: 3x tokens, código inconsistente, arquitectura subóptima
```

**El flujo SDD:**

```
"Quiero authentication system"
        ↓
1. SPEC: Definir QUÉ necesitas
   - Login con email/password
   - Reset password
   - OAuth (Google, GitHub)
   - 2FA opcional
        ↓
2. PLAN: Definir CÓMO construirlo
   - Phase 1: Database schema
   - Phase 2: Email/password auth
   - Phase 3: Password reset
   - Phase 4: OAuth integration
   - Phase 5: 2FA
        ↓
3. TASKS: Definir pasos granulares
   - Task 1.1: Prisma schema (2h)
   - Task 1.2: Migrations (1h)
   - Task 2.1: JWT service (2h)
   - ...
        ↓
4. IMPLEMENT: Claude ejecuta tasks
   - Task by task
   - Validación después de cada una
   - Commit después de cada una
        ↓
Resultado: 1x tokens, código consistente, arquitectura sólida
```

**Las 4 capas de SDD:**

```
┌────────────────────────────────────────┐
│ CONSTITUTION (Principios universales)  │ ← "CÓMO trabajo siempre"
└────────────────────────────────────────┘
              ↓ informa
┌────────────────────────────────────────┐
│ SPEC (Qué construir)                   │ ← "QUÉ necesito en esta feature"
└────────────────────────────────────────┘
              ↓ guía
┌────────────────────────────────────────┐
│ PLAN (Cómo construirlo)                │ ← "CÓMO lo divido en fases"
└────────────────────────────────────────┘
              ↓ detalla
┌────────────────────────────────────────┐
│ TASKS (Pasos ejecutables)              │ ← "HACER cada paso"
└────────────────────────────────────────┘
```

**Ejercicio de comprensión:**

Clasifica estos elementos:

1. "Este endpoint debe validar email con Zod"
2. "Feature de notificaciones por email"
3. "Task 3.2: Implementar email service (2h)"
4. "Phase 2: Core business logic"
5. "TypeScript strict mode siempre"

_Respuestas:_

- Constitution: 5
- Spec: 2
- Plan: 4
- Tasks: 3
- Spec/Plan: 1 (detalle técnico)

#### 4.2 Escribir Tu Primera Spec (90 min)

**Feature ejemplo:** "Password Reset Flow"

**Estructura de Spec:**

```markdown
# Feature Specification: [Nombre]

## 1. Executive Summary (2-3 líneas)

¿Qué es y por qué?

## 2. Problem Statement

### Current State (qué NO funciona hoy)

### Desired State (cómo DEBERÍA funcionar)

### Success Metrics (cómo sabrás que funciona)

## 3. User Stories

As a [user], I want to [action], so that [benefit]

## 4. Functional Requirements

Must have / Should have / Nice to have / Out of scope

## 5. Technical Specification

APIs, Database, Models, Security

## 6. Testing Strategy

Qué testear y cómo
```

**Acción:** Crea `~/specs/password-reset/spec.md`

Vamos parte por parte:

**Parte 1: Executive Summary**

```markdown
# Feature Specification: Password Reset Flow

**Status:** Draft
**Author:** Luis Ricardo
**Created:** 2026-03-23

## Executive Summary

Self-service password reset via email to reduce support tickets
(currently 40% of volume) and improve user satisfaction.
```

**Por qué esto primero:**
Si alguien lee solo 3 líneas, debe entender: qué es, por qué importa, qué problema resuelve.

**Validación:** ¿Puedes explicar la feature en 30 segundos?

**Parte 2: Problem Statement**

```markdown
## Problem Statement

### Current State

- Users who forget password must contact support
- Average resolution time: 2 hours
- User frustration high (NPS -20 for this flow)
- 40% of support tickets related to password recovery

### Desired State

- Self-service reset in < 5 minutes
- Zero support intervention
- Secure flow with email verification

### Success Metrics

- Reduce password-related tickets by 80%
- Reset completion rate > 90%
- Time to reset < 5 minutes average
- Zero security incidents
```

**Por qué métricas:**
Sin métricas, no sabrás si la feature fue exitosa. "Mejorar UX" es vago. "Reducir tickets 80%" es medible.

**Ejercicio:** Para tu propia feature, responde:

1. ¿Qué problema específico resuelve?
2. ¿Cómo medirás el éxito?
3. ¿Cuál es el comportamiento actual vs deseado?

**Parte 3: User Stories**

```markdown
## User Stories

### US1: Request Password Reset

**As a** registered user who forgot password
**I want to** request password reset via my email
**So that** I can regain access to my account

**Acceptance Criteria:**

- [ ] Given I'm on login page, when I click "Forgot Password", then I see reset form
- [ ] Given I enter valid email, when I submit, then I receive email within 2 min
- [ ] Given I enter invalid email, when I submit, then I see generic success message (no user enumeration)
- [ ] Given I request reset 3+ times in 1 hour, when I submit 4th time, then I'm rate limited

### US2: Reset Password

**As a** user with reset link
**I want to** set new password
**So that** I can login with new credentials

**Acceptance Criteria:**

- [ ] Given valid reset link (<1h old), when I click, then I see password form
- [ ] Given expired link (>1h old), when I click, then I see error + option to request new
- [ ] Given I set new password, when I submit, then old password invalidated
- [ ] Given I use reset link twice, when I click 2nd time, then link is invalid (single-use)
```

**Por qué formato "Given/When/Then":**
No ambigüedad. Testeable directamente. Cualquiera puede convertir esto en test automatizado.

**Validación de comprensión:**

Convierte este acceptance criteria a test:

"Given valid reset link, when I click, then I see password form"

```go
// internal/application/usecase/verify_reset_token_test.go
func TestVerifyResetToken_ValidToken(t *testing.T) {
    // Arrange (Given)
    repo := repository.NewMockTokenRepository()
    validToken := "abc123"
    repo.Add(entity.PasswordResetToken{Token: validToken, ExpiresAt: time.Now().Add(time.Hour)})
    uc := usecase.NewVerifyResetToken(repo)

    // Act (When)
    result, err := uc.Execute(validToken)

    // Assert (Then)
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if !result.Valid {
        t.Error("expected token to be valid")
    }
}
```

**Parte 4: Technical Specification**

````markdown
## Technical Specification

### API Endpoints

```

POST /api/auth/forgot-password
Body: { email: string }
Response: { message: "If email exists, reset link sent" }
Rate Limit: 3 requests per hour per email

POST /api/auth/reset-password
Body: { token: string, newPassword: string }
Response: { message: "Password updated" }

GET /api/auth/verify-reset-token/:token
Response: { valid: boolean, expiresAt?: string }

```

### Database Schema

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

### Data Flow

```
User → Forgot Password Form
  ↓
API: POST /forgot-password
  ↓
Generate token (32 bytes random)
  ↓
Store in DB (expires 1 hour)
  ↓
Send email with link
  ↓
User clicks link
  ↓
API: Verify token valid
  ↓
Show password form
  ↓
API: POST /reset-password
  ↓
Hash new password
  ↓
Update user.password
  ↓
Mark token as used
  ↓
Success
```

### Security Requirements

- Token: cryptographically secure random (32 bytes)
- Expiration: 1 hour from creation
- Single-use: mark as used after successful reset
- Rate limiting: 3 attempts per email per hour
- No user enumeration: "If email exists..." message
- Password requirements: min 8 chars, 1 uppercase, 1 number
````

**Por qué este nivel de detalle:**
Elimina ambigüedad. Claude (o humano) puede implementar directamente sin preguntar.

**Ejercicio:**
Dibuja el data flow de TU feature. ¿Cuáles son los pasos?

#### 4.3 Generar Plan desde Spec (60 min)

**Concepto:** Plan divide spec en phases implementables.

**Acción:**

```bash
cd ~/specs/password-reset
claude

/model opus-4  # Usa Opus para planning

"Lee spec.md.

Genera implementation plan siguiendo estas reglas:
1. Dividir en phases de max 4 horas cada una
2. Identificar dependencies entre phases
3. Listar archivos a crear/modificar en cada phase
4. Incluir validation checkpoints
5. Seguir constitution.md principles

Output a plan.md"
```

**Claude generará algo como:**

````markdown
# Implementation Plan: Password Reset Flow

**Based on:** spec.md
**Total Estimate:** 14 hours

## Phase 1: Database & Models (2h)

**Objective:** Setup data layer

**Tasks:**

1. Crear entity `PasswordResetToken` en `internal/domain/entity/`
2. Crear interface `PasswordResetTokenRepository` en `internal/domain/repository/`
3. Escribir migration SQL en `migrations/`
4. Implementar `PostgresPasswordResetTokenRepo` en `internal/infrastructure/persistence/`

**Files:**

- `internal/domain/entity/password_reset_token.go` (create)
- `internal/domain/repository/password_reset_token_repository.go` (create)
- `migrations/003_add_password_reset_tokens.sql` (create)
- `internal/infrastructure/persistence/postgres_token_repo.go` (create)

**Validation:**

```bash
make down-clean && make up   # Aplica migration
make db-shell                # Verificar tabla existe: \d password_reset_tokens
go build ./...               # Compila sin errores
```

**Dependencies:** None

---

## Phase 2: Core Use Cases (4h)

**Objective:** Business logic

**Tasks:**

1. Use case `RequestPasswordReset` (genera token, llama email service)
2. Use case `VerifyResetToken` (verifica validez y expiración)
3. Use case `ResetPassword` (aplica nuevo password, invalida token)
4. Unit tests con mocks (>80% coverage)

**Files:**

- `internal/application/usecase/request_password_reset.go` (create)
- `internal/application/usecase/verify_reset_token.go` (create)
- `internal/application/usecase/reset_password.go` (create)
- `internal/application/usecase/*_test.go` (create)

**Validation:**

```bash
go test ./internal/application/usecase/...
# All tests pass
go test -race ./internal/application/usecase/...
# No race conditions
```

**Dependencies:** Phase 1 complete

---

## Phase 3: API Layer (4h)

[Continue similarly...]
````

**Validación de comprensión:**

**Pregunta:** "¿Por qué Phase 2 depende de Phase 1?"
_Respuesta esperada:_ "Phase 2 (services) necesita los models de Phase 1 (database) para funcionar. No puedo crear service que usa PasswordResetToken si el modelo no existe."

**Pregunta:** "¿Qué pasa si implemento Phase 3 antes de Phase 2?"
_Respuesta esperada:_ "Phase 3 (API layer) llama a Phase 2 (services). Si Phase 2 no existe, Phase 3 no compila. Respeto dependencies."

#### 4.4 Generar Tasks desde Plan (45 min)

**Concepto:** Tasks son pasos < 2 horas, implementables independientemente.

```bash
claude

"Lee plan.md.

Genera task breakdown donde cada task:
- Completable en < 2 horas
- Tiene acceptance criteria clara
- Lista archivos específicos
- Incluye comandos de validación

Output a tasks.md"
```

**Claude genera:**

````markdown
# Task Breakdown: Password Reset

## Phase 1: Database & Models

### Task 1.1: Domain Entity & Repository Interface

**Time:** 30 min
**Dependencies:** None

**Description:**
Crear entity `PasswordResetToken` y su interface de repositorio

**Files:**

- `internal/domain/entity/password_reset_token.go`
- `internal/domain/repository/password_reset_token_repository.go`

**Implementation:**
[Struct entity + interface con métodos Create, FindByToken, MarkUsed]

**Acceptance Criteria:**

- [ ] Struct compila sin errores
- [ ] Interface define todos los métodos necesarios
- [ ] Error sentinel `ErrTokenNotFound` exportado

**Validation:**

```bash
go build ./internal/domain/...
```

---

### Task 1.2: Database Migration

**Time:** 30 min
**Dependencies:** Task 1.1

[Continue...]
````

**Lo crítico de tasks:**

- ✅ Pequeñas (< 2h)
- ✅ Independientes (una vez dependencies satisfechas)
- ✅ Verificables (comando de validación)
- ✅ Claras (no ambigüedad)

**Ejercicio:**

Esta task está mal. ¿Por qué?

```markdown
### Task: Implement authentication

**Time:** 8 hours
**Description:** Add auth to the app
```

_Problemas:_

- ❌ Muy grande (8 horas)
- ❌ Vaga ("add auth" = muchas cosas)
- ❌ No tiene acceptance criteria
- ❌ No tiene validation commands

**Corrección:**
Dividir en 8+ tasks pequeñas: JWT service, login endpoint, register endpoint, middleware, tests...

---

## 🎯 Día 5: Ejecución Agéntica - Implementación Automatizada

### Objetivo

Ejecutar el plan task-by-task usando Claude en modo agéntico, con validaciones automáticas y commits incrementales.

### Estrategia

Claude trabaja autónomamente siguiendo tasks.md, pero con checkpoints de validación para mantener calidad.

### Justificación

Implementación manual = context switches constantes. Agéntica = flujo continuo, Claude valida cada paso antes de siguiente. 60% más rápido, menos errores.

### Qué Harás

#### 5.1 Setup de Workspace Agéntico (30 min)

**Concepto:** Preparar proyecto para que Claude tenga TODO el contexto necesario sin re-preguntar.

**Acción:**

```bash
cd ~/projects/password-reset-impl
git init
git checkout -b feature/password-reset

# Estructura de docs
mkdir -p docs/{specs,plans,tasks}

# Copiar artifacts de planning
cp ~/specs/password-reset/spec.md ./docs/specs/
cp ~/specs/password-reset/plan.md ./docs/plans/
cp ~/specs/password-reset/tasks.md ./docs/tasks/

# Link constitution
ln -s ~/dev-constitution/constitution.md ./docs/constitution.md
```

**Crear CLAUDE.md para workflow agéntico:**

````markdown
# Password Reset Implementation

## 🎯 Context Documents

1. **Spec:** `docs/specs/spec.md` (QUÉ construir)
2. **Plan:** `docs/plans/plan.md` (CÓMO en fases)
3. **Tasks:** `docs/tasks/tasks.md` (pasos ejecutables)
4. **Constitution:** `docs/constitution.md` (principios)

## 🤖 Agentic Workflow Rules

### Task Execution Protocol

```

For each task in tasks.md:

1. Read task details
2. Implement ONLY that task (no extra features)
3. Run validation commands
4. If validation fails → Fix → Re-validate
5. If validation passes → Commit → Mark complete
6. Move to next task

```

### Validation Commands (run after EACH task)

```bash
go build ./...         # Compila sin errores
go vet ./...           # Análisis estático
golangci-lint run      # Linting
go test ./...          # Tests pasan
```

### Commit Protocol

```bash
git add .
git commit -m "<type>(<scope>): <task description>

Implements task X.Y from plan

Ref: docs/tasks/tasks.md#taskXY"
```

### Blocker Protocol

If blocked (unclear requirement, missing dependency, etc):

1. **STOP** - Do not proceed
2. Document blocker in tasks.md
3. List possible solutions
4. Wait for human decision

### Deviation Protocol

If spec needs change (found better approach, spec incomplete):

1. **PAUSE** - Do not implement deviation
2. Explain issue with spec
3. Suggest spec update
4. Wait for approval
5. Update spec FIRST
6. Then implement

## 📊 Current Status

**Phase:** 1 - Database & Models
**Current Task:** 1.1 - Prisma Schema
**Completed:** 0/15 tasks

## 🚦 State

[ ] Phase 1: Database & Models
[ ] Phase 2: Core Services
[ ] Phase 3: API Layer
[ ] Phase 4: Testing
[ ] Phase 5: Documentation
````

**Por qué este nivel de detalle:**
Claude en modo agéntico necesita reglas claras de qué hacer, cuándo parar, cómo validar. Sin esto, puede desviarse o tomar malas decisiones.

**Validación de comprensión:**

**Pregunta:** "¿Por qué 'STOP' en blocker protocol en vez de que Claude intente resolver?"
_Respuesta esperada:_ "Porque bloqueos requieren decisiones humanas (arquitectura, trade-offs). Claude podría elegir mal. Mejor parar y preguntar que proceder incorrectamente."

#### 5.2 Sesión Agéntica: Phase 1 (90 min)

**Acción:**

```bash
claude
/model opus-4  # Usa Opus para reasoning inicial

"AGENTIC SESSION START

Context:
- Read docs/specs/spec.md
- Read docs/plans/plan.md
- Read docs/tasks/tasks.md
- Read CLAUDE.md (workflow rules)
- Read docs/constitution.md

Objective:
Complete Phase 1 following agentic workflow protocol.

Instructions:
1. Start with Task 1.1
2. After EACH task:
   - Run validation commands
   - Commit with proper message
   - Update tasks.md (mark [x] complete)
   - Report status
3. STOP if:
   - Validation fails (after attempting fix)
   - Requirement unclear
   - Need human decision

Begin autonomous execution."
```

**Qué debería pasar:**

```
[Claude - Opus 4]
Reading specification... ✓
Reading plan... ✓
Reading tasks... ✓
Reading constitution... ✓

Starting Phase 1: Database & Models

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Task 1.1: Domain Entity & Repository Interface
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Implementation:
- Creating internal/domain/entity/password_reset_token.go
- Creating internal/domain/repository/password_reset_token_repository.go
- Adding ErrTokenNotFound sentinel

[Shows code being written]

Validation:
$ go build ./internal/domain/...
✓ Compila sin errores

$ go vet ./internal/domain/...
✓ Sin issues

Committing:
$ git add internal/domain/
$ git commit -m "feat(domain): add password reset token entity and repository interface

Implements task 1.1 from plan

- PasswordResetToken entity with all required fields
- Repository interface: Create, FindByToken, MarkUsed
- ErrTokenNotFound sentinel exported

Ref: docs/tasks/tasks.md#task11"

✓ Task 1.1 complete
✓ Updated tasks.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Task 1.2: Database Migration
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Implementation:
- Creating migrations/003_add_password_reset_tokens.sql

$ make down-clean && make up   # Aplica migration
✓ Migration applied

$ make db-shell
\d password_reset_tokens       # Verify schema
✓ Table exists with correct columns and indexes

Committing:
[Shows commit]

✓ Task 1.2 complete

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Phase 1 Status: 2/2 tasks complete
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Phase 1 complete. Ready for Phase 2.

Continue to Phase 2? [y/n]
```

**Tu respuesta:** "Yes, continue. Switch to Sonnet for implementation."

```bash
/model sonnet-3.7

[Claude - Sonnet 3.7]
Switching to Sonnet for cost-effective implementation.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Phase 2: Core Services
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Task 2.1: Token Generation Service
[Continues autonomously...]
```

**Lo que observar:**

- ✅ Claude lee tasks.md y sigue secuencia
- ✅ Valida después de cada task
- ✅ Commits individuales por task
- ✅ Actualiza tasks.md marcando completes
- ✅ Reporta status claramente

**Validación de comprensión:**

**Pregunta:** "¿Por qué commit después de cada task en vez de al final?"
_Respuesta esperada:_ "Commits pequeños = historia clara, fácil rollback si algo falla, checkpoint de progreso, mejor code review."

#### 5.3 Manejo de Blockers (30 min)

**Escenario:** Claude encuentra blocker.

```
[Claude]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⚠️  BLOCKER DETECTED
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Task 3.2: Rate Limiting Middleware

Issue:
El proyecto no tiene rate limiting. Spec dice "3 requests per hour per email".

Options:
1. In-memory map con mutex (simple)
   - Pros: Sin dependencias extra, zero setup
   - Cons: Se resetea al reiniciar el servidor, no escala horizontalmente

2. Redis con `go-redis` (persistente)
   - Pros: Persiste entre reinicios, funciona con múltiples instancias
   - Cons: Requiere levantar Redis en docker-compose

3. PostgreSQL como backend de rate limit
   - Pros: Sin deps extra (ya tenemos Postgres)
   - Cons: Más lento que Redis para operaciones de contador

Recommendation: Option 2 (Redis)
Reasoning: Spec requiere persistencia entre requests.
           In-memory falla si el server se reinicia.

Decision needed:
A) Approve Option 2, agrego Redis a docker-compose.yml
B) Option 1 (in-memory es suficiente para este MVP)
C) Option 3 (usar Postgres existente)
D) Revisar spec para no requerir persistencia

Waiting for human input...
```

**Tu decisión:**

```
"Option A approved.

Also:
1. Document this decision in docs/decisions.md (create if needed)
2. Add redis to docker-compose.yml
3. Update README with redis requirement

Proceed."
```

**Por qué este protocolo:**
Claude identifica problema, propone soluciones, justifica recomendación, espera aprobación. No procede ciegamente.

**Ejercicio de comprensión:**

Claude dice:

```
"Spec says use JWT but doesn't specify algorithm.
Should I use HS256 (symmetric) or RS256 (asymmetric)?"
```

¿Qué debería hacer Claude?

- A) Elegir HS256 (más común)
- B) Elegir RS256 (más seguro)
- C) PARAR y preguntar
- D) Implementar ambos

_Respuesta correcta:_ C - PARAR y preguntar

_Por qué:_ Decisión de seguridad crítica. HS256 vs RS256 tiene implicaciones (key distribution, microservices, etc). Requiere decisión humana.

#### 5.4 Session Management & Token Optimization (60 min)

**Concepto:** Sesiones largas consumen tokens. Optimiza cortando y resumiendo.

**Cuándo cortar sesión:**

```
Triggers para `/clear` o nueva sesión:
- Phase completo (ej: terminó Phase 2, empieza Phase 3)
- Conversación > 50 mensajes
- Contexto > 150k tokens
- Cambio de dominio (backend → frontend)
```

**Estrategia de continuación:**

```bash
# Al terminar Phase 2
/compact

# Claude resume todo en bullet points
# Guardas summary

exit

# Nueva sesión para Phase 3
claude

"RESUME SESSION

Context:
- Phases 1-2 complete (see git log for details)
- Current codebase in repo
- Next: Phase 3 (API Layer)

Read:
- docs/tasks/tasks.md (start at Task 3.1)
- Latest code (check git log to understand what's done)

Continue agentic execution from Task 3.1."
```

**Comparación de tokens:**

| Estrategia                     | Tokens Usados             |
| ------------------------------ | ------------------------- |
| Sesión continua (todos phases) | ~300k tokens              |
| `/compact` cada phase          | ~180k tokens (40% ahorro) |
| Nueva sesión cada phase        | ~200k tokens (33% ahorro) |

**Validación:**
Después de cada phase:

```bash
claude
/context

# Si > 100k tokens, considera /compact o nueva sesión
```

---

## 🎯 Día 6: Patterns Agénticos Avanzados

### Objetivo

Dominar subagents para proyectos grandes y workflows paralelos.

### Estrategia

Dividir proyecto en dominios, cada dominio con su propio agente especializado.

### Justificación

Proyectos monolíticos → Contexto explosivo → Pérdida de foco. Subagents → Contexto acotado → Mayor precisión → 50% menos tokens en proyectos grandes.

### Qué Harás

#### 6.1 Cuándo Usar Subagents (45 min)

**Concepto:** Subagent = Claude especializado en un dominio del proyecto.

**Triggers para subagents:**

```markdown
Usa subagents cuando:
✅ Proyecto > 30k líneas de código
✅ Múltiples tecnologías (backend + frontend + mobile)
✅ Equipos separados (frontend team, backend team)
✅ Módulos independientes (auth, payments, analytics)
✅ Microservicios architecture

NO uses subagents cuando:
❌ Proyecto pequeño (< 10k líneas)
❌ Stack simple (solo backend O solo frontend)
❌ Features fuertemente acopladas
❌ Solo tú desarrollando
```

**Ejemplo de cuándo SI:**

```
Proyecto: E-commerce Platform (microservicios)

Estructura:
services/
├── catalog/       # Go API: productos (15k líneas)
├── orders/        # Go API: pedidos (20k líneas)
├── auth/          # Go API: autenticación (10k líneas)
└── notifications/ # Go worker: emails/SMS (8k líneas)

web/               # Frontend React (20k líneas)
admin/             # Admin panel (15k líneas)

Total: 88k líneas → PERFECTO para subagents
```

**Ejemplo de cuándo NO:**

```
Proyecto: Todo List API

Estructura:
internal/
├── domain/        # (0.5k líneas)
├── application/   # (1k líneas)
├── infrastructure/ # (1k líneas)
└── presentation/  # (1k líneas)

Total: 3.5k líneas → Single agent suficiente
```

**Ejercicio de comprensión:**

Para cada proyecto, ¿subagents o single agent?

1. Blog personal (Next.js, 5k líneas)
2. Plataforma bancaria (Backend, Frontend, Mobile, Admin - 200k líneas)
3. API REST simple (Express, 8k líneas)
4. Monorepo con 6 microservicios (150k líneas total)

_Respuestas:_

1. Single (pequeño)
2. Subagents (grande, múltiples apps)
3. Single (pequeño, single tech)
4. Subagents (grande, separable por servicio)

#### 6.2 Arquitectura de Subagents (90 min)

**Setup de ejemplo: E-commerce Platform**

```bash
cd ~/projects/ecommerce-platform

# Estructura de subagents
mkdir -p .claude/agents
```

**Subagent 1: Catalog Service**

`.claude/agents/catalog/CLAUDE.md`:

```markdown
# Catalog Service Subagent

## Domain Boundary

**Scope:** services/catalog/ directory ONLY

**I can:**

- Modify files in services/catalog/
- Read docs/api-contracts/ para contratos con otros servicios

**I cannot:**

- Modify services/orders/ (orders agent's domain)
- Modify services/auth/ (auth agent's domain)
- Change API contracts sin coordinación

## Tech Stack

- Go (stdlib net/http)
- PostgreSQL con database/sql
- Clean Architecture (domain/application/infrastructure/presentation)
- go test + testify para tests

## Architecture
```

services/catalog/
├── cmd/api/main.go
├── internal/domain/
├── internal/application/usecase/
├── internal/infrastructure/persistence/
└── internal/presentation/

```

## API Integration
**Contract:** `docs/api-contracts/`
- Tipos y endpoints definidos en YAML
- Este agente mantiene sus contratos
- Consume contratos de otros servicios, no los modifica

## Focus Areas
- Performance de queries (< 50ms p95)
- Paginación en todos los listados
- Datos correctamente validados

## Constitution
Follows: `../../constitution.md`

## Current Feature
[Updated per feature being implemented]
```

**Subagent 2: Orders Service**

`.claude/agents/orders/CLAUDE.md`:

```markdown
# Orders Service Subagent

## Domain Boundary

**Scope:** services/orders/ directory ONLY

**I can:**

- Modify files in services/orders/
- Modify docs/api-contracts/orders.yaml (I own this contract)
- Call Catalog service via HTTP (read-only)

**I cannot:**

- Modify services/catalog/ (catalog agent's domain)
- Change database schema without migration file
- Deploy without tests passing

## Tech Stack

- Go (stdlib net/http)
- PostgreSQL con database/sql
- Redis para distributed locks
- Clean Architecture

## Architecture
```

services/orders/
├── cmd/api/main.go
├── internal/domain/
├── internal/application/usecase/
├── internal/infrastructure/
│   ├── persistence/   # Postgres repos
│   └── http/          # HTTP clients (catalog service)
└── internal/presentation/

```

## API Contract Management
**I maintain:** `docs/api-contracts/orders.yaml`
**When I change API:**
1. Update contract first
2. Notify dependent services (add comment in contract)
3. Implement backend change
4. Verify contract matches implementation

## Focus Areas
- Consistencia de datos (transacciones)
- Idempotencia en operaciones críticas
- API performance (< 200ms p95)

## Constitution
Follows: `../../constitution.md`
```

**API Contract Document** (comunicación entre agentes):

`docs/api-contracts/products.yaml`:

```yaml
# Product API Contract
# Owner: Backend API Subagent
# Consumers: Web Frontend, Mobile, Admin

endpoints:
  - path: /api/products
    method: GET
    description: List products with pagination
    auth: optional
    query_params:
      - name: page
        type: integer
        default: 1
      - name: limit
        type: integer
        default: 20
        max: 100
      - name: category
        type: string
        optional: true
    response:
      type: object
      properties:
        data:
          type: array
          items: Product
        pagination:
          type: Pagination

  - path: /api/products/:id
    method: GET
    description: Get single product
    auth: optional
    params:
      - name: id
        type: string
        format: cuid
    response:
      type: Product

types:
  Product:
    id: string (cuid)
    name: string
    price: number
    category: string
    inStock: boolean
    createdAt: string (ISO 8601)

  Pagination:
    total: number
    page: number
    limit: number
    hasMore: boolean

# CHANGELOG
# 2026-03-23: Initial version
# [Backend agent adds entries here when changing API]
```

**Validación de comprensión:**

**Pregunta:** "¿Por qué backend agent 'owns' API contracts?"
_Respuesta esperada:_ "Porque backend implementa el contrato. Si frontend modificara contrato pero backend no, habría desincronización. Backend es source of truth."

**Pregunta:** "Frontend agent necesita agregar campo a Product. ¿Qué hace?"
_Respuesta esperada:_ "1) Comenta en el contract solicitando campo, 2) Backend agent evalúa, 3) Backend agrega campo a contract + implementa, 4) Frontend usa nuevo campo."

#### 6.3 Workflow de Subagents (90 min)

**Escenario:** Implementar "Product Search" (toca frontend + backend)

**Step 1: Crear Feature Spec (compartida)**

`docs/specs/product-search.md`:

```markdown
# Feature: Product Search

## Backend Requirements

- Elasticsearch integration
- POST /api/search endpoint
- Filters: price range, category, rating
- Pagination
- Response time < 150ms

## Frontend Requirements

- Search bar with autocomplete
- Filter UI (checkboxes, sliders)
- Results grid
- Loading states
- Empty state

## API Contract

See: docs/api-contracts/search.yaml
```

**Step 2: Catalog Service Agent Session**

```bash
cd services/catalog

claude

"CATALOG SERVICE SUBAGENT SESSION

Context:
- Read ../../docs/specs/product-search.md
- Read ../../.claude/agents/catalog/CLAUDE.md (my boundaries)
- Focus: Catalog service implementation ONLY

Tasks:
1. Design API contract (create ../../docs/api-contracts/search.yaml)
2. Implement full-text search con PostgreSQL tsvector
3. Create GET /api/products/search endpoint
4. Add use case tests + integration tests
5. Update API documentation

Follow agentic workflow protocol.
Begin."
```

Catalog agent trabaja en su dominio.

**Step 3: Web Frontend Agent Session (paralelo)**

```bash
cd web

claude

"WEB FRONTEND SUBAGENT SESSION

Context:
- Read ../docs/specs/product-search.md
- Read ../docs/api-contracts/search.yaml (catalog contract)
- Read ../.claude/agents/web/CLAUDE.md (my boundaries)
- Focus: Frontend implementation ONLY

Tasks:
1. Create SearchBar component
2. Create Filters component
3. Create Results component
4. Integrate with API (use contract as mock during development)
5. Add tests

Assume catalog API works per contract.
Use mock data for development.

Begin."
```

Frontend agent trabaja independientemente.

**Step 4: Integration**

```bash
cd ../..  # Root

claude

"INTEGRATION REVIEW SESSION

Context:
- Backend agent completed: apps/api/...
- Frontend agent completed: apps/web/...
- API Contract: docs/api-contracts/search.yaml

Verify:
1. Backend implementation matches contract
2. Frontend consumption matches contract
3. Types are synchronized (generate TypeScript types from contract)
4. Integration tests pass
5. E2E test needed

Generate integration test and E2E test."
```

**Beneficios de este approach:**

| Aspecto       | Single Agent              | Subagents                          |
| ------------- | ------------------------- | ---------------------------------- |
| Context size  | 150k tokens               | 50k + 40k = 90k (40% ahorro)       |
| Focus         | Difuso (todo el proyecto) | Preciso (solo su dominio)          |
| Parallel work | No                        | Sí (backend + frontend simultáneo) |
| Expertise     | Generalista               | Especialista por dominio           |
| Conflicts     | Alto (toca todo)          | Bajo (boundaries claros)           |

**Validación:**

Después de ambas sessions:

```bash
# Verificar sincronización
git log --all --oneline --graph

# Deberías ver dos branches:
# - feature/search-catalog (catalog agent)
# - feature/search-frontend (frontend agent)

# Merge ambos
git checkout main
git merge feature/search-catalog
git merge feature/search-frontend

# Run integration tests
make up
go test -tags=integration ./...
```

---

## 🎯 Día 7: Production Mastery & Troubleshooting

### Objetivo

Consolidar todo en workflow de producción bulletproof con troubleshooting para problemas comunes.

### Estrategia

Template reutilizable end-to-end + playbook de debugging.

### Justificación

Workflow adhoc = inconsistente. Template estandarizado = repetible, predecible, optimizado.

### Qué Harás

#### 7.1 El Master Workflow Template (120 min)

**Crear template reutilizable para todas tus features:**

`~/templates/feature-workflow/README.md`:

````markdown
# Feature Development Workflow Template

Use este template para CADA nueva feature.

## Phase 0: Setup (5 min)

```bash
# 1. Crear carpeta de feature
mkdir -p ~/features/[feature-name]
cd ~/features/[feature-name]

# 2. Copiar templates
cp -r ~/templates/feature-workflow/* ./

# 3. Inicializar git
git init
git checkout -b feature/[feature-name]

# 4. Link constitution
ln -s ~/dev-constitution/constitution.md ./
```

## Phase 1: Specification (60-90 min)

**Modelo:** Opus 4
**Output:** `spec.md`

```bash
claude
/model opus-4

"New feature specification session.

Feature: [describe in 2-3 sentences]

Guide me through spec creation:
1. Problem statement
2. User stories
3. Technical requirements
4. Success metrics

Use template in docs/templates/spec-template.md"
```

**Checklist antes de continuar:**

- [ ] Problem statement clara
- [ ] User stories con acceptance criteria
- [ ] Technical spec detallada
- [ ] Success metrics definidas
- [ ] Reviewed by human

## Phase 2: Planning (30-45 min)

**Modelo:** Opus 4
**Output:** `plan.md`

```bash
claude

"Read spec.md.

Generate implementation plan:
- Divide en phases (max 4h each)
- Identify dependencies
- List files to create/modify
- Include validation checkpoints
- Follow constitution.md

Output to plan.md"
```

**Checklist:**

- [ ] Phases have clear objectives
- [ ] Dependencies identified
- [ ] Time estimates realistic
- [ ] Validation steps included

## Phase 3: Task Breakdown (30 min)

**Modelo:** Sonnet 3.7
**Output:** `tasks.md`

```bash
/model sonnet-3.7

"Read plan.md.

Generate task breakdown:
- Each task < 2 hours
- Clear acceptance criteria
- Specific files listed
- Validation commands

Output to tasks.md"
```

**Checklist:**

- [ ] All tasks < 2h
- [ ] No ambiguous tasks
- [ ] Validation commands present
- [ ] Numbered sequentially

## Phase 4: Implementation (varies)

**Modelo:** Sonnet 3.7 (implementation), Opus 4 (complex logic)
**Output:** Code + commits

```bash
claude
/model sonnet-3.7

"AGENTIC IMPLEMENTATION SESSION

Context:
- spec.md
- plan.md
- tasks.md
- constitution.md

Execute tasks sequentially:
1. Implement task
2. Validate (lint, types, tests)
3. Commit
4. Mark complete in tasks.md
5. Next task

STOP on:
- Validation failure (after fix attempt)
- Unclear requirement
- Blocker

Begin with Task 1.1"
```

**Monitor cada 30 min:**

```bash
/context           # Check token usage
git log --oneline  # Verify commits
go test ./...      # Tests siguen pasando
```

**Si contexto > 100k tokens:**

```bash
/compact
# O nueva sesión
```

## Phase 5: Review (30 min)

**Modelo:** Opus 4

```bash
claude
/model opus-4

/review

"Comprehensive code review:
- Security issues (SQL injection, auth bypass)
- Go-specific issues (goroutine leaks, error ignoring)
- Constitution compliance (dependency rule, naming)
- Test coverage
- Documentation (godoc en exported symbols)

Generate review report."
```

**Fix issues identificados, re-run review hasta green.**

## Phase 6: Documentation (30 min)

```bash
claude

"Generate documentation:
- README update (if new feature)
- API docs (if new endpoints)
- Architecture decision records (if architectural change)
- CHANGELOG entry"
```

## Phase 7: Deploy Preparation (15 min)

```bash
/deploy

# Claude ejecuta checklist:
# - go test ./... pasa
# - go build ./... compila
# - golangci-lint run sin warnings
# - Env vars configuradas
# - Migrations listas
# - Monitoring configurado
```

## Token Budget

| Phase          | Model         | Est. Tokens | Est. Cost |
| -------------- | ------------- | ----------- | --------- |
| Spec           | Opus 4        | 40k         | $2        |
| Plan           | Opus 4        | 20k         | $1        |
| Tasks          | Sonnet        | 15k         | $0.20     |
| Implementation | Sonnet + Opus | 150k        | $8        |
| Review         | Opus 4        | 30k         | $1.50     |
| Docs           | Sonnet        | 10k         | $0.15     |
| **Total**      |               | **265k**    | **~$13**  |

## Time Budget

| Phase          | Time                |
| -------------- | ------------------- |
| Spec           | 90 min              |
| Plan           | 45 min              |
| Tasks          | 30 min              |
| Implementation | (varies by feature) |
| Review         | 30 min              |
| Documentation  | 30 min              |
| Deploy Prep    | 15 min              |

## Success Metrics

After completing:

- [ ] `go test ./...` pasa (>80% coverage en use cases)
- [ ] `golangci-lint run` sin errores
- [ ] `go build ./...` compila
- [ ] Constitution compliant (dependency rule respetada)
- [ ] Spec fully implemented
- [ ] Godoc en todos los exported symbols
- [ ] Ready for code review
- [ ] Deployable (migrations versionadas, env vars documentadas)
````

**Usar template:**

```bash
# Nueva feature
cp -r ~/templates/feature-workflow ~/features/user-notifications
cd ~/features/user-notifications

# Seguir README.md paso a paso
```

#### 7.2 Troubleshooting Playbook (90 min)

**Crear playbook de problemas comunes:**

`~/docs/troubleshooting.md`:

````markdown
# Claude Code Troubleshooting Playbook

## Problem 1: "Context too large" Error

**Symptoms:**

- Error: "Maximum context length exceeded"
- Claude responses slow/incomplete

**Diagnosis:**

```bash
claude
/context

# Check total tokens
# Identify large files
```

**Solutions:**

1. **Immediate fix:**

```bash
/clear  # Clear conversation history
/remove path/to/large/files/*  # Remove unnecessary files
```

1. **Preventive fix:**

```bash
# Add to .claudeignore:
vendor/
bin/
*.log
coverage.out
coverage.html
```

1. **Nuclear option:**

```bash
exit  # Close session
# Start fresh session
```

**Validation:**
Context should be < 150k tokens after fix.

---

## Problem 2: Claude Violates Constitution

**Symptoms:**

- Generated code doesn't follow your standards
- Uses prohibited patterns
- Wrong naming conventions

**Diagnosis:**
Claude didn't read or apply constitution.

**Solutions:**

1. **Verify constitution is linked:**

```bash
ls -la constitution.md
# Should exist and link to ~/dev-constitution/constitution.md
```

1. **Make constitution enforceable in CLAUDE.md:**

```markdown
# In CLAUDE.md

⚠️ CRITICAL: Read constitution.md FIRST

Before ANY code:

1. Read constitution.md
2. Verify alignment
3. If violation detected, STOP and report
```

1. **Explicit reminder:**

```bash
claude

"Before proceeding, read constitution.md and confirm:
1. You've read it
2. You understand the non-negotiables
3. You'll apply it to all code

Confirm understanding."
```

**Validation:**
Ask Claude: "What are my non-negotiables from constitution?"
Should list them accurately.

---

## Problem 3: Skills Not Loading

**Symptoms:**

- `/skillname` does nothing
- No error, just silent failure

**Diagnosis:**

```bash
# Check skill exists
ls -la ~/.claude/skills/skillname/

# Check SKILL.md exists
cat ~/.claude/skills/skillname/SKILL.md
```

**Common causes:**

1. **Missing frontmatter:**

```markdown
❌ WRONG:

# My Skill

✅ CORRECT:

---

name: skillname
description: What it does

---

# My Skill
```

1. **Wrong filename:**

```bash
❌ ~/.claude/skills/skillname/skill.md  (lowercase)
✅ ~/.claude/skills/skillname/SKILL.md  (uppercase)
```

1. **Permissions:**

```bash
chmod +r ~/.claude/skills/skillname/SKILL.md
```

**Solutions:**

1. Verify structure:

```bash
~/.claude/skills/skillname/
└── SKILL.md  ← Must exist, must be readable
```

1. Validate frontmatter:

```bash
head -n 5 ~/.claude/skills/skillname/SKILL.md

# Should show:
# ---
# name: skillname
# description: ...
# ---
```

1. Test:

```bash
claude
/skillname

# Should execute skill
```

---

## Problem 4: MCP Server Connection Failed

**Symptoms:**

- MCP calls timeout
- "Server not responding" errors

**Diagnosis:**

```bash
# List configured servers
claude mcp list

# Check logs
tail -f ~/.claude/logs/mcp.log
```

**Solutions:**

1. **Verify server is running:**

```bash
# For remote servers:
curl https://mcp-server-url/health

# For STDIO servers:
which npx
npm list -g @modelcontextprotocol/server-*
```

1. **Restart server:**

```bash
claude mcp remove servername
claude mcp add servername --transport remote --url [url]
```

1. **Check credentials:**

```bash
# Verify API keys in env
env | grep API_KEY
env | grep TOKEN
```

1. **Network issues:**

```bash
# Ping server
ping mcp-server-domain.com

# Check firewall
# [Arch Linux specific]
sudo iptables -L
```

**Validation:**

```bash
claude

"Test MCP server: [servername]
Make simple call to verify connectivity."
```

---

## Problem 5: Agentic Execution Stuck

**Symptoms:**

- Claude stops mid-execution
- Infinite loop on task
- No progress for >5 min

**Diagnosis:**
Check where it's stuck:

```bash
# Check git log
git log --oneline

# Last completed task?
cat docs/tasks/tasks.md | grep "$$x$$"
```

**Solutions:**

1. **Interrupt and resume:**

```bash
# In Claude session
[Ctrl+C or interrupt]

"Status check:
- What task were you executing?
- What step failed?
- What's the blocker?

Report and wait for instructions."
```

1. **Manual fix:**

```bash
# Fix the issue manually
# Then tell Claude:

"I manually fixed [issue].
Resume from Task X.Y"
```

1. **Skip problematic task:**

```bash
"Skip Task X.Y for now.
Mark as [skipped] in tasks.md.
Continue to Task X.(Y+1)"
```

**Prevention:**
Add timeout to CLAUDE.md:

```markdown
## Timeout Protocol

If task takes >15 min:

1. STOP
2. Report: "Task X.Y timeout (>15 min)"
3. Ask: "Continue, skip, or debug?"
```

---

## Problem 6: Generated Code Doesn't Match Spec

**Symptoms:**

- Implementation diverges from spec
- Features missing
- Wrong behavior

**Diagnosis:**

```bash
claude

"Spec alignment audit:

Compare:
- spec.md (requirements)
- Implemented code

Report:
1. Requirements NOT implemented
2. Implementations NOT in spec
3. Deviations"
```

**Solutions:**

1. **Identify gap:**

```bash
# Claude generates gap report:
# - Missing: Rate limiting en POST /forgot-password (req 3.2)
# - Extra: Admin endpoint para reset manual (not in spec)
# - Deviation: Token expiry 2h en vez de 1h especificada
```

1. **Prioritize:**

```markdown
Critical (blocking):

- [ ] Implement missing requirement 3.2

Medium (important):

- [ ] Remove extra feature (not in scope)

Low (nice to fix):

- [ ] Align deviation (or update spec)
```

1. **Fix task-by-task:**

```bash
"Implement missing requirement 3.2 from spec.
Follow agentic protocol (implement, validate, commit)."
```

**Prevention:**
Add to CLAUDE.md:

```markdown
## Spec Compliance Check

After each Phase:

1. Run spec alignment audit
2. Fix gaps before next Phase
3. Update spec if intentional deviation
```

---

## Problem 7: Token Cost Too High

**Symptoms:**

- Bills higher than expected
- Burning through quota quickly

**Analysis:**

```bash
# Check usage
# (Depends on Claude pricing dashboard)

# Estimate tokens per session
claude
/context
```

**Solutions:**

1. **Apply all optimizations:**

- [ ] .claudeignore comprehensive (vendor/, bin/, coverage.out)
- [ ] CLAUDE.md < 200 tokens
- [ ] /clear between features
- [ ] /compact long sessions
- [ ] Model selection (Sonnet default)

1. **Audit expensive operations:**

```bash
# Review git log
git log --all --oneline

# Count commits per day
git log --since="1 week ago" --oneline | wc -l

# If >50 commits/day, sessions are too granular
```

1. **Batch operations:**

```bash
❌ EXPENSIVE:
Task 1 → commit → Task 2 → commit → Task 3 → commit

✅ CHEAPER:
Task 1 → Task 2 → Task 3 → commit all
(Only if tasks are related)
```

1. **Use cheaper models:**

```bash
# Default to Sonnet
/model sonnet-3.7

# Use Opus only for:
# - Architecture decisions
# - Complex debugging
# - Critical logic (concurrency, security)
```

**Target costs:**

- Small feature (< 1k LOC): $5-10
- Medium feature (1-5k LOC): $10-20
- Large feature (5k+ LOC): $20-40

If above target, optimize more aggressively.
````

**Validación de comprensión:**

**Ejercicio:** Tienes estos síntomas:

1. Claude genera código que usa GORM en vez de `database/sql` directo
2. Naming conventions incorrectas (camelCase en nombres de archivo)
3. Tests no incluidos

¿Qué problema es? ¿Cómo lo arreglas?

_Respuesta:_

- Problema: #2 (Claude Violates Constitution)
- Fix: 1) Verificar constitution.md está linkeada, 2) Hacer enfático en CLAUDE.md, 3) Confirmar que Claude leyó constitution antes de continuar

---

¿Listo para implementar todo? ¿Algún día específico necesitas que profundice más? También puedo ayudarte con integración específica para tu stack de Go si quieres.
