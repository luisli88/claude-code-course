# Week 1 — Claude Code CLI: Vibe Programming

## Objective

Ship a real feature using Claude Code as your co-pilot. Understand the full command surface, how to configure Claude's behavior with rules, how to connect external tools via MCP, and how to steer the agent effectively.

---

## Core concept: vibe programming

Vibe programming is working at a higher level of abstraction. You describe *what* you want and *how to verify it*. Claude handles *how to build it*. Your job:

1. Write a clear spec
2. Give Claude the spec
3. Review the diff critically
4. Run the tests
5. Iterate with prompts — never by editing output manually

---

## 1. Command flow

### Interactive mode vs. one-shot

```bash
# Interactive mode — persistent session, full context window
claude
# Then type naturally: "Add a /ping endpoint and test it."

# One-shot — run and exit (good for scripts and CI)
claude "Add a GET /ping endpoint that returns 200 with body 'pong'."

# One-shot with stdin (pipe a spec file in)
cat specs/spec-login.md | claude "Implement this spec exactly."

# Print mode — output only, no interactive UI (good for piping)
claude -p "Summarize the changes in this diff." < my.diff
```

### Slash commands (inside interactive mode)

Type these inside a `claude` session:

| Command | What it does |
|---------|-------------|
| `/help` | Show all available commands |
| `/clear` | Clear conversation history (reset context) |
| `/compact` | Summarize and compress conversation to save context |
| `/memory` | Open your `CLAUDE.md` rules file to edit inline |
| `/init` | Scaffold a `CLAUDE.md` for the current project |
| `/review` | Ask Claude to review the current git diff |
| `/pr_comments` | Pull open PR review comments into context |
| `/quit` | Exit the session |

### Useful CLI flags

```bash
# Specify a model
claude --model claude-opus-4-5 "Implement SPEC-001."

# Limit how many turns Claude takes autonomously
claude --max-turns 5 "Refactor the handlers package."

# Skip confirmation prompts (careful — auto-approves file writes)
claude --dangerously-skip-permissions "Fix the failing tests."

# Continue the most recent session
claude --continue

# Resume a specific session by ID
claude --resume <session-id>
```

---

## 2. Rules — CLAUDE.md

`CLAUDE.md` is the file Claude reads before every session to understand your project's conventions, constraints, and preferences. It is the most powerful way to tune Claude's behavior.

### File locations (layered, all are read)

```
~/.claude/CLAUDE.md          # Global — applies to all your projects
<repo-root>/CLAUDE.md        # Project-level — committed, shared with team
<repo-root>/.claude/CLAUDE.md # Alternative project-level location
```

### Bootstrap a project CLAUDE.md

```bash
# Inside a claude session:
/init

# Or create manually:
cat > CLAUDE.md << 'EOF'
# Project rules

## Stack
- Go 1.20, stdlib only — no external packages
- All HTTP handlers in handlers/ package
- All tests are table-driven

## Conventions
- Return errors as JSON: {"error": "message"}
- Use log/slog for all logging
- Do not modify existing tests — only add new ones

## Commands
- Test: go test ./...
- Coverage: go test -cover ./...
EOF
```

### What to put in CLAUDE.md

```markdown
# CLAUDE.md

## Stack
- Language/framework versions
- Which packages are allowed or forbidden

## Code conventions
- Naming patterns, file structure rules
- Error response format
- Logging style

## Test conventions
- Test file naming, test helper patterns
- Coverage requirements

## What Claude must never do
- Never change test files (only add new tests)
- Never add new dependencies without asking
- Never remove error handling

## Commands to know
- How to run tests
- How to build
- How to lint
```

### Edit rules mid-session

```bash
# Inside claude interactive mode:
/memory
# Opens your CLAUDE.md in $EDITOR — save and close, Claude picks it up immediately
```

---

## 3. MCP — Model Context Protocol

MCP connects Claude to external tools: databases, APIs, filesystems, GitHub, browser automation, etc. An MCP server exposes tools that Claude can call during a session.

### Add an MCP server

```bash
# Add from npm package
claude mcp add @modelcontextprotocol/server-github

# Add with a custom name
claude mcp add github-tools @modelcontextprotocol/server-github

# Add a local server (path to binary or script)
claude mcp add my-db-tool /path/to/mcp-server

# Add with environment variables
claude mcp add github-tools @modelcontextprotocol/server-github \
  --env GITHUB_TOKEN=ghp_xxx

# Scope to current project only (not global)
claude mcp add --scope project @modelcontextprotocol/server-filesystem
```

### Manage MCP servers

```bash
claude mcp list           # Show all configured servers
claude mcp remove <name>  # Remove a server
claude mcp get <name>     # Show server details and status
```

### MCP config file

