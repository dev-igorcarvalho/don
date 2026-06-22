# Don - Artefatti Package Guidelines & Rulebook

Welcome, Agent. This document outlines the architectural context, development standards, and conventions for the **artefatti** subproject within the **Don** monorepo. Consult this file before proposing or making any changes in this directory.

---

## 1. Project Overview

The `artefatti` package (`don-artefatti`) is a Node/npm-based scaffolding package. Its sole purpose is to install and configure the `.artefatti` workspace structure inside a target directory to support agentic pipeline orchestration.

### Folder Structure
*   [bin/cli.js](file:///home/igor/Documents/projetos/don/artefatti/bin/cli.js) — The CLI execution entrypoint that handles directory creation and asset copying.
*   [assets/](file:///home/igor/Documents/projetos/don/artefatti/assets) — Template assets distributed by the package:
    *   [assets/README.md](file:///home/igor/Documents/projetos/don/artefatti/assets/README.md) — Explains the structure and DSL to developers in their generated workspace.
    *   [assets/rules/pipeline-workflow-rules.md](file:///home/igor/Documents/projetos/don/artefatti/assets/rules/pipeline-workflow-rules.md) — Reference rules for agents run on target projects.
    *   [assets/rules/workflow.md](file:///home/igor/Documents/projetos/don/artefatti/assets/rules/workflow.md) — The authoritative DSL spec document.

---

## 2. Technology Stack & Requirements

*   **Node.js Version:** `>= 14.0.0`.
*   **Dependencies:** Must remain **zero-dependency**. Rely only on Node.js core modules (`fs`, `path`, etc.).

---

## 3. Strict Coding Conventions & Rules

### Zero-Dependency Constraint
*   **Do NOT** add any external dependencies to `package.json` unless explicitly authorized by Don Igor Dal Bosco.
*   Use synchronous fs methods (`fs.existsSync`, `fs.mkdirSync`, `fs.copyFileSync`) inside the CLI for straightforward execution flow control.

### Thematic Output (The Family Style)
*   The CLI prints thematic, Godfather-inspired Italian messages on success. Keep these phrases intact and maintain this flavor for any new interactive or error messages (e.g., "🌹 Salutiamo il nostro nuovo amico!").

### Asset Synchronization
*   If changing assets in `assets/`, ensure the copy routines inside [bin/cli.js](file:///home/igor/Documents/projetos/don/artefatti/bin/cli.js) are aligned (e.g., if a new rule file or template is added).
*   Always ensure target directory paths are relative to `process.cwd()` and assets are relative to `__dirname`.

---

## 4. Operational Workflows & Commands

Run CLI executions or local link operations in this directory:

| Action | Command |
| :--- | :--- |
| **Run CLI locally** | `node bin/cli.js` |
| **Link package globally** | `npm link` |
| **Test scaffold run** | `mkdir -p test-scaffold && cd test-scaffold && node ../bin/cli.js` |

---

## 5. Development Guardrails

*   **No Unauthorized Actions:** NEVER commit any changes or execute destructive actions without presenting a plan and getting explicit user approval.
*   **Backward Compatibility:** Ensure changes to templates do not break standard workspace structure for projects utilizing the orchestration pipeline.
*   **Semantic Commits:** Commit messages must strictly follow the conventional commits specification (e.g., `feat(artefatti): add rule template`, `fix(artefatti): resolve cli crash on missing target dir`).
