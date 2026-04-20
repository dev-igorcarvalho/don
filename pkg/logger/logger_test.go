package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvironment_Val(t *testing.T) {
	tests := []struct {
		env  Environment
		want string
	}{
		{Environment("sandbox"), "SANDBOX"},
		{Environment("production"), "PRODUCTION"},
		{Environment("DEV"), "DEV"},
	}
	for _, tt := range tests {
		t.Run(string(tt.env), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.env.Val())
		})
	}
}

func TestExtractDefaultAttributesFromContext(t *testing.T) {
	keys := []string{"trace_id", "tenant_id"}
	
	t.Run("nil context", func(t *testing.T) {
		attrs := extractDefaultAttributesFromContext(nil, keys)
		assert.Empty(t, attrs)
	})

	t.Run("empty keys", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "trace_id", "123")
		attrs := extractDefaultAttributesFromContext(ctx, []string{})
		assert.Empty(t, attrs)
	})

	t.Run("with values", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "trace_id", "123")
		ctx = context.WithValue(ctx, "tenant_id", "456")
		attrs := extractDefaultAttributesFromContext(ctx, keys)
		
		require.Len(t, attrs, 2)
		
		foundTrace := false
		foundTenant := false
		for _, attr := range attrs {
			if attr.Key == "trace_id" && attr.Value.String() == "123" {
				foundTrace = true
			}
			if attr.Key == "tenant_id" && attr.Value.String() == "456" {
				foundTenant = true
			}
		}
		
		assert.True(t, foundTrace, "did not find trace_id")
		assert.True(t, foundTenant, "did not find tenant_id")
	})

	t.Run("missing values", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "trace_id", "123")
		attrs := extractDefaultAttributesFromContext(ctx, keys)
		
		require.Len(t, attrs, 1)
		assert.Equal(t, "trace_id", attrs[0].Key)
	})
}

func TestLogging(t *testing.T) {
	var buf bytes.Buffer
	
	t.Run("Setup Sandbox (TextHandler)", func(t *testing.T) {
		buf.Reset()
		handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
		slog.SetDefault(slog.New(handler))

		ctx := context.WithValue(context.Background(), "trace_id", "test-trace")
		Debug(ctx, "debug message", slog.String("extra", "value"))
		
		output := buf.String()
		assert.Contains(t, output, "level=DEBUG")
		assert.True(t, strings.Contains(output, "msg=\"debug message\"") || strings.Contains(output, "msg=debug message"))
		assert.Contains(t, output, "trace_id=test-trace")
		assert.Contains(t, output, "extra=value")
	})

	t.Run("Setup Production (JSONHandler)", func(t *testing.T) {
		buf.Reset()
		handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
		slog.SetDefault(slog.New(handler))

		ctx := context.WithValue(context.Background(), "tenant_id", "test-tenant")
		Info(ctx, "info message")
		
		var logMap map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logMap)
		require.NoError(t, err)
		
		assert.Equal(t, "INFO", logMap["level"])
		assert.Equal(t, "info message", logMap["msg"])
		assert.Equal(t, "test-tenant", logMap["tenant_id"])
	})

	t.Run("Warn and Error", func(t *testing.T) {
		buf.Reset()
		handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
		slog.SetDefault(slog.New(handler))

		Warn(context.Background(), "warn message")
		Error(context.Background(), "error message")
		
		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		require.Len(t, lines, 2)
		
		var log1, log2 map[string]interface{}
		json.Unmarshal([]byte(lines[0]), &log1)
		json.Unmarshal([]byte(lines[1]), &log2)
		
		assert.Equal(t, "WARN", log1["level"])
		assert.Equal(t, "ERROR", log2["level"])
	})
}

func TestSetup(t *testing.T) {
	t.Run("Setup Sandbox", func(t *testing.T) {
		assert.NotPanics(t, func() { Setup(DevelopmentEnvironment) })
	})

	t.Run("Setup Production", func(t *testing.T) {
		assert.NotPanics(t, func() { Setup(Environment("PRODUCTION")) })
	})
}
