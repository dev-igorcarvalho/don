package echoserver

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimitMiddleware(t *testing.T) {
	e := echo.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("allows requests within limit", func(t *testing.T) {
		mw := RateLimitMiddleware(ctx, 10, 1)
		h := mw(func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("blocks requests exceeding limit", func(t *testing.T) {
		// RPS: 1, Burst: 1
		mw := RateLimitMiddleware(ctx, 1, 1)
		h := mw(func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		// First request (allowed)
		req1 := httptest.NewRequest(http.MethodGet, "/", nil)
		rec1 := httptest.NewRecorder()
		c1 := e.NewContext(req1, rec1)
		err1 := h(c1)
		assert.NoError(t, err1)
		assert.Equal(t, http.StatusOK, rec1.Code)

		// Second request (blocked)
		req2 := httptest.NewRequest(http.MethodGet, "/", nil)
		rec2 := httptest.NewRecorder()
		c2 := e.NewContext(req2, rec2)
		err2 := h(c2)

		assert.Error(t, err2)
		he, ok := err2.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusTooManyRequests, he.Code)
		assert.Equal(t, "1", rec2.Header().Get("Retry-After"))
	})

	t.Run("limits different IPs separately", func(t *testing.T) {
		mw := RateLimitMiddleware(ctx, 1, 1)
		h := mw(func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		// First request from IP 1 (allowed)
		req1 := httptest.NewRequest(http.MethodGet, "/", nil)
		req1.RemoteAddr = "1.1.1.1:1234"
		rec1 := httptest.NewRecorder()
		c1 := e.NewContext(req1, rec1)
		_ = h(c1)
		assert.Equal(t, http.StatusOK, rec1.Code)

		// First request from IP 2 (allowed)
		req2 := httptest.NewRequest(http.MethodGet, "/", nil)
		req2.RemoteAddr = "2.2.2.2:1234"
		rec2 := httptest.NewRecorder()
		c2 := e.NewContext(req2, rec2)
		err2 := h(c2)
		assert.NoError(t, err2)
		assert.Equal(t, http.StatusOK, rec2.Code)
	})
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	e := echo.New()

	t.Run("injects security headers for GET request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mw := SecurityHeadersMiddleware(echo.MIMEApplicationJSON)
		h := mw(func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		err := h(c)
		assert.NoError(t, err)
		assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
		assert.Equal(t, "default-src 'none'", rec.Header().Get("Content-Security-Policy"))
	})

	t.Run("returns 415 for POST request with missing Content-Type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mw := SecurityHeadersMiddleware(echo.MIMEApplicationJSON)
		h := mw(func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		err := h(c)
		assert.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusUnsupportedMediaType, he.Code)
	})

	t.Run("allows POST request with application/json Content-Type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mw := SecurityHeadersMiddleware(echo.MIMEApplicationJSON)
		h := mw(func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		err := h(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	})

	t.Run("allows POST request with multiple allowed Content-Types", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set(echo.HeaderContentType, "application/xml")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mw := SecurityHeadersMiddleware(echo.MIMEApplicationJSON, "application/xml")
		h := mw(func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		err := h(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestContextFromHeaderMiddleware(t *testing.T) {
	e := echo.New()
	t.Run("extracts headers and injects them into context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Trace-Id", "test-trace-id")
		req.Header.Set("X-Tenant-Id", "test-tenant-id")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mw := ContextFromHeaderMiddleware("X-Trace-Id", "X-Tenant-Id", "X-Missing")
		h := mw(func(c echo.Context) error {
			ctx := c.Request().Context()
			assert.Equal(t, "test-trace-id", ctx.Value("X-Trace-Id"))
			assert.Equal(t, "test-tenant-id", ctx.Value("X-Tenant-Id"))
			assert.Nil(t, ctx.Value("X-Missing"))
			return c.String(http.StatusOK, "ok")
		})

		err := h(c)
		assert.NoError(t, err)
	})
}

func TestLoggerMiddleware(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	logger := slog.New(handler)

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
		assert.NoError(t, err)

		var logEntry map[string]interface{}
		err = json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "test-service", logEntry["service"])
		assert.Equal(t, "GET", logEntry["method"])
		assert.Equal(t, "/test-path", logEntry["path"])
		assert.Equal(t, float64(http.StatusOK), logEntry["status"])
		assert.Contains(t, logEntry, "latency")
	})

	t.Run("logs errors with error level", func(t *testing.T) {
		buf.Reset()

		mw := LoggerMiddleware()
		h := mw(func(c echo.Context) error {
			return echo.NewHTTPError(http.StatusBadRequest, "bad request error")
		})

		_ = h(c)

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "ERROR", logEntry["level"])
		assert.NotNil(t, logEntry["error"])
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

	oldLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(oldLogger)

	t.Run("recovers from panic and logs error with stack trace", func(t *testing.T) {
		buf.Reset()

		mw := RecoveryMiddleware()
		h := mw(func(c echo.Context) error {
			panic("something went wrong")
		})

		e.HTTPErrorHandler = DefaultErrorHandler

		_ = h(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "ERROR", logEntry["level"])
		assert.Equal(t, "panic recovered", logEntry["msg"])
		assert.Equal(t, "something went wrong", logEntry["error"])
		assert.Contains(t, logEntry, "stack")
	})
}

