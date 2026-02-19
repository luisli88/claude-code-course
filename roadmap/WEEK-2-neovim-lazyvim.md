# Week 2 — Neovim + LazyVim: Navigate Without a Mouse

## Objective

Navigate, edit, search, and refactor a real codebase entirely from the keyboard. Understand how Neovim's plugin system works so you can extend it yourself.

---

## 1. Config structure

LazyVim starter gives you this layout out of the box:

```
~/.config/nvim/
├── init.lua                  ← one line: require("config.lazy") — don't touch
└── lua/
    ├── config/
    │   ├── lazy.lua          ← lazy.nvim bootstrap + LazyVim spec — don't touch
    │   ├── options.lua       ← add your vim options here
    │   ├── keymaps.lua       ← add your keymaps here
    │   └── autocmds.lua      ← add your autocmds here
    └── plugins/
        └── init.lua          ← your plugins (the only file from dotfiles/)
```

**You only edit files in `lua/plugins/` and `lua/config/options|keymaps|autocmds.lua`.** Everything else is owned by LazyVim.

Add a plugin by appending to `lua/plugins/init.lua` and running `:Lazy sync`.

---

## 2. Essential keybinds

### Navigation

```
gd          Go to definition (LSP)
gr          Go to all references (LSP)
K           Hover documentation (LSP)
gi          Go to implementation (LSP)
<ctrl+o>    Jump back in location history
<ctrl+i>    Jump forward in location history
%           Jump to matching bracket/paren
[[  ]]      Jump to previous/next function
```

### File finding (Telescope / fzf-lua)

```
<leader>ff  Find file by name (fuzzy)
<leader>fr  Recent files
<leader>sg  Live grep — search file contents
<leader>sw  Search current word under cursor
<leader>sb  Search open buffers
<leader>e   Toggle file explorer
```

### Code actions (LSP)

```
<leader>ca  Code action (auto-fix, imports, etc.)
<leader>cr  Rename symbol across project
<leader>cf  Format current file
<leader>cd  Show diagnostic (error detail)
[d  ]d      Previous/next diagnostic
```

### Windows and buffers

```
<ctrl+h/j/k/l>   Move between splits
:vsp <file>       Vertical split
:sp <file>        Horizontal split
<leader>bd        Close current buffer
<leader>bb        Switch to previous buffer
gt  gT            Next/previous tab
```

### Editing

```
gcc         Comment/uncomment line (LazyVim built-in)
gc          Comment selection (visual mode)
ci"         Change inside quotes
da(         Delete around parentheses
=           Auto-indent selection
>  <        Indent/dedent selection
```

---

## 3. LSP setup per language

### Go

```bash
go install golang.org/x/tools/gopls@latest
```

In `init.lua`:

```lua
require('lspconfig').gopls.setup({
  settings = {
    gopls = {
      analyses = { unusedparams = true },
      staticcheck = true,
    },
  },
})
```

Verify: open a `.go` file → `:LspInfo` → should show `gopls` attached.

### TypeScript / JavaScript

```bash
npm install -g typescript typescript-language-server
```

```lua
require('lspconfig').ts_ls.setup({})
```

### Lua (for editing init.lua itself)

```bash
brew install lua-language-server
```

```lua
require('lspconfig').lua_ls.setup({
  settings = { Lua = { diagnostics = { globals = { 'vim' } } } }
})
```

---

## 4. Treesitter — syntax-aware navigation

Treesitter understands code structure (not just text). Install parsers for your languages:

```vim
:TSInstall go javascript typescript lua python
```

With Treesitter active you get:
- Accurate syntax highlighting
- `]m` / `[m` — jump to next/previous method
- `]f` / `[f` — jump to next/previous function
- `vaf` — select around function (visual mode)
- `vif` — select inside function

---

## 5. Adding a plugin

Example: add `nvim-autopairs` (auto-close brackets):

```lua
require('lazy').setup({
  { 'neovim/nvim-lspconfig' },
  { 'nvim-treesitter/nvim-treesitter' },
  { 'mfussenegger/nvim-dap' },
  -- add here:
  {
    'windwp/nvim-autopairs',
    event = 'InsertEnter',
    config = true,
  },
})
```

Then: `:Lazy sync` to install.

---

## 6. Custom keymaps

Add to `init.lua` after the `require('lazy')` block:

```lua
local map = vim.keymap.set

-- Save with ctrl+s
map('n', '<C-s>', ':w<CR>', { desc = 'Save file' })

-- Move lines up/down in visual mode
map('v', 'J', ":m '>+1<CR>gv=gv")
map('v', 'K', ":m '<-2<CR>gv=gv")

-- Quick fix list navigation
map('n', '<leader>xn', ':cnext<CR>', { desc = 'Next quickfix' })
map('n', '<leader>xp', ':cprev<CR>', { desc = 'Prev quickfix' })

-- Run go test on save (for Go files)
vim.api.nvim_create_autocmd('BufWritePost', {
  pattern = '*.go',
  callback = function()
    vim.cmd('!go test ./... &')
  end,
})
```

---

## 7. Useful commands to know

```vim
:LspInfo          -- Show attached language servers
:LspRestart       -- Restart the LSP for this buffer
:Lazy             -- Open plugin manager UI
:Lazy sync        -- Install/update all plugins
:Lazy clean       -- Remove unused plugins
:TSInstall <lang> -- Install a Treesitter parser
:checkhealth      -- Diagnose nvim/plugin issues
:Mason            -- UI to install LSP servers (if using mason.nvim)
```

---

## Exercise 1 — Codebase navigation

```bash
cd examples/go-microservice
nvim .
```

1. Find `main.go` with `<leader>ff`
2. Jump to the `http.HandleFunc` call, press `gd` — go to the handler definition
3. Press `gr` — see all references to that handler
4. Press `K` on `http.ResponseWriter` — read the hover docs
5. Search for all TODO comments: `<leader>sg` → type `TODO`
6. Open `handler_test.go` in a vertical split: `:vsp handler_test.go`
7. Navigate between splits with `<ctrl+h>` and `<ctrl+l>`

## Exercise 2 — Refactor with LSP

Rename the inline handler to a named function using only Neovim:

1. Position cursor on the handler variable/name
2. Press `<leader>cr` → type new name → Enter
3. Press `gr` to verify all usages updated
4. Press `<leader>cf` to format the file
5. Check diagnostics: `<leader>cd` on any red underline

## Exercise 3 — Extend the config

Add `nvim-autopairs` and a keymap that runs `go test ./...` from inside nvim:

```lua
map('n', '<leader>tt', ':!go test ./...<CR>', { desc = 'Run tests' })
```

---

## Resources

- [LazyVim docs](https://lazyvim.org)
- [nvim-lspconfig server list](https://github.com/neovim/nvim-lspconfig/blob/master/doc/configs.md)
- [Treesitter supported languages](https://github.com/nvim-treesitter/nvim-treesitter#supported-languages)

---

## Pass gate

You can open an unfamiliar Go file, jump to any function's definition, find all usages, rename a symbol project-wide, and view a diagnostic — all without leaving Neovim or touching a mouse.