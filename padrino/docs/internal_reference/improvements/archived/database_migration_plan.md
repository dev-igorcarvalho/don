# Database Migration and Seeding System Proposal

This document outlines the proposed architecture for implementing database migrations and seeds in the **Don** project.

## 1. Migration Tool: `golang-migrate/migrate`

We propose using `golang-migrate/migrate` for managing database schema versions. It is an industry-standard, robust tool that ensures database state integrity.

### Key Features:
- **Strict State Management**: Fails fast if a migration is out of sync or fails, requiring explicit intervention.
- **Programmatic API**: Can be easily integrated into the Go application.
- **Support for Multiple Drivers**: Supports PostgreSQL, MySQL, SQLite, and more.
- **Embedding Support**: Allows embedding `.sql` files directly into the compiled Go binary using `embed.FS`.

## 2. Proposed Structure

We will treat migrations as a first-class operational concern, decoupled from generic database handlers. This follows Uber/Hashicorp standards of separating state management from data access.

```text
pkg/migrator/
├── migrator.go        # Wrapper for golang-migrate logic
└── seeder.go          # SQL script executor for seeding

scripts/database/
├── migrations/        # Directory for .up.sql and .down.sql files
└── seeds/             # Directory for seeding .sql scripts
```

## 3. Implementation Details

To ensure the system is engine-agnostic and fully decoupled, we offer two architectural options. Both options **strictly avoid `init()` functions** and global state, ensuring all dependencies are explicit.



###  Explicit Registry (Best for Multi-Driver Support)

This approach uses a `Migrator` instance that must be explicitly initialized and wired. There are **no `init()` side effects**.

**Pros:** Centralizes driver logic if multiple drivers are needed. High visibility of supported engines.
**Cons:** Requires an extra "wiring" step during application startup.

```go
// pkg/migrator/migrator.go
package migrator

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//@ @CReate funcs for the main drives, mysql, slqlite, postgres, etc
//@ Agreed. We should provide pre-defined factory helpers in a `factories.go` file 
//@ (e.g., `NewPostgresFactory`, `NewMySQLFactory`) to make the setup boilerplate-free 
//@ while maintaining the decoupling.
type DriverFactory func(db *sql.DB) (database.Driver, error)

type Migrator struct {
	factories map[string]DriverFactory
}

// New initializes a migrator with a map of explicit driver factories.
func New(factories map[string]DriverFactory) *Migrator {
	return &Migrator{factories: factories}
}

//go:embed ../../scripts/migrations/*.sql
var migrationFiles embed.FS

//@cant this recevie a DriverFactory and a cfg with dbname and sql folders ?
//@ Yes, that would make the `Migrator` purely functional/stateless. 
//@ We could define a `MigrationOptions` struct to group these parameters.
//@ However, the current design uses the `Migrator` struct to "bake in" the supported 
//@ drivers once, simplifying the call site so it only needs to provide the runtime 
//@ context (db, driver name, db name). Both are valid; the choice depends on 
//@ whether we want to manage the "registry" of drivers centrally or at the call site.
func (m *Migrator) MigrateUp(db *sql.DB, driverName, dbName string) error {
	factory, ok := m.factories[driverName]
	if !ok {
		return fmt.Errorf("migrator: unsupported driver %q", driverName)
	}

	dbDriver, err := factory(db)
	if err != nil {
		return fmt.Errorf("migrator: failed to initialize %s driver: %w", driverName, err)
	}

	//@ this path should be a input
	//@ Agreed. Hardcoding relative paths is fragile. We should either pass this as an argument 
	//@ to `MigrateUp` or include a `MigrationsPath` field in the `Migrator` struct during initialization.
	sourceDriver, err := iofs.New(migrationFiles, "../../scripts/database/migrations")
	if err != nil {
		return fmt.Errorf("migrator: failed to create source driver: %w", err)
	}

	mig, err := migrate.NewWithInstance("iofs", sourceDriver, dbName, dbDriver)
	if err != nil {
		return fmt.Errorf("migrator: failed to initialize: %w", err)
	}

	if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrator: up failed: %w", err)
	}

	return nil
}
```

### Seeding Mechanism (`seeder.go`)

The seeder remains decoupled as it uses the generic `*sql.DB` interface.

