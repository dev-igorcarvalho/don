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

type EchoServer struct {
	app  *echo.Echo
	port string
}

type Option func(*EchoServer)

func WithPort(port string) Option {
	return func(s *EchoServer) {
		s.port = port
	}
}

func WithMiddleware(middlewares ...echo.MiddlewareFunc) Option {
	return func(s *EchoServer) {
		s.app.Use(middlewares...)
	}
}

func WithErrorHandler(handler echo.HTTPErrorHandler) Option {
	return func(s *EchoServer) {
		s.app.HTTPErrorHandler = handler
	}
}

func WithHealthCheck(h server.Handler) Option {
	return func(s *EchoServer) {
		s.app.GET("/health", h.Handle)
	}
}

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

func (s *EchoServer) Start() error {
	addr := fmt.Sprintf(":%s", s.port)
	if err := s.app.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *EchoServer) Shutdown(ctx context.Context) error {
	return s.app.Shutdown(ctx)
}
