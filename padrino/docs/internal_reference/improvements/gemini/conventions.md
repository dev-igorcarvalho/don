---
title: Development Conventions
description: Coding standards, logging practices, and error handling rules for the project.
last_updated: 2026-04-26
type: standards
---

# Development Conventions

## 1. Structured Logging
- Use the custom `pkg/logger` package which wraps `log/slog`.
- Log levels are automatically adjusted based on the `ENVIRONMENT`:
  - `SANDBOX`: Debug level, Text output.
  - Others: Info level, JSON output.
- Always pass `context.Context` to logging functions to ensure trace/correlation IDs are captured.

## 2. Configuration Management
- Add new configuration parameters to `internal/config/types.go` using `env` tags.
- Use the generic loader in `pkg/config` to load configurations.
- Example: `pkgConfig.Load[config.AppConfig]()`.

## 3. Error Handling
- Use the `must` helper in `main.go` for critical initialization errors.
- Prefer explicit error handling and wrapping for business logic.

## 4. Testing and Harness Engineering
- Follow the patterns outlined in [docs/harness-engineering.md](../harness-engineering.md).
- Use `testify/suite` for complex setup and lifecycle management.
- Prefer in-memory adapters over complex mocks for domain logic.
