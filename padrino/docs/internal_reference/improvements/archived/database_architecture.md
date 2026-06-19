# Database Architecture & Evolution
**Role:** Principal Engineer (Infrastructure/Platforms)
**Project:** Don
**Scope:** `pkg/database/sql.go`

## 1. Executive Summary
The current implementation of `pkg/database` provides a solid baseline for basic SQL connectivity. However, to support Uber-scale reliability, observability, and developer ergonomics, several architectural shifts are recommended. The primary focus should move from simple connection management to a **resilient data access layer** that provides automatic observability and prevents common pitfalls like replica lag issues.

---

## 2. Key Improvement Areas

### 2.1. API Design: Functional Options Pattern
The current `NewConfig` and `NewSQL` functions suffer from "parameter bloat." As we add more configuration (e.g., TLS, timeouts, interceptors), the constructor becomes unmanageable.

**Recommendation:** Adopt the functional options pattern for better extensibility and cleaner defaults.

```go
// Example of idiomatic construction
db, err := database.Open(driver, dsn,
    database.WithMaxOpenConns(100),
    database.WithMaxIdleConns(10),
    database.WithWarmup(5 * time.Second),
)
```

### 2.2. Reliability: Handling DSNs and Drivers
Passing a raw DSN string is risky as it often contains secrets. 

**Recommendation:**
- **Context Deadlines:** Ensure `NewSQL` uses the `ConnectTimeout` not just for Ping, but as a hard limit for the entire initialization phase.

---

## 3. Resilience: Exponential Backoff
Startup failure due to DB unavailability is a common "flapping" cause in Kubernetes. We will implement a retry strategy for the `Warmup` phase.

```go
func (c *Config) WarmupWithRetry(ctx context.Context, db *sql.DB) error {
    backoff := 500 * time.Millisecond
    for i := 0; i < 5; i++ {
        if err := db.PingContext(ctx); err == nil {
            return nil
        }
        time.Sleep(backoff)
        backoff *= 2
    }
    return ErrDatabaseUnreachable
}
```

---

## 4. Advanced Health Probes
A binary health check is insufficient for Primary/Replica setups.

```go
type HealthStatus struct {
    WriterAlive bool
    ReaderAlive bool
    OpenConns   int
	IdleConns int
    Message     string
}

func (p *SQLPair) HealthCheck(ctx context.Context) HealthStatus {
    // Pings both and returns aggregated status
}
```

---

## 5. Proposed Refactored Implementation Snippet

Here is how a "Level 5" implementation might look:

```go
type Client struct {
    writer *sql.DB
    reader *sql.DB
    logger *slog.Logger
    // metrics registry
}

func Open(cfg Config, opts ...Option) (*Client, error) {
    // 1. Apply Options
    // 2. Initialize Writer and Reader with interceptors
    // 3. Start background metrics collection
    // 4. Return wrapped Client
}
```

---

## 6. Implementation Plan
1. **Refactor `Config`**: Move to unexported fields with functional options.
2. **Implement `Transactor`**: Provide a concrete implementation for `sql.DB`.
3. **Add Middleware Hooks**: Allow injecting loggers and tracers.
4. **Update Adapters**: Migrate `internal/adapters` to use the `Transactor` and `SQLPair` interfaces.
