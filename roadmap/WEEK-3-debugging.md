# Week 3 — Debugging: Step Through Code with nvim-dap

## Objective

Debug programs interactively inside Neovim using `nvim-dap`. No more `print` debugging. Cover Go, JavaScript, TypeScript, Swift, Kotlin, and Java — all with the same keybind muscle memory.

---

## Core concept: how nvim-dap works

nvim-dap is a client for the Debug Adapter Protocol (DAP). Each language needs a **debug adapter** — a bridge between Neovim and the language runtime:

```
Neovim (nvim-dap) ←→ Debug Adapter ←→ Language Runtime
```

You install the adapter once per language, configure it in `init.lua`, then use the same keybinds for everything.

---

## Universal keybinds (all languages)

```
<leader>db   Toggle breakpoint
<leader>dB   Set conditional breakpoint
<leader>dc   Continue (start or resume)
<leader>ds   Step over (next line, don't enter function)
<leader>di   Step into (enter function call)
<leader>do   Step out (finish current function, return to caller)
<leader>dr   Open REPL (evaluate expressions interactively)
<leader>dK   Hover variable value under cursor
<leader>dl   Re-run last debug session
<leader>dt   Terminate session
```

Add to `init.lua`:

```lua
local dap = require('dap')
local map = vim.keymap.set

map('n', '<leader>db', dap.toggle_breakpoint, { desc = 'Toggle breakpoint' })
map('n', '<leader>dB', function()
  dap.set_breakpoint(vim.fn.input('Condition: '))
end, { desc = 'Conditional breakpoint' })
map('n', '<leader>dc', dap.continue,      { desc = 'Continue' })
map('n', '<leader>ds', dap.step_over,     { desc = 'Step over' })
map('n', '<leader>di', dap.step_into,     { desc = 'Step into' })
map('n', '<leader>do', dap.step_out,      { desc = 'Step out' })
map('n', '<leader>dr', dap.repl.open,     { desc = 'Open REPL' })
map('n', '<leader>dK', require('dap.ui.widgets').hover, { desc = 'Hover value' })
map('n', '<leader>dt', dap.terminate,     { desc = 'Terminate' })
```

---

## Go

### Install adapter (delve)

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
dlv version   # verify
```

### Configure in init.lua

```lua
local dap = require('dap')

dap.adapters.go = {
  type = 'server',
  port = '${port}',
  executable = {
    command = vim.fn.exepath('dlv'),
    args = { 'dap', '-l', '127.0.0.1:${port}' },
  },
}

dap.configurations.go = {
  {
    type = 'go',
    name = 'Debug package',
    request = 'launch',
    program = '${fileDirname}',
  },
  {
    type = 'go',
    name = 'Debug test',
    request = 'launch',
    mode = 'test',
    program = '${fileDirname}',
  },
  {
    type = 'go',
    name = 'Debug test (specific)',
    request = 'launch',
    mode = 'test',
    program = '${fileDirname}',
    args = { '-test.run', '${input:testName}' },
  },
}
```

### Debug flow

```bash
cd examples/go-microservice
nvim handler_test.go

# 1. Set breakpoint inside TestHealth with <leader>db
# 2. Start debug session: <leader>dc → choose "Debug test"
# 3. Step into the handler: <leader>di
# 4. Hover over 'w' to see response recorder value: <leader>dK
# 5. Open REPL and evaluate: <leader>dr → type: w.Code
# 6. Continue to end: <leader>dc
# 7. Terminate: <leader>dt
```

**Tip:** Use `nvim-dap-go` plugin for a simpler setup:

```lua
{ 'leoluz/nvim-dap-go', config = true }
-- Then use :DapGoDebugTest to run the test under cursor
```

---

## JavaScript / TypeScript

### Install adapter

```bash
# Option A: vscode-js-debug (recommended, supports Node + browser)
mkdir -p ~/.local/share/nvim/dap/vscode-js-debug
cd ~/.local/share/nvim/dap/vscode-js-debug
git clone https://github.com/microsoft/vscode-js-debug .
npm install
npm run compile

# Option B: simpler — js-debug via mason
# :MasonInstall js-debug-adapter   (if using mason.nvim)
```

### Configure in init.lua

```lua
require('dap-vscode-js').setup({
  debugger_path = vim.fn.expand('~/.local/share/nvim/dap/vscode-js-debug'),
  adapters = { 'pwa-node', 'pwa-chrome' },
})

