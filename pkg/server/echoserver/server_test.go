package echoserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		s := New()
		assert.Equal(t, defaultPort, s.port)
		assert.NotNil(t, s.app)
	})

	t.Run("with custom port", func(t *testing.T) {
		customPort := "9090"
		s := New(WithPort(customPort))
		assert.Equal(t, customPort, s.port)
	})

	t.Run("with health check", func(t *testing.T) {
		h := &testHandler{
			handleFunc: func(c echo.Context) error {
				return c.JSON(http.StatusOK, map[string]string{"status": "up"})
			},
		}
		s := New(WithHealthCheck(h))

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		s.app.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, `{"status":"up"}`+"\n", rec.Body.String())
	})

	t.Run("with middleware", func(t *testing.T) {
		middlewareCalled := false
		mw := func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				middlewareCalled = true
				return next(c)
			}
		}
		
		s := New(WithMiddleware(mw))
		
		s.RegisterRoutes(NewRoute("GET", "/test", &mockHandler{}))
		
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		s.app.ServeHTTP(rec, req)
		
		assert.True(t, middlewareCalled)
	})
}

func TestEchoServer_RegisterRoutes(t *testing.T) {
	s := New()
	handlerCalled := false
	
	h := &testHandler{
		handleFunc: func(c echo.Context) error {
			handlerCalled = true
			return c.String(http.StatusOK, "ok")
		},
	}
	
	s.RegisterRoutes(NewRoute("GET", "/hello", h))
	
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()
	s.app.ServeHTTP(rec, req)
	
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, handlerCalled)
}

func TestEchoServer_RegisterGroups(t *testing.T) {
	s := New()
	handlerCalled := false
	
	h := &testHandler{
		handleFunc: func(c echo.Context) error {
			handlerCalled = true
			return c.String(http.StatusOK, "ok")
		},
	}
	
	group := NewGroup("/api")
	group.AddRoutes(NewRoute("GET", "/v1", h))
	
	s.RegisterGroups(group)
	
	req := httptest.NewRequest(http.MethodGet, "/api/v1", nil)
	rec := httptest.NewRecorder()
	s.app.ServeHTTP(rec, req)
	
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, handlerCalled)
}

func TestEchoServer_Shutdown(t *testing.T) {
	s := New()
	
	go func() {
		_ = s.Start()
	}()
	
	time.Sleep(100 * time.Millisecond)
	
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	err := s.Shutdown(ctx)
	assert.NoError(t, err)
}

type testHandler struct {
	handleFunc func(c echo.Context) error
}

func (h *testHandler) Handle(c echo.Context) error {
	return h.handleFunc(c)
}
