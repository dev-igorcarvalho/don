// ---
// title: Server Types
// description: Defines core interfaces and structures for HTTP server implementations, including Server, Handler, Route, and Group.
// last_updated: 2026-05-03
// type: Interface
// ---

// Package server defines the core interfaces and types for building
// and managing HTTP servers within the application.
package server

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
	Method     string
	Path       string
	Handler    Handler
	Middleware []echo.MiddlewareFunc
}

// NewRoute creates a new Route instance requiring a Handler interface.
func NewRoute(method, path string, h Handler, mws ...echo.MiddlewareFunc) *Route {
	return &Route{
		Method:     method,
		Path:       path,
		Handler:    h,
		Middleware: mws,
	}
}

// Group encapsulates a set of routes under a common prefix.
type Group struct {
	Prefix     string
	Middleware []echo.MiddlewareFunc
	Routes     []*Route
}

// NewGroup creates a new Group instance.
func NewGroup(prefix string, mws ...echo.MiddlewareFunc) *Group {
	return &Group{
		Prefix:     prefix,
		Middleware: mws,
		Routes:     []*Route{},
	}
}

// AddRoutes adds one or more routes to the group using variadic arguments.
func (g *Group) AddRoutes(routes ...*Route) *Group {
	g.Routes = append(g.Routes, routes...)
	return g
}
