# GEMINI.md - Project Context: Don

This document provides architectural context, development standards, and operational guidelines for the **Don** project.

## Project Overview
**Don** is a Go-based application designed with **Hexagonal Architecture** (Ports and Adapters). It is scaffolded to support multiple entry points, including a REST API, a CLI, and a background worker.

- **Main Technologies:** Go (Targeting 1.26+), Echo framework, `slog` for structured logging, `Netflix/go-env` for configuration.
- **Architecture:** 
  - `cmd/`: Entry points for different runtimes (API, CLI, Worker).
  - `internal/core/domain/`: Pure business entities and logic.
  - `internal/core/usecases/`: Application-specific business rules.
  - `internal/adapters/`: Implementations of external interfaces (e.g., `HTTPServer` using Echo, DB, Messaging).
  - `internal/handlers/`: Transport layer implementations (e.g., HTTP controllers).
  - `internal/config/`: Application-specific configuration structures.
  - `pkg/`: Shared utility packages (e.g., `logger`, `config` loader).

## Building and Running

### Prerequisites
- Go 1.26 or later (as specified in `go.mod`).
- Environment variables configured (see [Configuration](#configuration)).

### Key Commands
- **Run API:** `go run cmd/api/main.go`
- **Build:** `go build -o bin/api ./cmd/api`
- **Test:** `go test ./...`
- **Lint:** (TBD) `golangci-lint run` (standard recommendation)

### Configuration
The application uses environment variables for configuration. The following variables are required by `AppConfig`:

| Variable      | Description                                      |
|---------------|--------------------------------------------------|
| `NAME`        | Application name                                 |
| `VERSION`     | Application version                              |
| `ENVIRONMENT` | Deployment environment (e.g., `SANDBOX`, `PROD`) |
| `HTTP_PORT`    | Port for the HTTP server (default: 8080)         |

## Development Conventions

### 1. Hexagonal Architecture
- Keep business logic in `internal/core`.
- Ensure `domain` has no dependencies on other internal packages.
- Use interfaces for dependency inversion between `usecases` and `adapters`.

### 2. Structured Logging
- Use the custom `pkg/logger` package which wraps `log/slog`.
- Log levels are automatically adjusted based on the `ENVIRONMENT`:
  - `SANDBOX`: Debug level, Text output.
  - Others: Info level, JSON output.
- Always pass `context.Context` to logging functions to ensure trace/correlation IDs are captured.

### 3. Configuration Management
- Add new configuration parameters to `internal/config/types.go` using `env` tags.
- Use the generic loader in `pkg/config` to load configurations.
- Example usage in `main.go`: `pkgConfig.Load[config.AppConfig]()`.

### 4. Error Handling
- Use the `must` helper in `main.go` for critical initialization errors.
- Prefer explicit error handling and wrapping for business logic.

### 5. Consumer-Side Interfaces (CRITICAL)
- **NEVER** define interfaces in the package that implements them (producer-side).
- **ALWAYS** define interfaces in the package that consumes the behavior (consumer-side).
- This is a non-negotiable Go standard for this project to ensure loose coupling and satisfy the Interface Segregation Principle.

## 🧠 Knowledge Base
All architectural decisions, domain rules, and developer guides are maintained in the [Internal Wiki](docs/improvements/gemini/index.md).

## Gemini CLI Directives
- **No Unauthorized Commits:** NEVER, absolutely never, commit anything without explicit user permission for each commit.
- **Mandatory Planning:** ALWAYS create a plan, let the user review it, and only implement it after receiving explicit approval. Always ask first.
- **Go Standards:** Always activate the `go-standards` skill before performing any Go-related coding tasks to ensure compliance with project and language conventions.
- **Semantic Commits:** Always use the `semantic-commit` skill for all commits to maintain a consistent and structured commit history.
- **Post-Execution:** ALWAYS run the command `/batch-parallel-frontmatter` after a successful execution.
- **Consumer-Side Interfaces:** NEVER define interfaces in the producer package. Interfaces MUST live in the consumer package. This is a hard requirement for all Go code in this project.


---
*Note: This project is in its early stages. Many directories contain only `.gitkeep` files as placeholders for future implementation.*
