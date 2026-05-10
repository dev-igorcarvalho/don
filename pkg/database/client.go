// ---
// title: SQL Connection Client
// description: Manages a pair of SQL database connections to support read/write splitting and lifecycle management.
// last_updated: 2026-05-08
// type: Implementation
// ---

package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"
)

var (
	// DefaultConnectTimeout is the default duration to wait for a database connection to be established.
	DefaultConnectTimeout = 5 * time.Second
	// DefaultQueryTimeout is the default duration to wait for a database query to complete.
	DefaultQueryTimeout = 30 * time.Second

	// WarmupMaxRetries is the maximum number of times to retry the database ping during warmup.
	WarmupMaxRetries = 5
	// WarmupBaseDelay is the initial delay between retries.
	WarmupBaseDelay = 500 * time.Millisecond
	// WarmupMaxDelay is the maximum delay between retries.
	WarmupMaxDelay = 5 * time.Second
)

var (
	ErrNoOptions      = errors.New("at least one database option must be provided")
	ErrWriterRequired = errors.New("writer configuration is required (use WithWriter)")
)

// Client manages a pair of database connections for read/write splitting.
type Client struct {
	writer        *sql.DB
	reader        *sql.DB
	writerTimeout time.Duration
	readerTimeout time.Duration
}

// Writer returns the writer database connection.
func (p *Client) Writer() *sql.DB {
	return p.writer
}

// Reader returns the reader database connection.
func (p *Client) Reader() *sql.DB {
	return p.reader
}

// NewClient creates a new Client using the provided options.
func NewClient(ctx context.Context, opts ...Option) (*Client, error) {
	if len(opts) == 0 {
		return nil, ErrNoOptions
	}

	var cfg options
	for _, opt := range opts {
		opt(&cfg)
	}

	// Writer is mandatory
	if cfg.writer == nil {
		return nil, ErrWriterRequired
	}

	writerTimeout := cfg.writer.QueryTimeout
	if writerTimeout == 0 {
		writerTimeout = DefaultQueryTimeout
	}

	client := &Client{
		writerTimeout: writerTimeout,
	}

	// Initialize Writer
	writer, err := newSQL(ctx, *cfg.writer)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize writer: %w", err)
	}
	client.writer = writer

	// Initialize Reader if provided
	if cfg.reader != nil {
		readerTimeout := cfg.reader.QueryTimeout
		if readerTimeout == 0 {
			readerTimeout = DefaultQueryTimeout
		}
		client.readerTimeout = readerTimeout

		reader, err := newSQL(ctx, *cfg.reader)
		if err != nil {
			// Clean up writer if reader fails
			_ = client.writer.Close()
			return nil, fmt.Errorf("failed to initialize reader: %w", err)
		}
		client.reader = reader
	}

	return client, nil
}

// Close closes both writer and reader connections.
func (p *Client) Close() error {
	var errs []error
	if p.writer != nil {
		if err := p.writer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close writer: %w", err))
		}
		p.writer = nil
	}
	if p.reader != nil {
		if err := p.reader.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close reader: %w", err))
		}
		p.reader = nil
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// Shutdown gracefully shuts down the database connections.
func (p *Client) Shutdown(_ context.Context) error {
	return p.Close()
}

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
