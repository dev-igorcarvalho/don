# Padrino Guidelines & Rulebook

Welcome, Agent. This document defines the specific architectural guidelines, technology stack, coding standards, and workflows for **Padrino**, a Go application structured around Hexagonal Architecture.

Consult this rulebook before drafting plans, making code edits, or running verification tasks.

---

## 1. Project Overview & Architecture

**Padrino** is a Go-based core service designed with **Hexagonal Architecture (Ports and Adapters)**. It supports multiple entry points, including a REST API, a CLI utility, and background worker runners.

### Directory Structure & Architectural Layers

*   `cmd/`: Entry points for different runtimes (e.g., `cmd/api/main.go`, `cmd/cli/main.go`).
*   `internal/core/domain/`: Pure business models and logic.
    *   **Rule:** This package **MUST NOT** import any other internal package or external frameworks. It must remain a pure Go representation of the business domain.
*   `internal/core/usecases/`: Application-specific business rules that orchestrate the flow of data to and from domain models.
*   `internal/adapters/`: Infrastructure adapters implementing external interfaces (e.g., Echo HTTP server, database adapters, messaging channels).
*   `internal/handlers/`: Transport controllers (e.g., HTTP controllers, CLI command dispatchers).
*   `internal/config/`: Configuration parsing and schemas.
*   `pkg/`: Shared utilities (e.g., custom configuration loader, logging utilities).

---

## 2. Technology Stack & Dependencies

*   **Runtime:** Go `1.26` or later.
*   **Web Framework:** Echo (`github.com/labstack/echo/v4`).
*   **Configuration Library:** Netflix env decoder (`github.com/Netflix/go-env`).
*   **Database & Migrations:** SQLite3 (`github.com/mattn/go-sqlite3`), and golang-migrate (`github.com/golang-migrate/migrate/v4`).
*   **Testing Frameworks:** Testify (`github.com/stretchr/testify`), SQL mock (`github.com/DATA-DOG/go-sqlmock`).

---

## 3. Strict Coding Conventions

### A. Consumer-Side Interfaces (CRITICAL)
*   **Rule:** **NEVER** define interfaces in the package that implements them (producer-side).
*   **Rule:** **ALWAYS** define interfaces in the package that consumes the behavior (consumer-side).
*   *Rationale:* This ensures loose coupling, promotes interface segregation, and prevents circular imports across layers.

### B. Structured Logging
*   Always use the custom logger package in `pkg/logger` (which wraps `log/slog`).
*   Always pass `context.Context` to all logging calls. This is essential for trace and correlation ID propagation.

### C. Configuration Management
*   Add configuration parameters in [internal/config/types.go](file:///home/igor/Documents/projetos/don/padrino/internal/config/types.go) utilizing `env` tags.
*   Load configurations using `pkg/config`.

### D. Error Handling
*   Handle errors explicitly and wrap them with domain context (e.g., `fmt.Errorf("failed to retrieve customer records: %w", err)`).
*   Use the custom `must` initializer helper inside `main.go` only for critical bootstrap-level startup failures.

---

## 4. Run, Build & Verification Commands

Execute the following commands within the [/padrino](file:///home/igor/Documents/projetos/don/padrino) subdirectory:

| Purpose | Target Action | Command |
| :--- | :--- | :--- |
| **Run API** | Bootstraps local API server | `go run cmd/api/main.go` |
| **Build** | Builds application binary | `go build -o bin/api ./cmd/api` |
| **Test** | Runs package test suite | `go test ./...` |
| **Lint** | Runs standard Go linter checks | `golangci-lint run ./...` |

---

## 5. Agentic Workflow Expectations

When implementing a task in **Padrino**, you must operate through the following workflow states:

1.  **Planning Phase:** Formulate a plan mapping target file dependencies. Generate a manifest of files to modify. Let the human review and approve it.
2.  **Implementation Phase:**
    *   Activate the `go-standards` skill before editing code.
    *   Implement changes surgically, keeping domain models pure.
    *   Accompany all edits with table-driven tests.
3.  **QA Validation:**
    *   Run `go test ./...` and `golangci-lint run ./...` to ensure zero errors.
    *   If failures occur, analyze logs and fix changes in an iterative retry loop (up to 3 times before requesting help).
4.  **Finalization:**
    *   Run `/batch-parallel-frontmatter` or apply the `frontmatter-adder` skill on Go files.
    *   Commit changes using the `semantic-commit` skill, referencing the original issue scope.
