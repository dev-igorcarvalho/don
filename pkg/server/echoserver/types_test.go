package echoserver

import (
	"testing"

	"github.com/labstack/echo/v4"
)

// mockHandler is a dummy implementation of the Handler interface for testing.
type mockHandler struct{}

func (m *mockHandler) Handle(c echo.Context) error {
	return nil
}

func TestNewRoute(t *testing.T) {
	method := "GET"
	path := "/test"
	handler := &mockHandler{}
	middleware := []echo.MiddlewareFunc{
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		},
	}

	route := NewRoute(method, path, handler, middleware...)

	if route.method != method {
		t.Errorf("expected method %s, got %s", method, route.method)
	}
	if route.path != path {
		t.Errorf("expected path %s, got %s", path, route.path)
	}
	if route.handler != handler {
		t.Errorf("expected handler %p, got %p", handler, route.handler)
	}
	if len(route.middleware) != len(middleware) {
		t.Errorf("expected %d middleware, got %d", len(middleware), len(route.middleware))
	}
}

func TestNewGroup(t *testing.T) {
	prefix := "/v1"
	middleware := []echo.MiddlewareFunc{
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		},
	}

	group := NewGroup(prefix, middleware...)

	if group.prefix != prefix {
		t.Errorf("expected prefix %s, got %s", prefix, group.prefix)
	}
	if len(group.middleware) != len(middleware) {
		t.Errorf("expected %d middleware, got %d", len(middleware), len(group.middleware))
	}
	if len(group.routes) != 0 {
		t.Errorf("expected 0 routes, got %d", len(group.routes))
	}
}

func TestGroup_AddRoutes(t *testing.T) {
	group := NewGroup("/v1")
	route1 := NewRoute("GET", "/users", &mockHandler{})
	route2 := NewRoute("POST", "/users", &mockHandler{})

	group.AddRoutes(route1, route2)

	if len(group.routes) != 2 {
		t.Errorf("expected 2 routes, got %d", len(group.routes))
	}
	if group.routes[0] != route1 {
		t.Errorf("expected first route to be route1")
	}
	if group.routes[1] != route2 {
		t.Errorf("expected second route to be route2")
	}

	// Test method chaining
	route3 := NewRoute("GET", "/status", &mockHandler{})
	group.AddRoutes(route3)

	if len(group.routes) != 3 {
		t.Errorf("expected 3 routes, got %d", len(group.routes))
	}
}
