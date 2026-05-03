// ---
// title: API Entry Point
// description: Main entry point for the REST API service, responsible for configuration loading, server wiring, and graceful shutdown.
// last_updated: 2026-05-03
// type: EntryPoint
// ---

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
	"github.com/dev-igorcarvalho/don/pkg/database"
	"github.com/dev-igorcarvalho/don/pkg/lifecycle"
	"github.com/dev-igorcarvalho/don/pkg/logger"
	"github.com/dev-igorcarvalho/don/pkg/must"
	"github.com/dev-igorcarvalho/don/pkg/server"
	"github.com/dev-igorcarvalho/don/pkg/server/echoserver"
	"github.com/labstack/echo/v4"
)

const shutdownTimeout = 10 * time.Second

func main() {
	must.Succeed(run())
}

func run() error {
	ctx := context.Background()

	appCfg := must.Get(pkgConfig.Load[config.AppConfig]())
	dbCfg := must.Get(pkgConfig.Load[config.SqlDbConfig]())
	logger.Setup(logger.Environment(appCfg.Environment))

	// Initialize Database
	dbWriterConfig := must.Get(dbCfg.Writer.ToSqlConnectorConfig())
	dbReaderConfig := must.Get(dbCfg.Reader.ToSqlConnectorConfig())
	sqlPair := must.Get(database.NewSQLPair(ctx, dbWriterConfig, dbReaderConfig))

	lm := lifecycle.NewManager(shutdownTimeout)
	srv := wireServer(ctx, appCfg)

	// Register components for graceful shutdown
	lm.Register("database", sqlPair)
	lm.Register("api-server", srv)

	// Start server
	go func() {
		logger.Info(ctx, "starting server", slog.String("port", appCfg.HTTPPort))
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(ctx, "server error", slog.Any("error", err))
		}
	}()

	// Wait for shutdown signal
	return lm.Wait(ctx)
}

func wireServer(ctx context.Context, cfg *config.AppConfig) server.Server {
	healthHandler := handlers.NewHealthHandler()

	middlewares := []echo.MiddlewareFunc{
		echoserver.LoggerMiddleware(),
		echoserver.RecoveryMiddleware(),
		echoserver.SecurityHeadersMiddleware(echo.MIMEApplicationJSON),
	}

	if cfg.RateLimitEnabled {
		middlewares = append(middlewares, echoserver.RateLimitMiddleware(ctx, cfg.RateLimitRPS, cfg.RateLimitBurst))
	}

	srv := echoserver.New(
		echoserver.WithPort(cfg.HTTPPort),
		echoserver.WithHealthCheck(healthHandler),
		echoserver.WithErrorHandler(echoserver.DefaultErrorHandler),
		echoserver.WithMiddleware(middlewares...),
	)

	return srv
}
