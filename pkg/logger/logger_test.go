package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
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
			if got := tt.env.Val(); got != tt.want {
				t.Errorf("Environment.Val() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractDefaultAttributesFromContext(t *testing.T) {
	keys := []string{"trace_id", "tenant_id"}
	
	t.Run("nil context", func(t *testing.T) {
		attrs := extractDefaultAttributesFromContext(nil, keys)
		if len(attrs) != 0 {
			t.Errorf("expected 0 attrs, got %d", len(attrs))
		}
	})

	t.Run("empty keys", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "trace_id", "123")
		attrs := extractDefaultAttributesFromContext(ctx, []string{})
		if len(attrs) != 0 {
			t.Errorf("expected 0 attrs, got %d", len(attrs))
		}
	})

	t.Run("with values", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "trace_id", "123")
		ctx = context.WithValue(ctx, "tenant_id", "456")
		attrs := extractDefaultAttributesFromContext(ctx, keys)
		
		if len(attrs) != 2 {
			t.Fatalf("expected 2 attrs, got %d", len(attrs))
		}
		
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
		
		if !foundTrace || !foundTenant {
			t.Errorf("did not find expected attributes: trace=%v, tenant=%v", foundTrace, foundTenant)
		}
	})

	t.Run("missing values", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "trace_id", "123")
		attrs := extractDefaultAttributesFromContext(ctx, keys)
		
		if len(attrs) != 1 {
			t.Fatalf("expected 1 attr, got %d", len(attrs))
		}
		if attrs[0].Key != "trace_id" {
			t.Errorf("expected trace_id, got %s", attrs[0].Key)
		}
	})
}

func TestLogging(t *testing.T) {
	// We use a buffer to capture output. 
	// Since Setup sets the global default logger, we need to be careful with parallel tests.
	// However, slog handlers usually take an io.Writer.
	
	var buf bytes.Buffer
	
	t.Run("Setup Sandbox (TextHandler)", func(t *testing.T) {
		buf.Reset()
		handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
		slog.SetDefault(slog.New(handler))

		ctx := context.WithValue(context.Background(), "trace_id", "test-trace")
		Debug(ctx, "debug message", slog.String("extra", "value"))
		
		output := buf.String()
		if !strings.Contains(output, "level=DEBUG") {
			t.Errorf("expected DEBUG level in output, got: %s", output)
		}
		if !strings.Contains(output, "msg=\"debug message\"") && !strings.Contains(output, "msg=debug message") {
			t.Errorf("expected message in output, got: %s", output)
		}
		if !strings.Contains(output, "trace_id=test-trace") {
			t.Errorf("expected trace_id in output, got: %s", output)
		}
		if !strings.Contains(output, "extra=value") {
			t.Errorf("expected extra attribute in output, got: %s", output)
		}
	})

	t.Run("Setup Production (JSONHandler)", func(t *testing.T) {
		buf.Reset()
		handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
		slog.SetDefault(slog.New(handler))

		ctx := context.WithValue(context.Background(), "tenant_id", "test-tenant")
		Info(ctx, "info message")
		
		var logMap map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &logMap); err != nil {
			t.Fatalf("failed to unmarshal JSON log: %v", err)
		}
		
		if logMap["level"] != "INFO" {
			t.Errorf("expected INFO level, got %v", logMap["level"])
		}
		if logMap["msg"] != "info message" {
			t.Errorf("expected msg 'info message', got %v", logMap["msg"])
		}
		if logMap["tenant_id"] != "test-tenant" {
			t.Errorf("expected tenant_id 'test-tenant', got %v", logMap["tenant_id"])
		}
	})

	t.Run("Warn and Error", func(t *testing.T) {
		buf.Reset()
		handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
		slog.SetDefault(slog.New(handler))

		Warn(context.Background(), "warn message")
		Error(context.Background(), "error message")
		
		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 2 {
			t.Fatalf("expected 2 log lines, got %d", len(lines))
		}
		
		var log1, log2 map[string]interface{}
		json.Unmarshal([]byte(lines[0]), &log1)
		json.Unmarshal([]byte(lines[1]), &log2)
		
		if log1["level"] != "WARN" {
			t.Errorf("expected WARN level, got %v", log1["level"])
		}
		if log2["level"] != "ERROR" {
			t.Errorf("expected ERROR level, got %v", log2["level"])
		}
	})
}

func TestSetup(t *testing.T) {
	// Since Setup uses os.Stdout, it's hard to capture its output directly without redirecting os.Stdout.
	// But we can at least verify it doesn't panic and sets a default logger.
	
	t.Run("Setup Sandbox", func(t *testing.T) {
		Setup(DevelopmentEnvironment)
		// No panic is a good sign.
	})

	t.Run("Setup Production", func(t *testing.T) {
		Setup(Environment("PRODUCTION"))
		// No panic is a good sign.
	})
}
