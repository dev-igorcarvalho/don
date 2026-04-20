package echoserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

func TestNew(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		s := New()
		if s.port != defaultPort {
			t.Errorf("expected default port %s, got %s", defaultPort, s.port)
		}
		if s.app == nil {
			t.Error("expected echo instance to be initialized")
		}
	})

	t.Run("with custom port", func(t *testing.T) {
		customPort := "9090"
		s := New(WithPort(customPort))
		if s.port != customPort {
			t.Errorf("expected port %s, got %s", customPort, s.port)
		}
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
		
		// We can't easily check if middleware was added without executing a request
		// but we can register a route and call it.
		s.RegisterRoutes(NewRoute("GET", "/test", &mockHandler{}))
		
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		s.app.ServeHTTP(rec, req)
		
		if !middlewareCalled {
			t.Error("expected middleware to be called")
		}
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
	
	if rec.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d", rec.Code)
	}
	if !handlerCalled {
		t.Error("expected handler to be called")
	}
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
	
	if rec.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d", rec.Code)
	}
	if !handlerCalled {
		t.Error("expected handler to be called")
	}
}

func TestEchoServer_Shutdown(t *testing.T) {
	s := New()
	
	// Start server in background
	go func() {
		_ = s.Start()
	}()
	
	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)
	
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	err := s.Shutdown(ctx)
	if err != nil {
		t.Errorf("expected nil error on shutdown, got %v", err)
	}
}

// testHandler is a helper for testing routes with a custom handle function.
type testHandler struct {
	handleFunc func(c echo.Context) error
}

func (h *testHandler) Handle(c echo.Context) error {
	return h.handleFunc(c)
}
