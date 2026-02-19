# Week 4 — Terminal Mastery: Find Anything in Seconds

## Objective

Master `ripgrep`, `fzf`, `bat`, `delta`, and `jq` so deeply that searching a codebase or processing data feels instant. Build reusable shell workflows you'll use every day.

---

## 1. ripgrep (rg) — fast search

rg replaces `grep` and `find` for code search. It respects `.gitignore` automatically.

### Essential patterns

```bash
# Basic search
rg "http.HandleFunc"

# Limit to a file type
rg "http.HandleFunc" --type go

# Show context (3 lines after, 1 before)
rg "ListenAndServe" -A 3 -B 1

# Only show filenames with matches
rg -l "package main"

# Count matches per file
rg -c "func"

# Case-insensitive
rg -i "login"

# Search for literal string (no regex)
rg -F "w.WriteHeader(200)"

# Exclude directory
rg "error" --glob '!vendor/**'

# Multiple patterns (OR)
rg "TODO|FIXME|HACK|XXX"

# Multiline match
rg -U "func.*\n.*return nil"

# Show only matching part (not whole line)
rg -o "func \w+"

# Search with filename and line numbers
rg -n "HandleFunc" --type go

# Replace preview (doesn't write — use sed for actual replace)
rg "oldName" --replace "newName" --passthru
```

### Search → edit workflow

```bash
# Find all TODO comments and open each file in nvim
rg -l "TODO" | xargs nvim

# Jump to exact line
nvim +42 main.go

# Open all files matching a pattern
nvim $(rg -l "LoginHandler")
```

---

## 2. fzf — fuzzy finder

fzf turns any list into a searchable, interactive picker.

### Core usage

```bash
# Fuzzy find a file by name and open in nvim
nvim $(fzf)

# With preview (syntax highlighted via bat)
nvim $(fzf --preview 'bat --color=always {}')

# Fuzzy select from git log
git log --oneline | fzf

# Fuzzy select a branch to checkout
git branch | fzf | xargs git checkout

# Fuzzy select and kill a process
ps aux | fzf | awk '{print $2}' | xargs kill

# Fuzzy search command history
history | fzf --tac | sed 's/^ *[0-9]* *//'
```

### Live grep with preview (killer combo)

```bash
# Search file contents with fzf + bat preview at the matching line
rg --line-number --no-heading "" \
  | fzf --delimiter : \
        --preview 'bat --highlight-line {2} --color=always {1}' \
        --preview-window '~3' \
  | awk -F: '{print "+"$2, $1}' \
  | xargs nvim
```

Add as alias in `~/.zshrc`:

```bash
alias rgf='rg --line-number --no-heading "" | fzf --delimiter : --preview "bat --highlight-line {2} --color=always {1}" | awk -F: "{print \"+\"\$2, \$1}" | xargs nvim'
```

### Useful shell aliases

```bash
# Fuzzy open file
alias vf='nvim $(rg --files | fzf --preview "bat --color=always {}")'

# Fuzzy checkout git branch
alias gcb='git branch | fzf | xargs git checkout'

# Fuzzy search and open at line
alias vg='rgf'

# Fuzzy kill process
alias fkill='ps aux | fzf --header-lines=1 | awk "{print \$2}" | xargs kill'
```

### fzf keybindings (add to ~/.zshrc)

```bash
# Enable built-in fzf keybindings
[ -f ~/.fzf.zsh ] && source ~/.fzf.zsh
# ctrl+r → fuzzy history search
# ctrl+t → fuzzy file insert
# alt+c  → fuzzy cd into directory
```

---

## 3. bat — better cat

```bash
# View file with syntax highlighting and line numbers
bat main.go

# No pager (print inline)
bat --pager=never main.go

# Just the content, no decorations (good for piping)
bat -p main.go

# Specific lines only
bat -r 10:25 main.go

# Show specific language highlighting on a file without extension
bat -l json response.txt

# Diff two files (colored)
bat --diff old.go new.go

# Use as MANPAGER
export MANPAGER="sh -c 'col -bx | bat -l man -p'"
```

