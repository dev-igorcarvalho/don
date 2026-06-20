# Don Workspace Agent Customizations (`.agents/`)

This directory contains workspace-level settings, rules, hooks, and skills that customize and guide the behavior of the **Antigravity (AGY)** AI agents within the Don monorepo.

---

## 📂 Directory Structure

```text
.agents/
├── AGENTS.md        # Symlink to the root AGENTS.md (project-scoped guidelines and rules)
├── hooks.json       # Triggers for lifecycle hooks
├── settings.json    # Workspace permissions, status line, and CLI preferences
├── hooks/           # Shell scripts executed by hooks.json
│   ├── lint.sh      # Code quality & validation script (runs formatting & linters)
│   ├── status-line.sh  # Dynamic terminal UI status bar script
│   └── stop-hook.sh # Cleans up and processes rulebook updates on session close
└── skills/          # Custom agent skills
    └── todo-tracker # Scans and documents todo: tags into issues.md
```

---

## ⚙️ Configuration Files

### 1. `settings.json`
Manages general configuration options for the Antigravity session in this workspace:
- **`permissions`**: Fine-tuned rules defining what files, commands, and folders are accessible.
- **`statusLine`**: Integrates the custom shell-based status bar showing quota and context limits.
- **`toolPermission`**: Set to `always-proceed` for streamlined local execution.

### 2. `hooks.json`
Maps agent lifecycle states (e.g., `SessionStart`, `BeforeAgent`, `AfterAgent`, `Stop`) to commands.

---

## 🚀 Lifecycle Hook Scripts (`hooks/`)

- **`status-line.sh`**: Executed to render the CLI status line. Fetches and displays token usage and context percentage.
- **`lint.sh`**: Run on `AfterAgent`. Automates Go formatting (`go fmt`), validation (`go vet`), security auditing (`gosec`), and full linting (`golangci-lint`) to guarantee codebase consistency.
- **`stop-hook.sh`**: Executed on agent `Stop` to summarize and structure rules updates in `AGENTS.md`.

---

## 🧠 Rules & Custom Skills

### `AGENTS.md` (Project Guidelines)
Synchronized with the root rulebook via a symbolic link, ensuring the developer can view it at the root, while the agent reads it directly from this customizations folder.

### `skills/todo-tracker`
Automates comment discovery and records them inside `issues.md`. Triggered when you ask to check pending items.
