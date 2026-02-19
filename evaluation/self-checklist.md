# Weekly Self-Checklist

Answer Yes/No for each item before advancing to the next week. Be honest — this is for you.

---

## Week 0 — Setup

- [x] `nvim --version` shows 0.9+
- [x] `go version` shows 1.20+
- [ ] `claude --version` works
- [ ] Ghostty opens with 3 panes (Claude / nvim / shell)
- [ ] Neovim loads without errors
- [ ] LazyVim plugins installed
- [ ] I can switch panes without touching the mouse

---

## Week 1 — Claude Code CLI

- [ ] I can ask Claude to implement a feature and review the diff
- [ ] I know how to re-prompt Claude when its output is wrong (instead of editing manually)
- [ ] I generated a PR description using Claude
- [ ] I handed `specs/spec-login.md` to Claude and got a passing implementation
- [ ] I did not write any implementation code myself

---

## Week 2 — Neovim + LazyVim

- [ ] I can jump to a function definition with `gd`
- [ ] I can find all usages of a symbol with `gr`
- [ ] I can fuzzy-search files with `<leader>ff`
- [ ] I can live-grep across the project with `<leader>sg`
- [ ] I can split windows and navigate between them with `ctrl+h/l`
- [ ] I can rename a symbol project-wide with `<leader>cr`

---

## Week 3 — Debugging

- [ ] I installed `delve` (`dlv version` works)
- [ ] I can set a breakpoint with `<leader>db`
- [ ] I can step into a function with `<leader>di`
- [ ] I can inspect a variable's value with `<leader>dK`
- [ ] I used the debugger (not print statements) to find a bug

---

## Week 4 — Terminal Mastery

- [ ] I can search file contents with `rg` and a type filter
- [ ] I can fuzzy-find a file with `fzf` and open it in nvim
- [ ] I can view a file with syntax highlighting using `bat`
- [ ] Git diffs show with `delta` (syntax highlighted, side-by-side)
- [ ] I have a shell alias that combines rg + fzf + nvim

---

## Week 5 — Advanced Git

- [ ] I can rebase interactively and squash commits
- [ ] I can stage specific lines with `git add -p`
- [ ] I can resolve a rebase conflict and continue
- [ ] I can use `git bisect` to find a regression commit
- [ ] My feature branches have clean, readable commit history

---

## Week 6 — Go Backend

- [ ] `go test ./...` passes
- [ ] `go test -cover ./...` shows >80%
- [ ] I used table-driven tests
- [ ] I added structured logging with `log/slog`
- [ ] CI passes on my PR

---

## Week 7 — Spec Driven Development

- [ ] I understand the GIVEN/WHEN/THEN format
- [ ] I wrote a spec with acceptance criteria checkboxes
- [ ] I implemented a spec using only Claude (no manual code)
- [ ] My spec covered at least 3 error scenarios
- [ ] Claude's implementation matched every acceptance criterion

---

## Week 8 — Agent Workflows

- [ ] I gave Claude a multi-spec task with checkpoints
- [ ] I reviewed diffs at each checkpoint before approving
- [ ] I used scope constraints in my agent prompts
- [ ] Claude completed all 3 specs without going out of scope
- [ ] I know the signs that an agent task is going wrong

