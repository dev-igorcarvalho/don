// ---
// title: SQL Database Initialization
// description: Generic constructor for sql.DB and SQLPair for primary/replica setups.
// last_updated: 2026-05-02
// type: Implementation
// ---

package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Config holds the database connection settings.
type Config struct {
	Driver          string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	Warmup          bool
	ConnectTimeout  time.Duration
}

// SQLPair manages a pair of database connections for read/write splitting.
type SQLPair struct {
	Writer *sql.DB
	Reader *sql.DB
}

// NewSQL creates a new sql.DB using the provided configuration.
func NewSQL(cfg Config) (*sql.DB, error) {
	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	}

	if cfg.Warmup {
		timeout := cfg.ConnectTimeout
		if timeout == 0 {
			timeout = 5 * time.Second
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("failed to ping database: %w", err)
		}
	}

	return db, nil
}

// NewSQLPair creates a new SQLPair using the provided writer and reader configurations.
func NewSQLPair(writerCfg, readerCfg Config) (*SQLPair, error) {
	writer, err := NewSQL(writerCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize writer: %w", err)
	}

	reader, err := NewSQL(readerCfg)
	if err != nil {
		_ = writer.Close()
		return nil, fmt.Errorf("failed to initialize reader: %w", err)
	}

	return &SQLPair{
		Writer: writer,
		Reader: reader,
	}, nil
}

// Close closes both writer and reader connections.
func (p *SQLPair) Close() error {
	var errs []error
	if p.Writer != nil {
		if err := p.Writer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close writer: %w", err))
		}
	}
	if p.Reader != nil {
		if err := p.Reader.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close reader: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close SQLPair: %v", errs)
	}
	return nil
}
