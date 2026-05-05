// ---
// title: SQL Database Initialization
// description: Generic constructor for sql.DB with exponential backoff and advanced health probes.
// last_updated: 2026-05-05
// type: Implementation
// ---

package database

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand/v2"
	"time"
)

// newSQL creates a new sql.DB using the provided configuration.
func newSQL(ctx context.Context, cfg Config) (*sql.DB, error) {

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("database driver config is invalid: %w", err)
	}

	timeout := cfg.ConnectTimeout
	if timeout == 0 {
		timeout = DefaultConnectTimeout
	}
	initCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConnections)
	db.SetMaxIdleConns(cfg.MaxIdleConnections)
	db.SetConnMaxLifetime(cfg.ConnectionsMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnectionsMaxIdleTime)

	if cfg.Warmup {
		var lastErr error
		for attempt := 0; attempt <= WarmupMaxRetries; attempt++ {
			if err := db.PingContext(initCtx); err == nil {
				return db, nil
			}
			lastErr = err

			if attempt == WarmupMaxRetries {
				break
			}

			// Calculate exponential backoff: delay = BaseDelay * 2^attempt
			delay := time.Duration(1<<attempt) * WarmupBaseDelay
			if delay > WarmupMaxDelay {
				delay = WarmupMaxDelay
			}

			// Apply jitter: +/- 20%
			jitter := time.Duration(rand.Float64() * 0.2 * float64(delay))
			if rand.IntN(2) == 0 {
				delay += jitter
			} else {
				delay -= jitter
			}

			select {
			case <-initCtx.Done():
				_ = db.Close()
				return nil, fmt.Errorf("failed to ping database (timeout): %w", initCtx.Err())
			case <-time.After(delay):
			}
		}

		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database after %d attempts: %w", WarmupMaxRetries+1, lastErr)
	}
	return db, nil
}
