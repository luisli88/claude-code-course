# Week 7 — Spec Driven Development: Write Specs Claude Executes Perfectly

## Objective

Master SDD end-to-end: writing specs, sequencing them, handing them to Claude, and verifying outcomes. A great spec is the difference between one correct implementation and five frustrating iterations.

---

## 1. What is a spec?

A spec is a **contract**, not a design doc. It describes:
- **What** the system does (observable behavior)
- **How to verify it** (acceptance criteria → tests)
- **Context** Claude needs (data shapes, constraints, what already exists)

It does **not** describe implementation internals.

---

## 2. Full spec format

```markdown
# SPEC-NNN — Short imperative title

## Objective
One sentence. What user-facing capability does this add?

## Context
- Relevant data shapes, existing endpoints, or constraints
- What this spec builds on (reference prior SPECs if applicable)
- Stack constraints (e.g., "stdlib only", "no new files")

## GIVEN / WHEN / THEN

- GIVEN [system state / precondition]
  WHEN [action / HTTP method + path + payload]
  THEN [observable outcome: status code, body, headers, side effects]

- GIVEN [error scenario]
  WHEN [...]
  THEN [...]

(Cover: happy path + all meaningful error paths)

## Acceptance criteria
- [ ] Specific, binary, testable statement
- [ ] Maps 1:1 to a test case
- [ ] Covers the happy path
- [ ] Covers each error scenario from GIVEN/WHEN/THEN
- [ ] Includes a coverage requirement if applicable

## Out of scope
- What this spec intentionally does NOT cover (prevents scope creep)
```

---

## 3. The SDD workflow

```
1. Write the spec
      ↓
2. Review it yourself: can I write a test from each acceptance criterion?
      ↓
3. Hand to Claude
   claude "Implement specs/spec-NNN.md exactly."
      ↓
4. Read the diff:
   - Does every acceptance criterion have a corresponding test?
   - Is anything outside scope being changed?
      ↓
5. Run tests: go test ./...
      ↓
6. If all criteria pass → commit
   If not → add a clarifying prompt (don't edit manually)
      ↓
7. Update the spec if Claude's question reveals an ambiguity
```

**Never edit Claude's output manually.** If it's wrong, the spec was ambiguous. Fix the spec, re-prompt.

---

## 4. What makes a spec excellent

### Context section

Bad:
```markdown
## Context
- User auth system
```

Good:
```markdown
## Context
- Extends SPEC-001 (POST /login)
- User store is `Store` in store.go — map[string]string (email → password)
- Pre-seeded: admin@example.com / secret
- Token format from SPEC-001: "token.for.<email>"
- Stack: Go 1.20, stdlib only, no new packages
```

### GIVEN/WHEN/THEN scenarios

Bad (missing error paths):
```markdown
- GIVEN a valid user
  WHEN POST /login with correct credentials
  THEN 200 + token
```

Good (all paths covered):
```markdown
- GIVEN user admin@example.com exists with password "secret"
  WHEN POST /login {"email":"admin@example.com","password":"secret"}
  THEN 200 {"token":"token.for.admin@example.com"}

- GIVEN user exists
  WHEN POST /login with wrong password
  THEN 401 {"error":"invalid credentials"}

- GIVEN any state
  WHEN POST /login with missing "email" field
  THEN 400 {"error":"email is required"}

- GIVEN any state
  WHEN POST /login with missing "password" field
  THEN 400 {"error":"password is required"}

- GIVEN any state
  WHEN POST /login with empty body {}
  THEN 400

- GIVEN any state
  WHEN POST /login with malformed JSON
  THEN 400
```

### Acceptance criteria

