# Logger Improvements

This document outlines the analysis, identified gaps, and proposed improvements for the structured logging implementation in the **Don** project.

## Current State Analysis

The logger implementation in `pkg/logger` provides a wrapper around `log/slog`. It successfully implements environment-based behavior:
- **SANDBOX:** Uses `TextHandler` with `Debug` level.
- **Production/Other:** Uses `JSONHandler` with `Info` level.
- **Context Awareness:** Manually extracts `trace_id`, `tenant_id`, and `correlation_id` from `context.Context`.

---

## Identified Gaps

### 1. Context Key Collisions
The current implementation uses raw strings (`"trace_id"`) as context keys. This violates Go best practices and can lead to collisions with other packages.
- **Risk:** Unintentional overwriting or retrieval of incorrect values.

**Explanation:** In Go, the `context` package uses `any` (interface{}) for keys. When you use a string like `"trace_id"`, any other package in your dependency tree that also uses the string `"trace_id"` will share the same "namespace." If a third-party library happens to set a value with that same string key, it could overwrite your value or cause your code to retrieve data it didn't expect. The standard way to prevent this is by using a custom, unexported type for keys, ensuring that even if another package uses the same underlying string, the type mismatch prevents a collision.

### 2. Missing Source Attribution
Logs do not include the file name or line number where the log event was triggered.
- **Risk:** Increased difficulty in pinpointing the source of errors in production.

**Explanation:** Structured logs are most powerful when they provide immediate context. Without "source" information (file and line number), a log message like `unexpected error occurred` requires manual searching through the codebase to find where it was emitted. By enabling source attribution, every log entry automatically includes a `source` field containing the file path and line number, drastically reducing the "Mean Time to Recovery" (MTTR) during incidents.

### 3. Wrapper Overheads & API Fragmentation
By wrapping `slog` functions, the package creates a proprietary API that must be maintained.
- **Risk:** If developers want to use advanced `slog` features (like `WithGroup` or `WithAttrs` on a logger instance), the current wrappers become an obstacle.

**Explanation:** The `slog` library is designed to be highly composable. It allows developers to create sub-loggers that carry specific attributes (e.g., `logger.With("component", "auth")`) or group attributes together. When you wrap `slog` in a custom package like `logger.Info(ctx, msg)`, you hide these powerful features. Developers then have to either add those features to your wrapper (increasing maintenance) or bypass your wrapper entirely, leading to inconsistent logging styles across the project.

### 4. Incorrect Caller Information
If `AddSource: true` were enabled in the current wrappers, the logs would report the location within `pkg/logger/logger.go` rather than the actual caller's location.

**Explanation:** When `slog` captures the source location, it looks at the call stack. If you use a wrapper function like `logger.Info`, the "caller" from `slog`'s perspective is always the line inside `logger.go` where `slog.Log` is called. To fix this in a wrapper, you must manually capture the Program Counter (PC) using `runtime.Callers` and pass it to `slog`, which adds complexity and a small performance penalty. Using a `slog.Handler` middleware avoids this because it allows the standard library to handle the stack depth correctly.

---

## Proposed Improvements

### 1. Custom Context Handler (Middleware Pattern)
Instead of manual extraction in every function, implement a custom `slog.Handler` that wraps the base handler. This handler will automatically extract IDs from the context during the `Handle` call.

**Benefits:**
- Simplifies the package API.
- Compatible with standard `slog.InfoContext`, `slog.ErrorContext`, etc.
- Seamlessly integrates with the standard library ecosystem.

**Detailed Explanation:**
A `slog.Handler` is an interface that decides how log records are processed. By creating a "Middleware Handler" (also known as a Decorator), we wrap a standard handler (like `JSONHandler`) and intercept every log request.

When a developer calls `slog.InfoContext(ctx, "msg")`, our handler:
1. Receives the `context.Context`.
2. Inspects the context for our specific keys (Trace ID, Tenant ID).
3. Injects these values as attributes into the `slog.Record`.
4. Passes the enriched record to the underlying `JSONHandler`.

#### Example Implementation:
```go
type ContextHandler struct {
    slog.Handler
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
    // 1. Extract values from context
    if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
        r.AddAttrs(slog.String("trace_id", traceID))
    }
    // ... repeat for other keys

    // 2. Pass the record to the wrapped handler
    return h.Handler.Handle(ctx, r)
}
```

#### Comparison: Call Site Efficiency

**Before (Wrapper Pattern):**
You must use the custom `logger` package and rely on its hardcoded extraction logic.
```go
// Only works if you use THIS specific wrapper
logger.Info(ctx, "user logged in", slog.String("user_id", "123"))
```

**After (Middleware Pattern):**
You can use the standard library directly. The context extraction happens automatically "under the hood."
```go
// Standard library call - automatically enriched by ContextHandler
slog.InfoContext(ctx, "user logged in", slog.String("user_id", "123"))

// Even works with third-party libraries that use slog!
```

### 2. Type-Safe Context Keys
Define an unexported custom type for context keys to ensure isolation.

```go
type ctxKey string
const traceIDKey ctxKey = "trace_id"
```

**Explanation:** By defining `type ctxKey string`, we create a new type that is distinct from the built-in `string` type. Even if another package does the exact same thing, their `ctxKey` is different from our `ctxKey` because they belong to different packages. This is the idiomatic Go way to ensure that the values you put into a context are only accessible by the code that is supposed to see them.

### 3. Automatic Source Attribution
Enable `AddSource: true` in `HandlerOptions`. When using the custom handler pattern, `slog` correctly handles the call stack, ensuring the logs point to the real source code location.

**Explanation:** This is a simple configuration toggle in `slog.HandlerOptions`. When paired with the custom handler approach (rather than wrappers), it provides the most accurate and useful debugging information with zero extra effort from the developer at the call site.

### 4. Integration with Transport Layer
Align the logger's context extraction with the middleware used in the `adapters/http_server.go`. If Echo is used, ensure the keys match what the Echo middleware injects.

**Explanation:** Logs are often part of a larger request lifecycle. If your HTTP framework (Echo) generates a Request ID and stores it in the context, the logger should be configured to look for that *exact* key. This creates a "Golden Thread" where a single ID can be followed from the initial HTTP request, through various use cases and adapters, all the way to the final log output or database query.

---

## Recommended Refactoring Strategy

1.  **Phase 1:** Implement `ContextHandler` in `pkg/logger`.
2.  **Phase 2:** Update `Setup` to use the new handler.
3.  **Phase 3:** Gradually replace wrapper calls (`logger.Info`) with standard `slog` calls or keep thin wrappers that pass the correct Program Counter (PC).
