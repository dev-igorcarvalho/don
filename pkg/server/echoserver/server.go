package echoserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

const defaultPort = "8080"

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
func (s *EchoServer) RegisterRoutes(routes ...*Route) {
	for _, r := range routes {
		s.app.Add(r.method, r.path, r.handler.Handle, r.middleware...)
	}
}

// RegisterGroups adds entire groups of routes.
func (s *EchoServer) RegisterGroups(groups ...*Group) {
	for _, g := range groups {
		echoGroup := s.app.Group(g.prefix, g.middleware...)
		for _, r := range g.routes {
			echoGroup.Add(r.method, r.path, r.handler.Handle, r.middleware...)
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
