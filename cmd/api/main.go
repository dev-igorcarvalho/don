package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/dev-igorcarvalho/don/internal/config"
	"github.com/dev-igorcarvalho/don/internal/handlers"
	pkgConfig "github.com/dev-igorcarvalho/don/pkg/config"
	"github.com/dev-igorcarvalho/don/pkg/lifecycle"
	"github.com/dev-igorcarvalho/don/pkg/logger"
	"github.com/dev-igorcarvalho/don/pkg/must"
	"github.com/dev-igorcarvalho/don/pkg/server"
	"github.com/dev-igorcarvalho/don/pkg/server/echoserver"
)

const shutdownTimeout = 10 * time.Second

func main() {
	must.Succeed(run())
}

func run() error {
	ctx := context.Background()

	cfg := must.Get(pkgConfig.Load[config.AppConfig]())
	logger.Setup(logger.Environment(cfg.Environment))

	lm := lifecycle.NewManager(shutdownTimeout)
	srv := wireServer(cfg)

	// Register server for graceful shutdown
	lm.Register("api-server", srv)

	// Start server
	go func() {
		logger.Info(ctx, "starting server", slog.String("port", cfg.HTTPPort))
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(ctx, "server error", slog.Any("error", err))
		}
	}()

	// Wait for shutdown signal
	return lm.Wait(ctx)
}

func wireServer(cfg *config.AppConfig) server.Server {
	healthHandler := handlers.NewHealthHandler()

	srv := echoserver.New(
		echoserver.WithPort(cfg.HTTPPort),
		echoserver.WithHealthCheck(healthHandler),
		echoserver.WithErrorHandler(echoserver.DefaultErrorHandler),
	)

	return srv
}
