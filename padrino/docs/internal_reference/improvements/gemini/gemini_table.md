# Gemini Project Index: Don

This index serves as the entry point for project-specific instructions and architectural standards.

## Core Identity

**Don** is a Go-based application using **Hexagonal Architecture**. It supports API, CLI, and Worker runtimes.

| Category         | Description                                                                      | Document Link                                                                        |
|:-----------------|:---------------------------------------------------------------------------------|:-------------------------------------------------------------------------------------|
| **Architecture** | High-level overview of Don's Hexagonal Architecture and directory structure.     | [docs/gemini/architecture.md](architecture.md)                           |
| **Operations**   | Instructions for building, running, and configuring the Don application.         | [docs/gemini/ops.md](ops.md)                                             |
| **Conventions**  | Coding standards, logging practices, and error handling rules for the project.   | [docs/gemini/conventions.md](conventions.md)                             |
| **Directives**   | Mandatory operational guidelines and skill activations for the Gemini CLI agent. | [docs/gemini/directives.md](directives.md)                               |
| **Testing**      | Harness engineering and testing strategies.                                      | [docs/improvements/harness-engineering.md](../harness-engineering.md) |

---

### Quick Mandates for Gemini CLI

1. **Skill Check:** Activate `go-standards` for code and `semantic-commit` for Git.
2. **Context:** Business logic belongs in `internal/core`. `domain` is pure.
3. **Safety:** Never bypass type safety or ignore linting rules.