local js_config = {
  {
    type = 'pwa-node',
    request = 'launch',
    name = 'Launch Node file',
    program = '${file}',
    cwd = '${workspaceFolder}',
  },
  {
    type = 'pwa-node',
    request = 'attach',
    name = 'Attach to process',
    processId = require('dap.utils').pick_process,
    cwd = '${workspaceFolder}',
  },
}

local dap = require('dap')
dap.configurations.javascript = js_config
dap.configurations.typescript = js_config
```

### Debug flow (Node.js)

```bash
# Start node with inspector in one pane
node --inspect src/index.js

# In another pane, open nvim
nvim src/index.js

# 1. Set breakpoint: <leader>db
# 2. Start session: <leader>dc → choose "Attach to process" → pick Node PID
# 3. Step through: <leader>ds / <leader>di
# 4. Hover a variable: <leader>dK
```

### Debug flow (TypeScript)

```bash
# Compile with source maps
tsc --sourceMap

# Or use ts-node with inspect
node --inspect -r ts-node/register src/index.ts

# Then attach from nvim the same way as Node above
```

**Tip:** Add `"sourceMap": true` to `tsconfig.json` so breakpoints map correctly to `.ts` lines.

---

## Swift

### Install adapter (codelldb)

```bash
# macOS — codelldb ships with Xcode command line tools
xcode-select --install

# Or install codelldb standalone
brew install llvm

# Verify
lldb --version
```

### Configure in init.lua

```lua
local dap = require('dap')

dap.adapters.swift = {
  type = 'executable',
  command = 'lldb-dap',   -- shipped with Xcode / LLVM
  -- fallback: 'lldb-vscode' on older Xcode
}

dap.configurations.swift = {
  {
    type = 'swift',
    name = 'Launch Swift executable',
    request = 'launch',
    program = '${workspaceFolder}/.build/debug/${workspaceFolderBasename}',
    cwd = '${workspaceFolder}',
    args = {},
  },
}
```

### Debug flow

```bash
# Build with debug symbols
swift build

# Open source file in nvim
nvim Sources/MyApp/main.swift

# 1. Set breakpoint: <leader>db
# 2. Start: <leader>dc → "Launch Swift executable"
# 3. Step through: <leader>ds / <leader>di
# 4. Inspect variables: <leader>dK or <leader>dr → type variable name
```

**Tip:** Use the `swift-lldb` plugin for Swift-specific variable formatters:

```lua
{ 'wojciech-kulik/xcodebuild.nvim' }  -- full Xcode workflow in nvim
```

---

## Kotlin

### Install adapter (kotlin-debug-adapter)

```bash
# Download KDA from GitHub releases
mkdir -p ~/.local/share/nvim/dap/kotlin-debug-adapter
cd ~/.local/share/nvim/dap/kotlin-debug-adapter
curl -L https://github.com/fwcd/kotlin-debug-adapter/releases/latest/download/adapter.zip -o adapter.zip
unzip adapter.zip

# Verify
ls bin/kotlin-debug-adapter
```

### Configure in init.lua

```lua
local dap = require('dap')

dap.adapters.kotlin = {
  type = 'executable',
  command = vim.fn.expand('~/.local/share/nvim/dap/kotlin-debug-adapter/bin/kotlin-debug-adapter'),
}

dap.configurations.kotlin = {
  {
    type = 'kotlin',
    name = 'Launch Kotlin main',
    request = 'launch',
    projectRoot = '${workspaceFolder}',
    mainClass = function()
      return vim.fn.input('Main class (e.g. com.example.MainKt): ')
    end,
  },
}
```

### Debug flow

```bash
# Build with Gradle (debug symbols included by default)
./gradlew build

# Open a Kotlin file
nvim src/main/kotlin/com/example/Main.kt

# 1. Set breakpoint: <leader>db
# 2. Start: <leader>dc → enter main class (e.g. com.example.MainKt)
# 3. Step through: <leader>ds / <leader>di
# 4. Inspect: <leader>dK
```

---

## Java

### Install adapter (java-debug via nvim-jdtls)

The cleanest Java debugging in Neovim uses `jdtls` (Eclipse JDT Language Server) with the `java-debug` extension:

```bash
# Install jdtls (the Java LSP + debug host)
# Option A: via Mason (recommended)
# In nvim: :MasonInstall jdtls java-debug-adapter

