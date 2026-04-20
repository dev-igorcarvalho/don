package adapters

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dev-igorcarvalho/don/internal/config"
	"github.com/dev-igorcarvalho/don/internal/handlers"
	"github.com/labstack/echo/v4"
)

type HTTPServer struct {
	echo *echo.Echo
	cfg  config.AppConfig
}

func NewHTTPServer(cfg config.AppConfig, healthHandler *handlers.HealthHandler) *HTTPServer {
	e := echo.New()

	// Routes
	e.GET("/health", healthHandler.Handle)

	return &HTTPServer{
		echo: e,
		cfg:  cfg,
	}
}

func (s *HTTPServer) Start() error {
	addr := fmt.Sprintf(":%s", s.cfg.HTTPPort)
	if err := s.echo.Start(addr); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}
