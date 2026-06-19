// ---
// title: Migration Entry Point
// description: Standalone entry point for running database migrations with optional seeding support.
// last_updated: 2026-05-09
// type: EntryPoint
// ---

package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/dev-igorcarvalho/don/internal/config"
	pkgConfig "github.com/dev-igorcarvalho/don/pkg/config"
	"github.com/dev-igorcarvalho/don/pkg/database"
	"github.com/dev-igorcarvalho/don/pkg/logger"
	"github.com/dev-igorcarvalho/don/pkg/migrator"
	"github.com/dev-igorcarvalho/don/pkg/must"
)

func main() {
	must.Succeed(run())
}

func run() error {
	ctx := context.Background()

	seed := flag.Bool("seed", false, "Run database seeds after migrations")
	flag.Parse()

	appCfg := must.Get(pkgConfig.Load[config.AppConfig]())
	dbCfg := must.Get(pkgConfig.Load[config.SqlDbConfig]())

	logger.Setup(logger.Environment(appCfg.Environment))

	logger.Info(ctx, "starting migrations")

	dbWriterConfig := must.Get(dbCfg.Writer.ToSqlConnectorConfig())

	sqlPair := must.Get(database.NewClient(ctx,
		database.WithWriter(dbWriterConfig),
	))
	defer func() {
		if err := sqlPair.Close(); err != nil {
			logger.Error(ctx, "failed to close database connection", slog.Any("error", err))
		}
	}()

	must.Succeed(migrator.MigrateUp(
		sqlPair.Writer(),
		appCfg.Name,
		migrator.PostgresFactory,
		"scripts/database/migrations",
	))

	logger.Info(ctx, "migrations completed successfully")

	// Run Seeds if requested
	if *seed {
		logger.Info(ctx, "starting database seeding")
		must.Succeed(migrator.RunSeeds(
			ctx,
			sqlPair.Writer(),
			os.DirFS("."),
			"scripts/database/seeds",
		))
		logger.Info(ctx, "seeding completed successfully")
	}

	return nil
}
