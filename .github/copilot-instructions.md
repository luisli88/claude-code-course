# Copilot Instructions

## Repository overview

A structured 8-week developer training program focused on two primary goals: becoming a **vibe programmer with Claude Code** and an expert in **Spec Driven Development (SDD)**. Content is in English.

Structure:
- `roadmap/` — weekly guides (WEEK-0 through WEEK-8), each with objectives, exercises, and a pass gate
- `evaluation/` — `self-checklist.md` (weekly Yes/No gates) and `skill-matrix.md` (1–5 ratings across 6 skills)
- `specs/` — example specs in `SPEC-NNN` format
- `examples/go-microservice/` — reference Go service with tests and CI
- `dotfiles/` — starter configs for Neovim (`nvim/init.lua`) and Ghostty (`ghostty/config.conf`)

## Go microservice

Located at `examples/go-microservice/`. Uses Go 1.20, stdlib only.

```bash
# Run all tests
cd examples/go-microservice && go test ./...

# Run a single test
cd examples/go-microservice && go test -run TestHealth ./...

# Coverage
cd examples/go-microservice && go test -cover ./...
```

CI runs `go test ./...` on PRs to `main` via `.github/workflows/ci.yml`.

## Key conventions

### Spec format (SPEC-NNN)
Specs live in `specs/` and follow:
```
# SPEC-NNN — Title
## Objective
## Context
## GIVEN / WHEN / THEN
## Acceptance criteria  (checkboxes mapping to tests)
## Out of scope
```

### CLAUDE.md (agent rules)
`CLAUDE.md` at the repo root is read by Claude Code before every session. It defines stack constraints (stdlib only, no external packages), forbidden actions (never modify `*_test.go`), test conventions (table-driven), and run commands. Edit mid-session with `/memory` inside a `claude` session.

### MCP servers
Project MCP config lives in `.claude/settings.json`. Add servers with `claude mcp add <name> <package>`. See Week 1 roadmap for the full MCP setup workflow.

### Roadmap file structure
Each `roadmap/WEEK-N-*.md` contains: Objective, Core concepts, Exercises, Resources, and a **Pass gate** (explicit binary test of readiness to advance).

### SDD workflow
The core loop: write spec → `claude "Implement specs/spec-NNN.md"` → review diff → run tests → commit. Never edit Claude's output manually; fix the spec or add a clarifying prompt instead.

### Neovim config
Uses LazyVim with `Space` as the leader key. Plugins: `nvim-lspconfig`, `nvim-treesitter`, `nvim-dap`.

### Ghostty keybinds
`ctrl+h` / `ctrl+l` to navigate between panes. Standard 3-pane layout: Claude | nvim | shell.

## Setup

```bash
# macOS (includes delta)
./install-macos.sh   # git neovim ripgrep fzf bat delta jq go node python3

# Arch Linux
./install-arch.sh    # git neovim ripgrep fzf bat jq go nodejs npm python
```
