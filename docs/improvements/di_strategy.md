# Improvement: Manual Dependency Injection Strategy

## Objective
Implement a robust, compile-time safe Dependency Injection (DI) strategy inspired by Google Wire, but using pure Go primitives. This approach eliminates the need for external code generation while maintaining explicit dependency graphs and high testability.

## Problem Statement
Currently, dependencies are manually wired in the `cmd` or `adapters` layer. As the project grows:
1. `main.go` will become cluttered with complex initialization logic.
2. Circular dependencies may become harder to track.
3. Mocking deep dependency trees for integration tests will be cumbersome.

## Proposed Strategy: "The Container Pattern"

Instead of using a magic reflection-based container or a code-generator, we will use a dedicated `container` package that acts as the "composition root" for each entry point (API, CLI, Worker).

### 1. Structure
New package: `internal/infra/container`

This package will contain:
- A `Container` struct holding all long-lived application dependencies.
- Provider functions (Constructors) for each component.
- A `NewContainer(cfg)` function that wires everything together.

### 2. Implementation Example

```go
package container

import (
    "github.com/dev-igorcarvalho/don/internal/config"
    "github.com/dev-igorcarvalho/don/internal/handlers"
    "github.com/dev-igorcarvalho/don/internal/adapters"
    // ... other imports
)

type Container struct {
    HTTPServer    *adapters.HTTPServer
    HealthHandler *handlers.HealthHandler
    // ... Repositories, UseCases, Services
}

func NewContainer(cfg config.AppConfig) *Container {
    // 1. Initialize Handlers/Adapters (Leaf nodes)
    healthHandler := handlers.NewHealthHandler()

    // 2. Initialize UseCases (Business Logic)
    // exampleUC := usecases.NewExampleUseCase(someRepo)

    // 3. Initialize Top-level Adapters
    httpServer := adapters.NewHTTPServer(cfg, healthHandler)

    return &Container{
        HTTPServer:    httpServer,
        HealthHandler: healthHandler,
    }
}
```

### 3. Usage in `main.go`

The entry point becomes extremely lean and focused on orchestration:

```go
func run() error {
    cfg := must(pkgConfig.Load[config.AppConfig]())
    logger.Setup(logger.Environment(cfg.Environment))

    // Initialize the dependency graph
    container := container.NewContainer(*cfg)

    // Start the server
    return container.HTTPServer.Start()
}
```

## Benefits
1. **Compile-time Safety**: Errors in wiring are caught by the compiler, not at runtime.
2. **Explicit Dependency Graph**: The `NewContainer` function serves as documentation for how the system is wired.
3. **Testability**: We can easily create a `NewTestContainer` or modify the `Container` struct in tests to inject mocks.
4. **No External Magic**: No dependencies on `google/wire` or reflection-based libraries like `dig` or `fx`.
5. **Architectural Alignment**: Fits perfectly with the Hexagonal Architecture by keeping the composition root outside of the core business logic.

## Next Steps
1. Create `internal/infra/container/container.go`.
2. Move the wiring logic from `internal/adapters/http_server.go` (constructor part) to the container.
3. Refactor `cmd/api/main.go` to use the new container.
