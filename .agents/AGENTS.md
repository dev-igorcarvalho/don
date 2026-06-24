# Don Monorepo Guidelines & Rulebook

Welcome, Agent. This document outlines the architectural context, technologies, and development standards for the **Don** monorepo. Consult this file before proposing or making any changes.

---

## 1. Project Overview & Architecture

**Don** is an agentic orchestration platform consisting of Go services, CLI tools, and scaffolding frameworks. The monorepo is divided into three main components:

### 📂 [consiglieri](file:///home/igor/Documents/projetos/don/consiglieri)
A Go-based framework for orchestrating AI agents and task pipelines using various LLM CLI backends (Claude Code, Gemini CLI, Agy, etc.).
*   **Key Concepts:**
    *   **Agent / AgentProvider:** Atomic workers and their provider abstraction layer.
    *   **Pipeline:** Sequential execution functions mapping out workflow stages.
    *   **Orchestrator:** Top-level workflow manager coordinating multiple pipelines.
    *   **Session:** Environment isolation, logging, and shared context state.
    *   **TUI Dashboard:** The terminal UI for discovering and launching registered workflows.

### 📂 [padrino](file:///home/igor/Documents/projetos/don/padrino)
A Go application structured around **Hexagonal Architecture (Ports & Adapters)**, scaffolded to support a REST API, CLI, and background workers.
*   **Layer Boundaries:**
    *   `internal/core/domain/`: Pure business domain models and rules. **Must have zero dependencies** on other internal packages.
    *   `internal/core/usecases/`: Application-specific business rules.
    *   `internal/adapters/`: Infrastructure adapters (HTTP server using Echo, database layers, messaging).
    *   `internal/handlers/`: Transport controllers (HTTP controllers, etc.).

### 📂 [artefatti](file:///home/igor/Documents/projetos/don/artefatti)
A Node/npm-based scaffolding package (`don-artefatti`) used to install and manage the `.artefatti` workspace structure for pipeline orchestration.

---

## 2. Technology Stack & Requirements

*   **Go Version:** `1.26` or later (as defined in each `go.mod`).
*   **Node.js Version:** `>= 14.0.0` (for `artefatti`).
*   **External CLI Providers:** Installed LLM CLIs (`claude`, `gemini`, `agy`) for orchestrator drivers.
*   **Linter:** `golangci-lint` (rules configured at [/.golangci.yml](file:///home/igor/Documents/projetos/don/.golangci.yml) and [/consiglieri/.golangci.yml](file:///home/igor/Documents/projetos/don/consiglieri/.golangci.yml)).

---

## 3. Strict Coding Conventions

### Go Guidelines (Core Projects)
All Go development must abide by standard Go idioms and the active `go-standards` skill.

> [!IMPORTANT]
> **Consumer-Side Interfaces (CRITICAL)**
> *   **NEVER** define interfaces in the package that implements them (producer-side).
> *   **ALWAYS** define interfaces in the package that consumes the behavior (consumer-side).
> *   This is a non-negotiable standard in this project to satisfy the Interface Segregation Principle.

*   **Hexagonal Isolation:** Ensure all business domain layers in `padrino` are kept completely pure under `internal/core`. Dependency inversion must be strictly maintained via interfaces.
*   **Logging:** Use `log/slog` (or the wrapper in `pkg/logger`). Pass `context.Context` through all calls to preserve trace correlation.
*   **Errors:** Use explicit error handling and wrap error chains with context (e.g., `fmt.Errorf("failed to execute agent step: %w", err)`).
*   **consiglieri Orchestration & Primitives:**
    *   Use `Agent[T]` for executing atomic tasks, ensuring custom target types implement `FailureChecker` (`Failure() error`) to catch semantic errors.
    *   Coordinate multi-step tasks using `Pipeline` and `Orchestrator` components, maintaining strict sequence validation.
    *   Propagate `context.Context` fully to retrieve the session logger using `Logger(ctx)`.
    *   Ensure helper process test patterns (`GO_WANT_HELPER_PROCESS`) are used for command execution mock stubs.
    *   Strictly preserve package boundaries: `pkg/` directories must have zero dependencies on any `internal/` packages.

### Node/JS Guidelines
*   Keep the CLI (`bin/cli.js`) zero-dependency where possible, relying on standard Node APIs.
*   Adhere to strict semantic versions in `package.json`.

---

## 4. Operational Workflows & Commands

### Monorepo Executions
Run commands in their respective package directories:

| Component | Target Action | Command |
| :--- | :--- | :--- |
| **consiglieri** | Build | `go build ./...` |
| | Test | `go test ./pkg/primitives/...` |
| | Lint | `golangci-lint run ./...` |
| **padrino** | Run API | `go run cmd/api/main.go` |
| | Build | `go build -o bin/api ./cmd/api` |
| | Test | `go test ./...` |
| | Lint | `golangci-lint run ./...` |

---

## 5. Development Guardrails for AI Agents

*   **No Unauthorized Actions:** NEVER commit any changes or execute destructive actions without presenting a plan and getting explicit user approval.
*   **No Commits under .caporegime:** NEVER stage or commit any files or directories inside or under `.caporegime` (such as `.caporegime/workflows/*` or `.caporegime/session/*`). These files are reserved for local run configurations, workflow files, and session logs and must remain untracked.
*   **Testing First (TDD):** Accompany every bug fix or feature implementation with test cases (table-driven tests are preferred where applicable).
*   **Semantic Commits:** Commit messages must strictly follow the conventional commits specification (e.g., `feat(api): add auth filter`, `fix(TUI): adjust layout boundaries`). Use the `semantic-commit` skill to run commits.
*   **Contextual Awareness:** Before making changes to shared packages in `pkg/`, verify if the changes impact multiple downstream services; run all relevant test suites before finalizing the PR.
