# Project Issues & TODOs

## Index

| Issue                                                   | File                         | Line |
|---------------------------------------------------------|------------------------------|------|
| [Use DSN Factory in Config](#use-dsn-factory-in-config) | `internal/config/types.go`   | 45   |
| [Refactor Client Internals](#refactor-client-internals) | `pkg/database/client.go`     | 38   |
| [Use Structured DSN Type](#use-structured-dsn-type)     | `pkg/database/config.go`     | 29   |
| [Review Transaction Logic](#review-transaction-logic)   | `pkg/database/operations.go` | 113  |
| [Decouple Server from Echo](#decouple-server-from-echo) | `pkg/server/types.go`        | 28   |

---

### Use DSN Factory in Config

**File:** `internal/config/types.go:45`

Currently, `SqlConfig` handles the `DSN` as a raw string. The goal is to integrate a DSN factory/builder (likely from
`pkg/database/dsn`) to construct connection strings programmatically. This will improve validation and facilitate the
handling of driver-specific parameters (like charset, timeouts, or parseTime) without manual string concatenation.

### Refactor Client Internals

**File:** `pkg/database/client.go:38`

The `Client` struct manages writer and reader connections and their respective timeouts as separate fields. This task
involves grouping these related fields into a single unexported internal struct (e.g.,
`type connection struct { db *sql.DB; timeout time.Duration }`). This refactoring will reduce redundancy in the `Client`
struct and simplify methods that need to operate on both pools.

### Use Structured DSN Type

**File:** `pkg/database/config.go:29`

Re-evaluate the usage of a raw string for `DSN` in the `Config` struct. Switching to a structured `DSN` type (as
produced by the DSN builder) would provide better type safety. It ensures that the connection string is valid and
properly formatted before it reaches the connection logic in `newSQL`.

### Review Transaction Logic

**File:** `pkg/database/operations.go:113`

The `InTransaction` implementation manages the lifecycle of a database transaction, including panic recovery and
error-based rollbacks. This task is to double-check the robustness of this implementation, especially regarding:

- Proper context propagation within the `TxFunc`.
- Edge cases in the commit/rollback sequence.
- Ensuring that the recovered panic doesn't shadow the original error if both occur.

### Decouple Server from Echo

**File:** `pkg/server/types.go:28`

The `Route` struct currently uses `echo.MiddlewareFunc`, creating a direct dependency on the Echo framework. This issue
aims to abstract the middleware and handler types into internal interfaces or custom function signatures. This will make
the `server` package framework-agnostic, allowing the underlying router to be swapped (e.g., to Chi or Gin) with minimal
changes to the rest of the application.
