// ---
// title: Database Context Utilities
// description: Provides context-based signals for database operations, such as forcing reads from the writer connection.
// last_updated: 2026-05-08
// type: Implementation
// ---

package database

import "context"

type contextKey string

const (
	writerForcedKey contextKey = "db.writer_forced"
)

// WithWriterContext returns a new context that signals database operations to use the writer connection.
// This is useful for reading data immediately after a write to avoid replication lag.
func WithWriterContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, writerForcedKey, true)
}

// isWriterForced checks if the context has the writer forced signal.
func isWriterForced(ctx context.Context) bool {
	forced, ok := ctx.Value(writerForcedKey).(bool)
	return ok && forced
}
