package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

var defaultContextKeys = []string{"trace_id", "tenant_id", "correlation_id"}

type Environment string

func (e Environment) Val() string {
	return strings.ToUpper(string(e))
}

const DevelopmentEnvironment = Environment("SANDBOX")

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

func Debug(ctx context.Context, msg string, attrs ...slog.Attr) {
	ctxAttr := extractDefaultAttributesFromContext(ctx, defaultContextKeys)
	attrs = append(ctxAttr, attrs...)
	slog.Default().LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
}

func Info(ctx context.Context, msg string, attrs ...slog.Attr) {
	ctxAttr := extractDefaultAttributesFromContext(ctx, defaultContextKeys)
	attrs = append(ctxAttr, attrs...)
	slog.Default().LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
}

func Warn(ctx context.Context, msg string, attrs ...slog.Attr) {
	ctxAttr := extractDefaultAttributesFromContext(ctx, defaultContextKeys)
	attrs = append(ctxAttr, attrs...)
	slog.Default().LogAttrs(ctx, slog.LevelWarn, msg, attrs...)
}

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