Bad (vague, can't verify):
```markdown
- [ ] Handle errors
- [ ] Write tests
- [ ] Fast response
```

Good (binary, maps to a test):
```markdown
- [ ] POST /login returns 200 + {"token":"token.for.<email>"} for valid credentials
- [ ] POST /login returns 401 for wrong password
- [ ] POST /login returns 400 {"error":"email is required"} for missing email
- [ ] POST /login returns 400 {"error":"password is required"} for missing password
- [ ] POST /login returns 400 for empty body
- [ ] All scenarios covered by table-driven tests in handler_test.go
- [ ] go test -cover ./... shows >80%
```

---

## 5. Spec sequencing

Features have dependencies. Write specs in the right order and reference previous specs in Context.

```
SPEC-001: POST /login
    ↓
SPEC-002: POST /register      (needs user store from SPEC-001)
    ↓
SPEC-003: GET /me             (needs JWT from SPEC-001, user from SPEC-002)
    ↓
SPEC-004: DELETE /account     (needs auth from SPEC-003)
```

Reference in Context:
```markdown
## Context
- Builds on SPEC-001: users are stored in store.go as map[string]string
- Builds on SPEC-002: registered users are queryable by email
- JWT validation uses the token format from SPEC-001
```

---

## 6. Spec types

### API endpoint spec (most common)
Cover: method, path, request shape, response shape, auth requirements, error codes.

### Data model spec
Cover: fields, types, validation rules, constraints (unique, required), example values.

```markdown
## GIVEN / WHEN / THEN
- GIVEN a User struct
  WHEN Email field is empty
  THEN validation returns error "email is required"

- GIVEN a User struct
  WHEN Email is not valid format
  THEN validation returns error "invalid email format"
```

### Background job spec
Cover: trigger condition, input, processing steps, output/side effects, failure behavior.

```markdown
- GIVEN a user registers
  WHEN registration completes successfully
  THEN a welcome email job is enqueued with the user's email
  AND the job is processed within 5 seconds
  AND on failure it retries up to 3 times
```

---

## 7. Common mistakes

| Mistake | Effect | Fix |
|---------|--------|-----|
| No error scenarios in GIVEN/WHEN/THEN | Claude produces no error handling | Add a GIVEN for every error path |
| Vague acceptance criteria | Hard to verify Claude's output | Make each criterion a binary test assertion |
| Spec describes implementation | Over-constrains Claude | Describe behavior, not internal structure |
| Missing context on existing code | Claude makes wrong assumptions | Name the exact types, files, and functions to use |
| One giant spec for a feature | Claude misses edge cases | Split into one spec per endpoint or concern |
| Editing Claude's output manually | Spec diverges from code | Re-prompt until the spec drives the output |

---

## Exercise 1 — Implement spec-login.md

```bash
cat specs/spec-login.md

claude "Implement specs/spec-login.md exactly. Do not add anything not in the spec."
go test -cover ./...

# Check every acceptance criterion against the actual output
```

## Exercise 2 — Write SPEC-002: POST /register

Write `specs/spec-002-register.md` covering:
- Success: 201 + `{"message":"registered"}`
- Duplicate email: 409 `{"error":"email already registered"}`
- Missing fields: 400
- Invalid email format: 400
- User is findable after registration (store query)

Then implement it:

```bash
claude "Implement specs/spec-002-register.md. Build on store.go from SPEC-001."
go test -cover ./...
```

## Exercise 3 — Spec-to-PR pipeline

Write SPEC-003 (`GET /me` — return logged-in user's email from the token), implement it, then:

```bash
claude -p "Write a GitHub PR description for the changes in this branch. Reference SPEC-001, SPEC-002, and SPEC-003."
gh pr create --title "feat: implement auth flow (SPEC-001 through SPEC-003)" --body "$(claude -p 'PR description for current changes')"
```

---

## Resources

- [BDD / Gherkin syntax primer](https://cucumber.io/docs/gherkin/)
- `specs/spec-login.md` — working example in this repo

---

## Pass gate

You write a spec for `POST /register` with at least 5 GIVEN/WHEN/THEN scenarios and 6 acceptance criteria, hand it to Claude, and get a correct implementation with >80% coverage on the first or second attempt — without writing any code yourself.