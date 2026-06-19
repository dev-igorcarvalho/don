---
title: Architecture & Project Overview
description: High-level overview of Don's Hexagonal Architecture and directory structure.
last_updated: 2026-04-26
type: documentation
---

# Architecture & Project Overview

**Don** is a Go-based application designed with **Hexagonal Architecture** (Ports and Adapters). It is scaffolded to support multiple entry points, including a REST API, a CLI, and a background worker.

## Main Technologies
- **Language:** Go (Targeting 1.26+)
- **Web Framework:** Echo framework
- **Logging:** `slog` (structured logging)
- **Configuration:** `Netflix/go-env`

## Directory Structure
- `cmd/`: Entry points for different runtimes (API, CLI, Worker).
- `internal/core/domain/`: Pure business entities and logic.
- `internal/core/usecases/`: Application-specific business rules.
- `internal/adapters/`: Implementations of external interfaces (e.g., `HTTPServer` using Echo, DB, Messaging).
- `internal/handlers/`: Transport layer implementations (e.g., HTTP controllers).
- `internal/config/`: Application-specific configuration structures.
- `pkg/`: Shared utility packages (e.g., `logger`, `config` loader).

## Hexagonal Architecture Rules
- Keep business logic in `internal/core`.
- Ensure `domain` has no dependencies on other internal packages.
- Use interfaces for dependency inversion between `usecases` and `adapters`.
