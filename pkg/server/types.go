package echoserver

import (
	"context"

	"github.com/labstack/echo/v4"
)

// Server is the interface that all servers must implement.
type Server interface {
	RegisterRoutes(routes ...*Route)
	RegisterGroups(groups ...*Group)
	Start() error
	Shutdown(ctx context.Context) error
}

// Handler is the interface that all route handlers must implement.
// This enforces the use of handler structs.
type Handler interface {
	Handle(c echo.Context) error
}

// Route encapsulates a single endpoint's data.
type Route struct {
	method     string
	path       string
	handler    Handler
	middleware []echo.MiddlewareFunc
}

// NewRoute creates a new Route instance requiring a Handler interface.
func NewRoute(method, path string, h Handler, mws ...echo.MiddlewareFunc) *Route {
	return &Route{
		method:     method,
		path:       path,
		handler:    h,
		middleware: mws,
	}
}

// Group encapsulates a set of routes under a common prefix.
type Group struct {
	prefix     string
	middleware []echo.MiddlewareFunc
	routes     []*Route
}

// NewGroup creates a new Group instance.
func NewGroup(prefix string, mws ...echo.MiddlewareFunc) *Group {
	return &Group{
		prefix:     prefix,
		middleware: mws,
		routes:     []*Route{},
	}
}

// AddRoutes adds one or more routes to the group using variadic arguments.
func (g *Group) AddRoutes(routes ...*Route) *Group {
	g.routes = append(g.routes, routes...)
	return g
}
