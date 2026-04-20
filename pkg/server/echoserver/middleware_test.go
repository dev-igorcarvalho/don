package echoserver

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestLoggerMiddleware(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	logger := slog.New(handler)

	// Set as default for the test
	oldLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(oldLogger)

	t.Run("logs with base attributes and request data", func(t *testing.T) {
		buf.Reset()
		
		mw := LoggerMiddleware(slog.String("service", "test-service"))
		h := mw(func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		err := h(c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var logEntry map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
			t.Fatalf("failed to parse log output: %v", err)
		}

		// Check base attributes
		if logEntry["service"] != "test-service" {
			t.Errorf("expected service=test-service, got %v", logEntry["service"])
		}

		// Check dynamic attributes
		if logEntry["method"] != "GET" {
			t.Errorf("expected method=GET, got %v", logEntry["method"])
		}
		if logEntry["path"] != "/test-path" {
			t.Errorf("expected path=/test-path, got %v", logEntry["path"])
		}
		if logEntry["status"] != float64(http.StatusOK) { // JSON unmarshals numbers as float64
			t.Errorf("expected status=200, got %v", logEntry["status"])
		}
		if _, ok := logEntry["latency"]; !ok {
			t.Error("expected latency field to exist")
		}
	})

	t.Run("logs errors with error level", func(t *testing.T) {
		buf.Reset()
		
		mw := LoggerMiddleware()
		h := mw(func(c echo.Context) error {
			return echo.NewHTTPError(http.StatusBadRequest, "bad request error")
		})

		_ = h(c) // Error is handled by c.Error inside middleware

		var logEntry map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
			t.Fatalf("failed to parse log output: %v", err)
		}

		if logEntry["level"] != "ERROR" {
			t.Errorf("expected level=ERROR, got %v", logEntry["level"])
		}
		if logEntry["error"] == nil {
			t.Error("expected error field to exist")
		}
	})
}

func TestRecoveryMiddleware(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	logger := slog.New(handler)

	// Set as default for the test
	oldLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(oldLogger)

	t.Run("recovers from panic and logs error with stack trace", func(t *testing.T) {
		buf.Reset()

		mw := RecoveryMiddleware()
		h := mw(func(c echo.Context) error {
			panic("something went wrong")
		})

		// Echo's default error handler will catch the error passed to c.Error
		e.HTTPErrorHandler = DefaultErrorHandler

		err := h(c)
		if err != nil {
			// In our middleware, we call c.Error(err) which doesn't necessarily return an error from the handler
			// but we want to make sure the recovery worked.
		}

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}

		var logEntry map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
			t.Fatalf("failed to parse log output: %v", err)
		}

		if logEntry["level"] != "ERROR" {
			t.Errorf("expected level=ERROR, got %v", logEntry["level"])
		}
		if logEntry["msg"] != "panic recovered" {
			t.Errorf("expected msg='panic recovered', got %v", logEntry["msg"])
		}
		if logEntry["error"] != "something went wrong" {
			t.Errorf("expected error='something went wrong', got %v", logEntry["error"])
		}
		if logEntry["stack"] == nil {
			t.Error("expected stack field to exist")
		}
	})
}
