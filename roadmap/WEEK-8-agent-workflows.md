# Week 8 — Agent Workflows: Claude as a Controlled Autonomous Agent

## Objective

Design and run multi-step Claude workflows for entire feature milestones. Learn how to structure agent prompts, set checkpoints, manage context, handle failure, and know when to take back control.

---

## 1. The spectrum: assistant → co-pilot → agent

| Mode | You write | Claude writes | Review frequency |
|------|-----------|---------------|-----------------|
| Assistant | Everything | Snippets on request | Per snippet |
| Co-pilot (Weeks 1–7) | Specs | Implementation | Per spec |
| Agent | Goal + constraints | Plan + multiple specs | Per checkpoint |

Agent mode is high leverage but requires good upfront design. A vague goal → unpredictable output.

---

## 2. Anatomy of a good agent prompt

Structure every agent prompt with four parts:

```
Goal:        What the end state looks like (observable, testable)
Context:     Existing code, specs, constraints to respect
Steps:       Ordered list of tasks (reference spec files)
Checkpoints: Where to pause and wait for your review before continuing
```

Example:

```
claude "
Goal: Implement the full user auth flow. All tests must pass. Coverage >80%.

Context:
- Go 1.20 stdlib only
- See CLAUDE.md for all project conventions
- Specs are in specs/ — implement them in order

Steps:
1. Implement specs/spec-001-login.md
2. Implement specs/spec-002-register.md
3. Implement specs/spec-003-get-me.md
4. Run go test -cover ./... and show coverage

Checkpoints:
- After step 1: show test output before continuing
- After step 3: show full coverage report before continuing
"
```

---

## 3. Checkpoint strategies

Checkpoints give you control without micromanaging.

### By spec

```
After implementing each spec, output:
CHECKPOINT [SPEC-NNN]: <test output>
Then wait for my 'continue' or feedback before the next spec.
```

### By risk

```
Before deleting or renaming any file, describe the change and wait for my approval.
Before modifying any *_test.go file, stop and explain why.
```

### By output size

```
If your planned changes exceed 100 lines, summarize the plan first and wait for my approval.
```

### Preview before write

```
Before writing any code, output the list of files you plan to create or modify. Wait for my 'go ahead'.
```

---

## 4. Scope constraints

The most important part of an agent prompt. Without them, Claude may:
- Refactor code you didn't ask to change
- Add dependencies
- Modify tests
- Create files in unexpected locations

```
Constraints:
- Only modify files in examples/go-microservice/
- Do not add new packages to go.mod
- Do not modify any *_test.go files — only add new test files
- Do not refactor code outside the scope of each spec
- If you are unsure about something, stop and ask rather than guess
```

---

## 5. Context management across long sessions

Agent tasks fill the context window. Manage it:

```bash
# Inside a claude session:

# Compact when context gets heavy (auto-summarizes history)
/compact

# Clear context entirely and start fresh (use between unrelated agent tasks)
/clear

# Resume a previous session (preserves conversation)
claude --continue
claude --resume <session-id>
```

For very long agent tasks, break into sessions:

```bash
# Session 1: specs 1-2
claude "Implement spec-001 and spec-002. Stop after both pass."

# Session 2: specs 3-4 (fresh context, full CLAUDE.md as anchor)
claude "CLAUDE.md is the project rules. spec-001 and spec-002 are already implemented. Now implement spec-003 and spec-004."
```

---

## 6. Parallel agent patterns

Some tasks are independent and can run in parallel (separate terminal panes):

```bash
# Pane 1: implement the store layer
claude "Implement store.go and store_test.go per CLAUDE.md conventions. No handlers yet."

# Pane 2: write the remaining specs (no code writes)
claude "Write specs/spec-002-register.md and specs/spec-003-get-me.md based on the patterns in spec-001-login.md."

# After both complete, session 3: wire it together
claude "store.go and the specs are done. Implement spec-002 and spec-003 using the existing store."
```

---

## 7. Handling agent failure