# Option B: manually
mkdir -p ~/.local/share/nvim/dap/java-debug
cd ~/.local/share/nvim/dap/java-debug
git clone https://github.com/microsoft/java-debug .
./mvnw clean install -q
```

### Configure in init.lua

```lua
-- jdtls handles both LSP and DAP for Java
-- This goes in a FileType autocommand for Java files:
vim.api.nvim_create_autocmd('FileType', {
  pattern = 'java',
  callback = function()
    local jdtls = require('jdtls')
    local bundles = {
      vim.fn.glob(
        vim.fn.expand('~/.local/share/nvim/dap/java-debug') ..
        '/com.microsoft.java.debug.plugin/target/com.microsoft.java.debug.plugin-*.jar'
      )
    }

    jdtls.start_or_attach({
      cmd = { 'jdtls' },
      root_dir = vim.fs.dirname(
        vim.fs.find({ 'gradlew', 'pom.xml', '.git' }, { upward = true })[1]
      ),
      init_options = { bundles = bundles },
      on_attach = function(_, _)
        jdtls.setup_dap({ hotcodereplace = 'auto' })
      end,
    })
  end,
})
```

Add to plugin list:

```lua
{ 'mfussenegger/nvim-jdtls' }
```

### Debug flow

```bash
# Compile (Maven or Gradle)
./mvnw compile -q
# or
./gradlew compileJava -q

# Open a Java file
nvim src/main/java/com/example/App.java

# 1. Set breakpoint: <leader>db
# 2. Start: <leader>dc → "Launch Java" (jdtls provides the config)
# 3. Step through: <leader>ds / <leader>di
# 4. Inspect: <leader>dK
# 5. Hot reload (jdtls supports it): edit code while paused, resume
```

---

## 4. nvim-dap-ui (recommended)

Add a visual debugger UI with variable panels, call stack, and watch expressions:

```lua
{
  'rcarriga/nvim-dap-ui',
  dependencies = { 'mfussenegger/nvim-dap', 'nvim-neotest/nvim-nio' },
  config = function()
    local dapui = require('dapui')
    dapui.setup()
    -- Auto-open/close UI with debug sessions
    require('dap').listeners.after.event_initialized['dapui_config'] = dapui.open
    require('dap').listeners.before.event_terminated['dapui_config'] = dapui.close
  end,
}
```

Once installed, the UI opens automatically when a debug session starts, showing:
- Variables panel (locals, upvalues, globals)
- Call stack
- Breakpoints list
- Watch expressions
- REPL

---

## Exercise 1 — Debug the Go health handler

```bash
cd examples/go-microservice
nvim handler_test.go

# Set breakpoint on the response assertion line
# <leader>db

# Start debug session
# <leader>dc → "Debug test"

# Step into the handler
# <leader>di

# Hover over 'w' (response recorder)
# <leader>dK

# Check status code in REPL
# <leader>dr → w.Code
```

## Exercise 2 — Find a bug with the debugger

```bash
claude "Introduce a subtle bug in main.go — wrong status code under a specific condition. Don't explain what it is."
go test ./...  # fails

# Now: open nvim, set breakpoints, step through, find it without reading the code top-to-bottom
```

## Exercise 3 — Cross-language: debug a Node.js script

```bash
cat > /tmp/debug-me.js << 'EOF'
function add(a, b) { return a - b; }  // bug: should be +
console.log(add(2, 3));  // expected 5, got -1
EOF

node --inspect /tmp/debug-me.js &
nvim /tmp/debug-me.js
# Attach, set breakpoint inside add(), step in, hover a and b
```

---

## Resources

- [nvim-dap GitHub](https://github.com/mfussenegger/nvim-dap)
- [nvim-dap-ui](https://github.com/rcarriga/nvim-dap-ui)
- [nvim-dap-go](https://github.com/leoluz/nvim-dap-go)
- [vscode-js-debug](https://github.com/microsoft/vscode-js-debug)
- [kotlin-debug-adapter](https://github.com/fwcd/kotlin-debug-adapter)
- [nvim-jdtls](https://github.com/mfussenegger/nvim-jdtls)
- [delve documentation](https://github.com/go-delve/delve/tree/master/Documentation)

---

## Pass gate

You can set a breakpoint, start a debug session for **at least two languages**, step into a function, inspect a variable's value in the hover UI, and identify the root cause of a bug — without adding a single print statement.