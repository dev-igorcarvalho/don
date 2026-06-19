// ---
// title: Echo Server
// description: Implementation of the server.Server interface using the Echo framework, featuring functional options for configuration.
// last_updated: 2026-05-03
// type: Implementation
// ---

// Package echoserver implements the server.Server interface using the Echo web framework.
// It provides a configurable HTTP server with support for routes, groups, and middleware.
package echoserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/dev-igorcarvalho/don/pkg/server"
	"github.com/labstack/echo/v4"
)

const defaultPort = "8080"

var _ server.Server = (*EchoServer)(nil)

// EchoServer implements the server.Server interface using the Echo framework.
type EchoServer struct {
	app  *echo.Echo
	port string
}

// Option defines a functional configuration type for EchoServer.
type Option func(*EchoServer)

// WithPort sets the listening port for the server.
func WithPort(port string) Option {
	return func(s *EchoServer) {
		s.port = port
	}
}

// WithMiddleware registers one or more global middlewares to the Echo application.
func WithMiddleware(middlewares ...echo.MiddlewareFunc) Option {
	return func(s *EchoServer) {
		s.app.Use(middlewares...)
	}
}

// WithErrorHandler sets a custom HTTP error handler for the server.
func WithErrorHandler(handler echo.HTTPErrorHandler) Option {
	return func(s *EchoServer) {
		s.app.HTTPErrorHandler = handler
	}
}

// WithHealthCheck registers a health check endpoint at /health.
func WithHealthCheck(h server.Handler) Option {
	return func(s *EchoServer) {
		s.app.GET("/health", h.Handle)
	}
}

// New creates and initializes a new EchoServer with the provided options.
func New(opts ...Option) *EchoServer {
	e := echo.New()
	s := &EchoServer{
		app:  e,
		port: defaultPort,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// RegisterRoutes adds individual routes to the server.
func (s *EchoServer) RegisterRoutes(routes ...*server.Route) {
	for _, r := range routes {
		s.app.Add(r.Method, r.Path, r.Handler.Handle, r.Middleware...)
	}
}

// RegisterGroups adds entire groups of routes.
func (s *EchoServer) RegisterGroups(groups ...*server.Group) {
	for _, g := range groups {
		echoGroup := s.app.Group(g.Prefix, g.Middleware...)
		for _, r := range g.Routes {
			echoGroup.Add(r.Method, r.Path, r.Handler.Handle, r.Middleware...)
		}
	}
}

// Start begins listening for HTTP requests on the configured port.
func (s *EchoServer) Start() error {
	addr := fmt.Sprintf(":%s", s.port)
	if err := s.app.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Shutdown gracefully stops the server within the provided context's deadline.
func (s *EchoServer) Shutdown(ctx context.Context) error {
	return s.app.Shutdown(ctx)
}