---

## 4. delta — better git diffs

delta integrates with git automatically once configured in `~/.gitconfig`.

### Full gitconfig setup

```ini
[core]
  pager = delta

[interactive]
  diffFilter = delta --color-only

[delta]
  navigate = true       # n/N to jump between diff sections
  side-by-side = true   # show old/new side by side
  line-numbers = true
  syntax-theme = Dracula  # or: GitHub, Monokai Extended, Nord

[merge]
  conflictstyle = zdiff3
```

### Using delta

```bash
# Normal diff (delta handles it automatically)
git diff
git show HEAD
git log -p

# Navigate between changed sections: n / N
# Toggle side-by-side: delta --side-by-side / delta --no-side-by-side

# Diff two arbitrary files
delta old.go new.go
```

---

## 5. jq — JSON processor

Essential for working with APIs and JSON data in the terminal.

```bash
# Pretty-print JSON
curl -s https://api.github.com/repos/cli/cli | jq .

# Extract a field
curl -s https://api.github.com/repos/cli/cli | jq .stargazers_count

# Filter an array
echo '[{"name":"alice","age":30},{"name":"bob","age":25}]' | jq '.[] | select(.age > 27)'

# Select specific fields
curl -s https://api.github.com/repos/cli/cli | jq '{name: .name, stars: .stargazers_count}'

# Count array items
echo '{"items":[1,2,3]}' | jq '.items | length'

# Parse JSON from a Go test output
go test -json ./... | jq 'select(.Action == "fail")'

# Convert JSON array to newline-separated values
echo '["a","b","c"]' | jq -r '.[]'

# Compact output (single line)
cat data.json | jq -c .
```

---

## 6. Building compound workflows

Real power comes from composing tools:

```bash
# Find all Go files with TODOs, show them with bat highlighting
rg -l "TODO" --type go | xargs bat --style=header,numbers

# Find failing tests, pretty-print JSON output
go test -json ./... | jq -r 'select(.Action=="fail") | "\(.Package): \(.Test)"'

# Search API response fields matching a pattern
curl -s https://api.github.com/users/octocat | jq 'to_entries[] | select(.key | test("url")) | .key'

# Find large files in the project
rg --files | xargs du -sh | sort -rh | head -20

# Interactive log search: find a commit, show its diff
git log --oneline | fzf | awk '{print $1}' | xargs git show | delta
```

---

## Exercise 1 — Codebase scavenger hunt

Answer these questions using only terminal tools (no nvim, no GitHub):

1. Which file registers the `/health` route? (`rg "health" --type go`)
2. How many functions are defined in the project? (`rg -c "^func"`)
3. Which test file has the most lines? (`wc -l $(rg -l "_test.go")`)
4. What is the Go module name? (`rg "^module" go.mod`)
5. Find every line that returns a non-200 status code.

## Exercise 2 — Build your personal workflow

Add these to your `~/.zshrc` and verify each works:

```bash
alias vf='nvim $(rg --files | fzf --preview "bat --color=always {}")'
alias gcb='git branch | fzf | xargs git checkout'
alias rgf='rg --line-number --no-heading "" | fzf --delimiter : --preview "bat --highlight-line {2} --color=always {1}" | awk -F: "{print \"+\"\$2, \$1}" | xargs nvim'
```

## Exercise 3 — jq API workflow

```bash
# Hit the GitHub API and find repos with >100 stars
curl -s "https://api.github.com/users/anthropics/repos" \
  | jq '[.[] | {name: .name, stars: .stargazers_count}] | sort_by(-.stars) | .[:5]'
```

---

## Resources

- [ripgrep user guide](https://github.com/BurntSushi/ripgrep/blob/master/GUIDE.md)
- [fzf examples](https://github.com/junegunn/fzf#examples)
- [delta configuration](https://dandavison.github.io/delta/configuration.html)
- [jq manual](https://jqlang.github.io/jq/manual/)

---

## Pass gate

You can find any function, string, or file in this repo in under 5 seconds, open it at the right line in nvim, and process a JSON API response with `jq` — all without leaving the terminal.