```go
package migrator

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
)

//go:embed ../../scripts/seeds/*.sql
var seedFiles embed.FS

func RunSeeds(ctx context.Context, db *sql.DB) error {
	entries, err := fs.ReadDir(seedFiles, "../../scripts/seeds")
	if err != nil {
		return fmt.Errorf("seeder: failed to read directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	for _, file := range files {
		content, err := fs.ReadFile(seedFiles, "../../scripts/seeds/"+file)
		if err != nil {
			return fmt.Errorf("seeder: failed to read %s: %w", file, err)
		}

		if _, err := db.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("seeder: execution failed for %s: %w", file, err)
		}
	}
	return nil
}
```

## 4. Operational Strategies: Startup vs. Dedicated Entry Point

Choosing when to run migrations is a critical architectural decision.

### Comparison Table

| Feature | Auto-Migrate at Startup | Dedicated Migrator (CI/CD / InitContainer) |
| :--- | :--- | :--- |
| **Use Case** | Local Development, Sandbox, Small Monoliths. | Production, High-Availability, Microservices. |
| **Simplicity** | High. One binary handles everything. | Medium. Requires extra deployment orchestration. |
| **Safety** | Low. Risk of race conditions if scaling pods. | High. Guaranteed to run once before app boots. |
| **Security** | Low. API needs high DB privileges (DDL). | High. API can have restricted DML-only privileges. |
| **Visibility** | Low. Migration failures crash the app. | High. Migration failures stop the deployment pipeline. |

### Recommended Hybrid Approach (Explicit Wiring)

1.  **Local/Sandbox:** Run migrations automatically in `main.go`.
2.  **Production:** Run migrations as a separate task during the deployment pipeline.

```go
// Example implementation in cmd/api/main.go (Using Option A)
import (
    "github.com/dev-igorcarvalho/don/pkg/migrator"
    "github.com/golang-migrate/migrate/v4/database/postgres"
)

func run() error {
    // ... setup db ...
    if cfg.Environment == "SANDBOX" {
        // Explicitly create driver and run migration
        driver, err := postgres.WithInstance(db, &postgres.Config{})
        if err != nil { return err }
        
        if err := migrator.MigrateUp(driver, "don"); err != nil {
            return fmt.Errorf("auto-migration failed: %w", err)
        }
    }
    // ... start server ...
}
```

## 5. Operational Tooling (Make)

For now, we will rely exclusively on `make` for developer operations rather than complicating the application CLI.

```makefile
# Makefile additions
.PHONY: db-migrate-up db-migrate-down db-migrate-create db-seed

DB_DSN ?= "postgres://user:password@localhost:5432/don?sslmode=disable"
MIGRATE_CMD=migrate -path scripts/migrations -database $(DB_DSN)

db-migrate-create: ## Create a new migration file (usage: make db-migrate-create name=init)
	@if [ -z "$(name)" ]; then echo "Error: name is required"; exit 1; fi
	migrate create -ext sql -dir scripts/migrations -seq $(name)

db-migrate-up: ## Run all up migrations
	$(MIGRATE_CMD) up

db-migrate-down: ## Run all down migrations
	$(MIGRATE_CMD) down

db-seed: ## Execute database seeds via ad-hoc run
	go run cmd/cli/main.go db seed # (or equivalent seeder entrypoint)
```

## 6. Advanced Operational Recommendations (Uber/Hashicorp Standards)

To achieve industry-leading operational excellence:
- **Dedicated Migrator Binary**: Instead of embedding migrations into the main API/Worker, compile a completely standalone `cmd/migrator` binary. This allows deploying a lightweight init-container in Kubernetes.
- **Idempotency & Upserts**: Seeds must be strictly idempotent (e.g., using `INSERT ... ON CONFLICT DO NOTHING`).
- **Distributed Locking**: For multi-instance deployments, utilize PostgreSQL advisory locks.
- **Telemetry**: Add tracing/metrics to the migrator to measure the duration of specific schema changes.

## 7. Implementation Plan

1. **Repository Setup**: Create `scripts/migrations/` and `scripts/seeds/`.
2. **Dependency**: `go get github.com/golang-migrate/migrate/v4`.
3. **Core Logic**: Create `pkg/migrator` implementing `migrator.go` and `seeder.go`.
4. **Tooling**: Add `db-migrate-*` and `db-seed` targets to the `Makefile`.
5. **Entrypoint**: Implement a lightweight `cmd/migrator` for executing the seeding logic.
6. **Testing**: Write integration tests using a Testcontainers instance.
