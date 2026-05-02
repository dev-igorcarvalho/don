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
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidDriver          = errors.New("database driver is required")
	ErrInvalidDSN             = errors.New("database DSN is required")
	ErrInvalidMaxOpenConns    = errors.New("max open connections cannot be negative")
	ErrInvalidMaxIdleConns    = errors.New("max idle connections cannot be negative")
	ErrInvalidConnMaxLifetime = errors.New("conn max lifetime cannot be negative")
	ErrInvalidConnMaxIdleTime = errors.New("conn max idle time cannot be negative")
	ErrInvalidConnectTimeout  = errors.New("connect timeout cannot be negative")
)

// Config holds the database connection settings.
type Config struct {
	Driver                 string
	DSN                    string
	MaxOpenConnections     int
	MaxIdleConnections     int
	ConnectionsMaxLifetime time.Duration
	ConnectionsMaxIdleTime time.Duration
	Warmup                 bool
	ConnectTimeout         time.Duration
}

// NewConfig creates a new database configuration and validates it.
func NewConfig(
	driver string,
	dsn string,
	maxOpenConns int,
	maxIdleConns int,
	connMaxLifetime time.Duration,
	connMaxIdleTime time.Duration,
	warmup bool,
	connectTimeout time.Duration,
) (Config, error) {
	cfg := Config{
		Driver:                 driver,
		DSN:                    dsn,
		MaxOpenConnections:     maxOpenConns,
		MaxIdleConnections:     maxIdleConns,
		ConnectionsMaxLifetime: connMaxLifetime,
		ConnectionsMaxIdleTime: connMaxIdleTime,
		Warmup:                 warmup,
		ConnectTimeout:         connectTimeout,
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// Validate ensures the database configuration is valid.
func (c Config) Validate() error {
	if c.Driver == "" {
		return ErrInvalidDriver
	}
	if c.DSN == "" {
		return ErrInvalidDSN
	}
	if c.MaxOpenConnections < 0 {
		return ErrInvalidMaxOpenConns
	}
	if c.MaxIdleConnections < 0 {
		return ErrInvalidMaxIdleConns
	}
	if c.ConnectionsMaxLifetime < 0 {
		return ErrInvalidConnMaxLifetime
	}
	if c.ConnectionsMaxIdleTime < 0 {
		return ErrInvalidConnMaxIdleTime
	}
	if c.ConnectTimeout < 0 {
		return ErrInvalidConnectTimeout
	}
	return nil
}

// SQLPair manages a pair of database connections for read/write splitting.
type SQLPair struct {
	Writer *sql.DB
	Reader *sql.DB
}

// NewSQL creates a new sql.DB using the provided configuration.
func NewSQL(cfg Config) (*sql.DB, error) {

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("database driver config is invalid: %w", err)
	}

	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConnections)
	db.SetMaxIdleConns(cfg.MaxIdleConnections)
	db.SetConnMaxLifetime(cfg.ConnectionsMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnectionsMaxIdleTime)

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

	return errors.Join(errs...)
}

// Shutdown gracefully shuts down the database connections.
func (p *SQLPair) Shutdown(_ context.Context) error {
	return p.Close()
}
