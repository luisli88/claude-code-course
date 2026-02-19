# Week 0 — Setup: Reproducible Dev Environment

## Objective

Get a fully working development environment that you can rebuild from scratch in under 10 minutes on any machine. Every tool has a purpose — understand what you're installing and why.

---

## 1. Install core tools

```bash
# macOS
./install-macos.sh
# Installs via Homebrew:
#   git        — version control
#   neovim     — terminal editor
#   ripgrep    — fast code search (rg)
#   fzf        — fuzzy finder
#   bat        — syntax-highlighted cat
#   delta      — syntax-highlighted git diffs
#   jq         — JSON processor
#   go         — Go language
#   node       — Node.js (needed for Claude Code and MCP servers)
#   python3    — scripting

# Arch Linux
./install-arch.sh
# Uses pacman — same tools except delta (install separately: yay -S git-delta)
```

Verify everything:

```bash
nvim --version        # should be 0.9+
go version            # should be 1.20+
node --version        # should be 18+
rg --version
fzf --version
bat --version
delta --version
jq --version
```

---

## 2. Install Claude Code CLI

```bash
npm install -g @anthropic-ai/claude-code

# Verify
claude --version

# Authenticate (opens browser on first run)
claude
# Follow the OAuth flow, then exit with /quit
```

---

## 3. Dotfiles setup

### Neovim

Install LazyVim using the official starter:

```bash
# Back up existing config if any
mv ~/.config/nvim{,.bak} 2>/dev/null

# Clone the official LazyVim starter
git clone https://github.com/LazyVim/starter ~/.config/nvim
rm -rf ~/.config/nvim/.git

# Drop in the project's custom plugins
cp dotfiles/nvim/lua/plugins/init.lua ~/.config/nvim/lua/plugins/

# First launch — LazyVim installs everything automatically
nvim
# Wait for plugin installation (~1 min), then :q
```

Then enable the required extras inside nvim with `:LazyExtras`:

| Extra | What it provides |
|-------|-----------------|
| `dap.core` | **Required** — nvim-dap config function, dap-ui, DAP keybinds. Without this, any lang extra that uses nvim-dap (go, java, typescript) will error on startup |
| `lang.go` | gopls, delve, goimports, golangci-lint |

> **Important:** always enable extras via `:LazyExtras` — never import them manually in `lua/plugins/`. LazyVim tracks them in `lazyvim.json` and importing them twice causes startup errors.

### Ghostty

```bash
mkdir -p ~/.config/ghostty
cp dotfiles/ghostty/config.conf ~/.config/ghostty/config
```

What the config sets:
- `font_size 12`
- `ctrl+h` → previous pane
- `ctrl+l` → next pane

### Git (delta integration)

Add to `~/.gitconfig`:

```ini
[core]
  pager = delta
[delta]
  navigate = true
  side-by-side = true
  line-numbers = true
[interactive]
  diffFilter = delta --color-only
```

### Shell config

Add to `~/.zshrc` (or `~/.bashrc`):

```bash
# fzf keybindings and completion
[ -f ~/.fzf.zsh ] && source ~/.fzf.zsh

# bat as default pager for man pages
export MANPAGER="sh -c 'col -bx | bat -l man -p'"

# Go binary path
export PATH="$PATH:$(go env GOPATH)/bin"

# Quick file open with fuzzy preview
alias vf='nvim $(rg --files | fzf --preview "bat --color=always {}")'
```

---

## 4. Ghostty pane layout

Open Ghostty. Split into three vertical panes:

| Pane | Command | Purpose |
|------|---------|---------|
| Left | `claude` | Claude Code — your co-pilot |
| Center | `nvim .` | Editor — read and review code |
| Right | shell | Tests, git, one-off commands |

**Pane navigation:** `ctrl+h` (left) / `ctrl+l` (right)

To split in Ghostty:
- `cmd+d` — vertical split (macOS)
- `cmd+shift+d` — horizontal split

---

## 5. Language servers (LSP)

Install the language servers you'll use. Neovim's LSP client connects to these automatically:

```bash
# Go
go install golang.org/x/tools/gopls@latest

# TypeScript / JavaScript
npm install -g typescript typescript-language-server

# Lua (for editing init.lua)
brew install lua-language-server   # macOS
# or: yay -S lua-language-server    # Arch

# Verify gopls is on PATH
gopls version
```

---

## 6. Verification checklist

- [ ] `nvim --version` shows 0.9+
- [ ] `go version` shows 1.20+
- [ ] `node --version` shows 18+
- [ ] `claude --version` works and you're authenticated
- [ ] `rg --version`, `fzf --version`, `bat --version`, `delta --version` all work
- [ ] `gopls version` works
- [ ] Ghostty opens with 3 panes
- [ ] Neovim opens without errors, LazyVim loads plugins
- [ ] `git diff` shows syntax-highlighted output via delta
- [ ] `vf` alias works (fuzzy file open)

---

## 7. Troubleshooting

| Problem | Fix |
|---------|-----|
| LazyVim shows errors on first open | Run `:Lazy sync` inside nvim |
| `gopls` not found | Check `$(go env GOPATH)/bin` is on `$PATH` |
| Claude auth loop | `claude auth logout` then re-run `claude` |
| `delta` not used in git | Verify `[core] pager = delta` in `~/.gitconfig` |
| Ghostty splits not working | Check `config` file is at `~/.config/ghostty/config` |
| fzf keybindings not working | Re-run `$(brew --prefix)/opt/fzf/install` |

---

## Pass gate

You can open this repo in Neovim, navigate files without a mouse, run `claude` in the left pane, run `go test ./...` in the right pane — and `git diff` shows delta-highlighted output.