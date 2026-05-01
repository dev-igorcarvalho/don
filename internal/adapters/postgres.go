package adapters

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dev-igorcarvalho/don/internal/config"
	"github.com/dev-igorcarvalho/don/pkg/logger"
)

// NewPostgresDB creates a new postgres database connection pool
func NewPostgresDB(cfg config.DBConfig) (*sql.DB, error) {
	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	if cfg.Warmup {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			return nil, fmt.Errorf("failed to ping database: %w", err)
		}
		logger.Info(ctx, "database connection established and warmed up")
	}

	return db, nil
}
