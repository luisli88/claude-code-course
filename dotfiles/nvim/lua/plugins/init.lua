-- Custom plugins beyond what LazyVim extras provide.
-- Enable extras the proper way: run :LazyExtras inside nvim and toggle:
--   dap.core   → full DAP UI, keybinds, mason-nvim-dap
--   lang.go    → gopls, delve, goimports, golangci-lint

return {
  -- nvim-dap has no setup() function; lang extras (go, java, typescript) add adapters
  -- via opts functions. Without this stub, Lazy.nvim tries to call dap.setup() → crash.
  -- dap.core LazyExtra also fixes this, but this stub works without it.
  {
    "mfussenegger/nvim-dap",
    config = function() end,
  },

  -- Extra Treesitter parsers not included by default
  {
    "nvim-treesitter/nvim-treesitter",
    opts = function(_, opts)
      vim.list_extend(opts.ensure_installed, { "tsx" })
    end,
  },
}
