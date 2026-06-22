# Don Consiglieri

A Go-based framework for orchestrating AI agents and task pipelines using various LLM CLI providers.

## Project Overview

Don Consiglieri provides a structured way to define and run AI agents. It abstracts the interaction with LLM CLIs (like Claude Code, Gemini CLI, etc.) through a consistent interface, allowing for complex multi-step workflows, pipelines, and orchestration.

### Core Architecture (`pkg/primitives`)

- **Agent**: The atomic unit of work. It takes a prompt, uses a `AgentProvider` to execute it, and parses the result. Supports `Before` and `After` hooks.
- **AgentProvider**: Interface for different LLM backends. Implementations include `ClaudeProvider`, `GeminiProvider`, and `AgyProvider`.
- **Pipeline**: A sequence of operations that can include agent calls or standard Go functions.
- **Orchestrator**: Manages a collection of workflows (Pipelines) and executes them in sequence.
- **Session**: Manages context, logging, and workspace isolation for agent runs.

## Building and Running

### Prerequisites
- Go 1.26.1+
- [Optional] `golangci-lint` for code quality checks.
- Installed LLM CLIs (e.g., `claude`, `gemini`, `agy`).

### Commands
- **Build**: `go build ./...`
- **Test**: `go test ./pkg/primitives/...`
- **Lint**: `golangci-lint run ./...`

## Development Conventions

### Coding Style
- Follow standard Go idioms and `gofmt` formatting.
- Use explicit error handling.
- Maintain a clean separation between the `pkg` layer (primitives) and potential application logic.

### Testing
- Every new feature or bug fix must be accompanied by unit tests in a corresponding `_test.go` file.
- Use table-driven tests where appropriate.
- Mock external dependencies (like `exec.Command`) to keep tests fast and deterministic.

### Commits
- Use [Conventional Commits](https://www.conventionalcommits.org/) (e.g., `feat:`, `fix:`, `chore:`, `test:`).
- Group related changes (e.g., a feature and its tests) into a single commit.

### Documentation
- Keep `GEMINI.md` updated with major architectural changes or new development workflows.
- Use descriptive comments for exported symbols.


## Glossary

### Core Domain

- **Orchestrator**: The top-level manager of workflows. It coordinates the execution of multiple **Pipelines**.
- **Pipeline**: A structured sequence of operations. It encapsulates any logic (Agents, standard Go code, or both) within a single execution function (`fn`).
- **Session**: The execution scope for a single **Orchestrator** run. It provides environment isolation, centralized logging, and shared state that is passed down to all child operations.
- **Agent**: An atomic worker that performs a specific task. It uses an **AgentProvider** to interact with an LLM and can be configured with hooks.
- **AgentProvider**: An abstraction layer for LLM backends (e.g., Claude, Gemini, Agy), responsible for executing commands and returning results.

### User Interface

- **Dashboard**: The main TUI view for selecting and launching registered **Workflows**.
- **Workflow**: A named set of **Pipelines** (or a single Pipeline) managed by an **Orchestrator** that performs a high-level task. Workflows are discovered as Go source files within the `.agentic/workflows` directory and executed as standalone processes using `go run` by the TUI.
- **Execution Tab**: A TUI component that displays the real-time progress and logs of a running **Workflow** by capturing the output of its process.
