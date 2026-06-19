---
title: Operations (Build, Run & Config)
description: Instructions for building, running, and configuring the Don application.
last_updated: 2026-04-26
type: operations
---

# Operations (Build, Run & Config)

## Building and Running

### Prerequisites
- Go 1.26 or later (as specified in `go.mod`).
- Environment variables configured.

### Key Commands
- **Run API:** `go run cmd/api/main.go`
- **Build:** `go build -o bin/api ./cmd/api`
- **Test:** `go test ./...`
- **Lint:** `golangci-lint run`

## Configuration
The application uses environment variables for configuration. Required by `AppConfig`:

| Variable      | Description                                      |
|---------------|--------------------------------------------------|
| `NAME`        | Application name                                 |
| `VERSION`        | Application version                              |
| `ENVIRONMENT` | Deployment environment (e.g., `SANDBOX`, `PROD`) |
| `HTTP_PORT`    | Port for the HTTP server (default: 8080)         |
