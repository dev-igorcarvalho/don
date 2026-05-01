// ---
// title: Logger
// description: Provides structured logging functionality using slog, supporting environment-based configuration and context-aware attributes.
// last_updated: 2026-05-01
// type: Package
// ---

// Package logger provides structured logging functionality using slog.
// It supports environment-based configuration and context-aware logging.
package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

var defaultContextKeys = []string{"trace_id", "tenant_id", "correlation_id"}

// Environment represents the deployment environment.
type Environment string

// Val returns the string value of the environment in uppercase.
func (e Environment) Val() string {
	return strings.ToUpper(string(e))
}

// DevelopmentEnvironment defines the sandbox/development environment constant.
const DevelopmentEnvironment = Environment("SANDBOX")

// Setup initializes the default slog logger based on the provided environment.
// It uses a text handler for DevelopmentEnvironment and a JSON handler for others.
func Setup(env Environment) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{}

	if env.Val() == DevelopmentEnvironment.Val() {
		opts.Level = slog.LevelDebug
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		opts.Level = slog.LevelInfo
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}

// Debug logs a message at LevelDebug with context attributes.
func Debug(ctx context.Context, msg string, attrs ...slog.Attr) {
	ctxAttr := extractDefaultAttributesFromContext(ctx, defaultContextKeys)
	attrs = append(ctxAttr, attrs...)
	slog.Default().LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
}

// Info logs a message at LevelInfo with context attributes.
func Info(ctx context.Context, msg string, attrs ...slog.Attr) {
	ctxAttr := extractDefaultAttributesFromContext(ctx, defaultContextKeys)
	attrs = append(ctxAttr, attrs...)
	slog.Default().LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
}

// Warn logs a message at LevelWarn with context attributes.
func Warn(ctx context.Context, msg string, attrs ...slog.Attr) {
	ctxAttr := extractDefaultAttributesFromContext(ctx, defaultContextKeys)
	attrs = append(ctxAttr, attrs...)
	slog.Default().LogAttrs(ctx, slog.LevelWarn, msg, attrs...)
}

// Error logs a message at LevelError with context attributes.
func Error(ctx context.Context, msg string, attrs ...slog.Attr) {
	ctxAttr := extractDefaultAttributesFromContext(ctx, defaultContextKeys)
	attrs = append(ctxAttr, attrs...)
	slog.Default().LogAttrs(ctx, slog.LevelError, msg, attrs...)
}

func extractDefaultAttributesFromContext(ctx context.Context, keys []string) []slog.Attr {
	if ctx == nil || len(keys) == 0 {
		return nil
	}

	attrs := make([]slog.Attr, 0, len(keys))

	for _, key := range keys {
		if val := ctx.Value(key); val != nil {
			attrs = append(attrs, slog.Any(key, val))
		}
	}

	return attrs
}
