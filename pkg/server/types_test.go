package echoserver

import (
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, method, route.method)
	assert.Equal(t, path, route.path)
	assert.Equal(t, handler, route.handler)
	assert.Len(t, route.middleware, len(middleware))
}

func TestNewGroup(t *testing.T) {
	prefix := "/v1"
	middleware := []echo.MiddlewareFunc{
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		},
	}

	group := NewGroup(prefix, middleware...)

	assert.Equal(t, prefix, group.prefix)
	assert.Len(t, group.middleware, len(middleware))
	assert.Empty(t, group.routes)
}

func TestGroup_AddRoutes(t *testing.T) {
	group := NewGroup("/v1")
	route1 := NewRoute("GET", "/users", &mockHandler{})
	route2 := NewRoute("POST", "/users", &mockHandler{})

	group.AddRoutes(route1, route2)

	assert.Len(t, group.routes, 2)
	assert.Equal(t, route1, group.routes[0])
	assert.Equal(t, route2, group.routes[1])

	// Test method chaining
	route3 := NewRoute("GET", "/status", &mockHandler{})
	group.AddRoutes(route3)

	assert.Len(t, group.routes, 3)
}
