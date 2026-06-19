// ---
// title: Database Configuration
// description: Defines the configuration structure and validation logic for database connections.
// last_updated: 2026-05-08
// type: Configuration
// ---

package database

import (
	"errors"
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
	ErrInvalidQueryTimeout    = errors.New("query timeout cannot be negative")
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
	QueryTimeout           time.Duration
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
	queryTimeout time.Duration,
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
		QueryTimeout:           queryTimeout,
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
	// Note: 0 is used by database/sql for "unlimited" or "no limit".
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
	// Note: 0 triggers DefaultConnectTimeout in newSQL.
	if c.ConnectTimeout < 0 {
		return ErrInvalidConnectTimeout
	}
	if c.QueryTimeout < 0 {
		return ErrInvalidQueryTimeout
	}
	return nil
}