| Symptom | Root cause | Recovery |
|---------|-----------|----------|
| Changes outside defined scope | Constraints too loose | `/clear`, tighten constraints, re-run |
| Modified test files | No constraint against it | Add to CLAUDE.md + re-run from last good commit |
| Added external dependency | No "stdlib only" constraint | `git checkout go.mod go.sum`, add constraint |
| Diff is 3× larger than expected | Task scope too broad | Split the spec, re-run one at a time |
| Claude asks many clarifying questions | Spec too vague | Answer in the spec's Context section |
| Implementation wrong after 2+ tries | Spec acceptance criteria unclear | Rewrite the failing criterion as a concrete test assertion |

Recovery workflow:

```bash
# Identify last good state
git log --oneline

# Reset to last good commit
git reset --hard <good-sha>

# Improve the spec or constraints
nvim specs/spec-NNN.md   # or CLAUDE.md

# Re-run
claude "Implement specs/spec-NNN.md. [tighter constraint]."
```

---

## 8. Agent prompt templates

### Feature milestone

```
claude "
Goal: Implement [feature name]. All specs must pass with >80% coverage.

Context: [summary of relevant existing code and CLAUDE.md rules]

Specs to implement in order:
1. specs/spec-NNN.md
2. specs/spec-MMM.md

After each spec: output 'CHECKPOINT: SPEC-NNN — [pass/fail] — [test count]'
Wait for my 'continue' before the next spec.
"
```

### Refactoring agent

```
claude "
Goal: Refactor [target] without changing behavior.

Constraints:
- All existing tests must still pass before and after
- Do not add new tests (only if a test was wrong)
- Show the new file/package structure before writing any code
- Do not touch files outside [directory]

Before starting: show me the current structure and your planned new structure.
Wait for my approval.
"
```

### Debug agent

```
claude "
Goal: Find and fix the bug causing [test/behavior] to fail.

Constraints:
- Do not modify test files
- Fix only the root cause — no unrelated changes
- Show me each file you plan to change before editing it

Start by explaining your hypothesis for the root cause.
"
```

### Documentation agent

```
claude "
Goal: Write godoc comments for all exported functions in [package].

Constraints:
- Only add comments — do not change any code
- Follow Go doc conventions (start with function name)
- Skip functions that already have a comment

Show me the list of functions you'll document before writing anything.
"
```

---

## Exercise 1 — Full auth milestone

```bash
# Write all three specs first (or use existing ones)
ls specs/

claude "
Goal: Implement the full auth flow. All tests pass. Coverage >80%.

Context: See CLAUDE.md. Go stdlib only.

Steps:
1. specs/spec-001-login.md (if not already done)
2. specs/spec-002-register.md
3. specs/spec-003-get-me.md

Checkpoints: After each spec, output test results and wait for 'continue'.
"
```

## Exercise 2 — Refactoring agent

```bash
claude "
Goal: Refactor examples/go-microservice to use this structure:
- handlers/ package: all HTTP handlers
- store/ package: data store
- main.go: wiring only

Constraints:
- All tests must pass before and after (go test ./...)
- Do not add new tests
- Do not add external packages

Show me the new structure before writing any code.
"
```

## Exercise 3 — Debug agent

```bash
claude "Introduce a bug that causes POST /login to accept any password for admin@example.com. Spread it across 2 files. Don't tell me what it is."
go test ./...  # should fail

claude "
Goal: Find and fix the security bug — POST /login accepts any password.

Constraints:
- Do not modify test files
- Fix only the root cause
- Show me each file before editing it

Start with your root cause hypothesis.
"
```

---

## Resources

- [Claude Code agent patterns](https://docs.anthropic.com/en/docs/claude-code/agent-patterns)
- [Claude Code hooks](https://docs.anthropic.com/en/docs/claude-code/hooks)
- [Multi-agent workflows](https://docs.anthropic.com/en/docs/claude-code/sub-agents)

---

## Pass gate

You run a 3-spec feature milestone as an agent with checkpoints, review and approve each checkpoint before continuing, and the final result has all tests passing and >80% coverage — with your total input being the initial prompt plus checkpoint approvals.