Servers are stored in `.claude/settings.json` (project) or `~/.claude/settings.json` (global):

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    },
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
    },
    "postgres": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-postgres"],
      "env": {
        "DATABASE_URL": "${DATABASE_URL}"
      }
    }
  }
}
```

### Using MCP tools in a session

Once configured, Claude can call MCP tools automatically. You can also prompt explicitly:

```bash
# With GitHub MCP configured:
claude "Create a GitHub issue titled 'Add rate limiting to POST /login' with label 'enhancement'."

# With filesystem MCP:
claude "Read /tmp/api-spec.yaml and implement the /users endpoint from it."

# With postgres MCP:
claude "Look at the users table schema and generate a handler that queries it."
```

### Useful MCP servers for this project

| Server | Install | Use case |
|--------|---------|----------|
| `@modelcontextprotocol/server-github` | `claude mcp add github ...` | Create issues, PRs, read repo |
| `@modelcontextprotocol/server-filesystem` | `claude mcp add fs ...` | Read/write files outside CWD |
| `@modelcontextprotocol/server-postgres` | `claude mcp add db ...` | Query DB schema, generate handlers |
| `@modelcontextprotocol/server-fetch` | `claude mcp add fetch ...` | Fetch URLs, read API docs |

---

## 4. Tuning the agent

### Permission control

Claude asks for approval before running shell commands or writing files. You can control this:

```bash
# Auto-approve everything in this session (use carefully)
claude --dangerously-skip-permissions

# Set default permission mode in settings
# In ~/.claude/settings.json:
{
  "permissions": {
    "allow": [
      "Bash(go test *)",
      "Bash(git diff*)",
      "Write(**/*.go)"
    ],
    "deny": [
      "Bash(rm *)",
      "Bash(git push*)"
    ]
  }
}
```

### Config settings

```bash
# View all current settings
claude config list

# Set a default model
claude config set model claude-sonnet-4-5

# Enable/disable auto-compact (compresses context when near limit)
claude config set autoCompact true

# View a specific setting
claude config get model
```

### Keeping context healthy

Claude's context window fills up. These habits keep it clean:

```bash
# Inside a session — summarize and compress history
/compact

# If Claude starts going off-track or repeating itself — reset
/clear

# Start fresh for a new task (new session)
# Just exit and re-run: claude
```

### When Claude goes wrong

| Symptom | Fix |
|---------|-----|
| Claude edits files outside scope | Add scope constraint to CLAUDE.md or the prompt |
| Claude modifies tests | Add "Never modify test files" to CLAUDE.md |
| Claude adds dependencies | Add "stdlib only" to CLAUDE.md |
| Claude's output is too large | Break the task into smaller specs |
| Claude repeats a mistake | `/clear` and re-prompt with the constraint made explicit |
| Claude asks too many clarifying questions | Add more context to CLAUDE.md or the spec |

---

## Exercise 1 — Full session flow

```bash
cd examples/go-microservice

# 1. Bootstrap a CLAUDE.md for this project
claude
/init
# Edit to add: stdlib only, table-driven tests, no test modifications
/memory

# 2. Implement a spec
claude "Implement specs/spec-login.md exactly."

# 3. Review the diff
/review

# 4. Check tests
# (in shell pane) go test ./...

# 5. Commit
# (in shell pane) git add -A && git commit -m "feat: implement SPEC-001 login"
```

## Exercise 2 — Add and use an MCP server

```bash
# Add the GitHub MCP server
claude mcp add github @modelcontextprotocol/server-github \
  --env GITHUB_TOKEN=$(gh auth token)

# Verify it loaded
claude mcp list

# Use it in a session
claude "Create a GitHub issue in this repo: 'Implement POST /register — see SPEC-002'. Add label 'spec'."
```

## Exercise 3 — CLAUDE.md iteration

Start with a minimal CLAUDE.md, run Claude on a task, identify what went wrong, and fix the rules:

```bash
# Round 1: Claude adds a dependency
# → Add to CLAUDE.md: "Never add external packages. Stdlib only."

# Round 2: Claude changes a test
# → Add to CLAUDE.md: "Never modify files matching *_test.go"

# Round 3: Claude makes a huge refactor when you wanted a small change
# → Add to CLAUDE.md: "Prefer minimal diffs. Don't refactor unless the spec requires it."
```

---

## Resources

- [Claude Code documentation](https://docs.anthropic.com/en/docs/claude-code/overview)
- [Claude Code slash commands](https://docs.anthropic.com/en/docs/claude-code/slash-commands)
- [MCP introduction](https://docs.anthropic.com/en/docs/claude-code/mcp)
- [CLAUDE.md reference](https://docs.anthropic.com/en/docs/claude-code/memory)
- [MCP server registry](https://github.com/modelcontextprotocol/servers)

---

## Pass gate

- You have a `CLAUDE.md` in the project that prevents Claude from modifying tests or adding dependencies
- You implemented `specs/spec-login.md` via Claude with tests passing
- You have at least one MCP server configured (`claude mcp list` shows it)
- You know how to use `/compact`, `/clear`, and `/memory` mid-session