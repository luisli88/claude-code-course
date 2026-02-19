# Week 5 — Advanced Git: Own Your History

## Objective

Use Git the way professional teams do — clean commits, intentional history, confident conflict resolution, and a review workflow that makes your PRs easy to approve.

---

## 1. Git config baseline

Set this up once:

```bash
git config --global user.name "Your Name"
git config --global user.email "you@example.com"
git config --global core.editor nvim
git config --global core.pager delta
git config --global pull.rebase true          # always rebase on pull
git config --global rebase.autoSquash true    # --fixup commits auto-squash
git config --global merge.conflictstyle zdiff3  # cleaner conflict markers
```

Useful aliases (add to `~/.gitconfig`):

```ini
[alias]
  lg   = log --oneline --graph --all --decorate
  st   = status -sb
  aa   = add --all
  cm   = commit -m
  co   = checkout
  sw   = switch
  oops = commit --amend --no-edit
  undo = reset HEAD~1 --mixed
  unstage = restore --staged
```

---

## 2. Commit hygiene

Every commit should be a single logical change that can be understood in isolation.

```bash
# Stage only specific lines (not whole files) — review every hunk
git add -p

# Stage a specific file
git add src/handler.go

# Write a good commit message
git commit
# Format: <type>(<scope>): <short summary>
# Types: feat, fix, refactor, test, docs, chore
# Example: feat(auth): add POST /login with JWT response

# Amend last commit (before pushing)
git commit --amend
git commit --amend --no-edit  # keep same message

# Quick alias for amend
git oops

# Undo last commit but keep changes staged
git undo
```

---

## 3. Interactive rebase

The most powerful Git tool. Rewrite history before it goes to a PR.

```bash
# Rebase last N commits
git rebase -i HEAD~5

# Rebase onto the remote main
git fetch origin
git rebase -i origin/main
```

Inside the rebase editor:

```
pick   a1b2c3 feat: add /ping endpoint
pick   d4e5f6 fix typo
pick   g7h8i9 add tests for /ping
pick   j0k1l2 fix another typo
pick   m3n4o5 refactor ping handler
```

Rewrite to:

```
pick   a1b2c3 feat: add /ping endpoint
fixup  d4e5f6 fix typo          ← squash, discard message
squash g7h8i9 add tests for /ping  ← squash, keep message
fixup  j0k1l2 fix another typo
reword m3n4o5 refactor ping handler  ← keep but edit message
```

```bash
# After rebase, force-push your feature branch
git push --force-with-lease  # safer than --force
```

### Fixup workflow (preferred)

```bash
# Make a small fix that should be squashed into an earlier commit
git add -p
git commit --fixup a1b2c3   # references the commit SHA to squash into

# When ready to clean up:
git rebase -i --autosquash origin/main
# fixup commits are automatically placed and marked
```

---

## 4. Branching and PR workflow

```bash
# Always branch from up-to-date main
git switch main
git pull
git switch -c feat/add-login-endpoint

# Keep branch up to date during development
git fetch origin
git rebase origin/main  # not merge — keeps history linear

# Push and set tracking
git push -u origin feat/add-login-endpoint

# After PR is merged — clean up
git switch main
git pull
git branch -d feat/add-login-endpoint
```

---

## 5. Conflict resolution

```bash
# During a rebase that hits a conflict:
git status              # see which files conflict

nvim <conflicting-file>
# Find conflict markers:
# <<<<<<< HEAD          ← your changes
# =======
# >>>>>>> origin/main   ← incoming changes
# Edit to the correct final state, remove all markers

git add <resolved-file>
git rebase --continue

# To abort and start over:
git rebase --abort
```

With `merge.conflictstyle = zdiff3`, you also see the **base** (common ancestor) which makes resolution much clearer:

```
<<<<<<< HEAD
  your version
||||||| base
  original version
=======
  their version
>>>>>>> origin/main
```

---

## 6. Inspection and archaeology

```bash
# Visual commit graph
git lg
# or:
git log --oneline --graph --all

# What changed in this commit?
git show <sha>

# Who changed this line and when?
git blame -L 10,20 main.go

# What commits touched this file?
git log --oneline -- main.go

# Search commit messages
git log --oneline --grep="login"

# Search code changes across history
git log -S "LoginHandler" --oneline  # commits that added/removed this string
git log -G "func.*Login" --oneline   # commits where diff matches this regex
```

---

## 7. git bisect — find regressions

```bash
# Start bisect
git bisect start
git bisect bad              # current commit is broken
git bisect good v1.0.0      # this tag/sha was working

# Git checks out a midpoint commit — run your test:
go test ./...

git bisect good   # if tests pass
git bisect bad    # if tests fail

# Git keeps narrowing — repeat until it identifies the breaking commit
# When done:
git bisect reset
```

Automate it:

```bash
git bisect start HEAD v1.0.0
git bisect run go test ./...  # Git runs this automatically at each step
```

---

## 8. GitHub CLI (gh)

```bash
# Install
brew install gh
gh auth login

# Create a PR from current branch
gh pr create --title "feat: add POST /login" --body "Implements SPEC-001"

# List open PRs
gh pr list

# Review a PR
gh pr checkout 42
gh pr review 42 --approve
gh pr review 42 --request-changes --body "Missing error handling on line 42"

# Merge a PR
gh pr merge 42 --squash --delete-branch

# Create an issue
gh issue create --title "SPEC-002: POST /register" --label "spec"

# View CI status
gh pr checks
```

---

## Exercise 1 — Clean up a messy branch

```bash
git switch -c practice/messy
claude "Make 6 small commits: add /ping, /pong, /version, /status, /ready, and /metrics endpoints. Use vague commit messages like 'update', 'fix stuff', 'changes'."

# Now clean it up into 2 commits: "feat: add health endpoints" and "feat: add metadata endpoints"
git rebase -i HEAD~6
```

## Exercise 2 — Bisect a regression

```bash
claude "Make 10 commits. In commit 6, introduce a bug that makes TestHealth fail. Use realistic commit messages."
go test ./...  # fails

git bisect start
git bisect bad
git bisect good HEAD~10
git bisect run go test ./...
# Verify Git identified commit 6
git bisect reset
```

## Exercise 3 — Full PR workflow with gh

```bash
git switch -c feat/add-ping
claude "Add GET /ping returning 200 'pong' with a test."
go test ./...
git add -p
git commit -m "feat(api): add GET /ping endpoint"
git push -u origin feat/add-ping
gh pr create --title "feat: add /ping endpoint" --body "$(claude -p 'Write a GitHub PR description for the current git diff')"
gh pr checks
```

---

## Resources

- [Git rebase docs](https://git-scm.com/docs/git-rebase)
- [GitHub CLI docs](https://cli.github.com/manual/)
- [Conventional Commits spec](https://www.conventionalcommits.org)
- [Oh Shit, Git!](https://ohshitgit.com) — recovery from common mistakes

---

## Pass gate

You can take a 6-commit branch with vague history and rebase it into 2 clean commits, write a good commit message in conventional format, and open a PR with `gh` — all without touching the GitHub website.