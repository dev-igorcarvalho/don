package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	closeCount int
)

// mockDriver is a dummy driver for testing purposes.
type mockDriver struct{}

func (m *mockDriver) Open(name string) (driver.Conn, error) {
	if name == "fail" {
		return nil, errors.New("connection failed")
	}
	return &mockConn{name: name}, nil
}

type mockConn struct {
	name string
}

func (m *mockConn) Prepare(query string) (driver.Stmt, error) { return nil, nil }
func (m *mockConn) Close() error {
	closeCount++
	if m.name == "close_fail" {
		return errors.New("close failed")
	}
	return nil
}
func (m *mockConn) Begin() (driver.Tx, error) { return nil, nil }

// Ping support for mockConn
func (m *mockConn) Ping(ctx context.Context) error {
	if m.name == "ping_fail" {
		return errors.New("ping failed")
	}
	return nil
}

func init() {
	sql.Register("mock", &mockDriver{})
}

func TestNewConfig(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		cfg, err := NewConfig(
			"postgres",
			"postgres://user:pass@localhost:5432/db",
			10,
			5,
			time.Hour,
			time.Minute,
			true,
			5*time.Second,
		)

		require.NoError(t, err)
		assert.Equal(t, "postgres", cfg.Driver)
		assert.Equal(t, 10, cfg.MaxOpenConnections)
		assert.Equal(t, 5, cfg.MaxIdleConnections)
		assert.True(t, cfg.Warmup)
	})

	t.Run("invalid configuration", func(t *testing.T) {
		_, err := NewConfig("", "", -1, -1, -1, -1, false, -1)
		assert.Error(t, err)
	})
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr error
	}{
		{
			name: "valid config",
			config: Config{
				Driver:                 "postgres",
				DSN:                    "dsn",
				MaxOpenConnections:     10,
				MaxIdleConnections:     5,
				ConnectionsMaxLifetime: time.Hour,
				ConnectionsMaxIdleTime: time.Minute,
				ConnectTimeout:         time.Second,
			},
			wantErr: nil,
		},
		{
			name: "missing driver",
			config: Config{
				DSN: "dsn",
			},
			wantErr: ErrInvalidDriver,
		},
		{
			name: "missing dsn",
			config: Config{
				Driver: "postgres",
			},
			wantErr: ErrInvalidDSN,
		},
		{
			name: "negative max open connections",
			config: Config{
				Driver:             "postgres",
				DSN:                "dsn",
				MaxOpenConnections: -1,
			},
			wantErr: ErrInvalidMaxOpenConns,
		},
		{
			name: "negative max idle connections",
			config: Config{
				Driver:             "postgres",
				DSN:                "dsn",
				MaxOpenConnections: 10,
				MaxIdleConnections: -1,
			},
			wantErr: ErrInvalidMaxIdleConns,
		},
		{
			name: "negative connections max lifetime",
			config: Config{
				Driver:                 "postgres",
				DSN:                    "dsn",
				MaxOpenConnections:     10,
				MaxIdleConnections:     5,
				ConnectionsMaxLifetime: -1,
			},
			wantErr: ErrInvalidConnMaxLifetime,
		},
		{
			name: "negative connections max idle time",
			config: Config{
				Driver:                 "postgres",
				DSN:                    "dsn",
				MaxOpenConnections:     10,
				MaxIdleConnections:     5,
				ConnectionsMaxLifetime: time.Hour,
				ConnectionsMaxIdleTime: -1,
			},
			wantErr: ErrInvalidConnMaxIdleTime,
		},
		{
			name: "negative connect timeout",
			config: Config{
				Driver:                 "postgres",
				DSN:                    "dsn",
				MaxOpenConnections:     10,
				MaxIdleConnections:     5,
				ConnectionsMaxLifetime: time.Hour,
				ConnectionsMaxIdleTime: time.Minute,
				ConnectTimeout:         -1,
			},
			wantErr: ErrInvalidConnectTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

var validCfg = Config{
	Driver:                 "mock",
	DSN:                    "success",
	MaxOpenConnections:     10,
	MaxIdleConnections:     5,
	ConnectionsMaxLifetime: time.Hour,
	ConnectionsMaxIdleTime: time.Minute,
	ConnectTimeout:         time.Second,
}

func TestNewSQL(t *testing.T) {
	t.Run("invalid config", func(t *testing.T) {
		cfg := Config{}
		db, err := NewSQL(context.Background(), cfg)
		assert.Nil(t, db)
		assert.ErrorIs(t, err, ErrInvalidDriver)
	})

	t.Run("invalid driver", func(t *testing.T) {
		cfg := validCfg
		cfg.Driver = "invalid"

		db, err := NewSQL(context.Background(), cfg)
		assert.Nil(t, db)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sql: unknown driver")
	})

	t.Run("successful connection without warmup", func(t *testing.T) {
		cfg := validCfg
		db, err := NewSQL(context.Background(), cfg)
		require.NoError(t, err)
		require.NotNil(t, db)
		_ = db.Close()
	})

	t.Run("successful connection with warmup", func(t *testing.T) {
		cfg := validCfg
		cfg.Warmup = true

		db, err := NewSQL(context.Background(), cfg)
		require.NoError(t, err)
		require.NotNil(t, db)
		_ = db.Close()
	})

	t.Run("failed ping with warmup", func(t *testing.T) {
		cfg := validCfg
		cfg.DSN = "ping_fail"
		cfg.Warmup = true

		db, err := NewSQL(context.Background(), cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to ping database")
		assert.Nil(t, db)
	})

	t.Run("successful connection with default warmup timeout", func(t *testing.T) {
		cfg := validCfg
		cfg.Warmup = true
		cfg.ConnectTimeout = 0

		db, err := NewSQL(context.Background(), cfg)
		require.NoError(t, err)
		require.NotNil(t, db)
		_ = db.Close()
	})
}

func TestNewSQLPair(t *testing.T) {
	t.Run("successful pair creation", func(t *testing.T) {
		cfg := validCfg
		pair, err := NewSQLPair(context.Background(), cfg, cfg)
		require.NoError(t, err)
		require.NotNil(t, pair)
		assert.NotNil(t, pair.Writer())
		assert.NotNil(t, pair.Reader())
		_ = pair.Close()
	})

	t.Run("failed writer creation", func(t *testing.T) {
		writerCfg := Config{}
		readerCfg := validCfg
		pair, err := NewSQLPair(context.Background(), writerCfg, readerCfg)
		assert.Error(t, err)
		assert.Nil(t, pair)
		assert.Contains(t, err.Error(), "failed to initialize writer")
	})

	t.Run("failed reader creation", func(t *testing.T) {
		writerCfg := validCfg
		readerCfg := Config{}
		pair, err := NewSQLPair(context.Background(), writerCfg, readerCfg)
		assert.Error(t, err)
		assert.Nil(t, pair)
		assert.Contains(t, err.Error(), "failed to initialize reader")
	})

	t.Run("reader failure cleanup", func(t *testing.T) {
		closeCount = 0
		writerCfg := validCfg
		writerCfg.Warmup = true
		writerCfg.MaxIdleConnections = 1

		readerCfg := validCfg
		readerCfg.Driver = "invalid"

		pair, err := NewSQLPair(context.Background(), writerCfg, readerCfg)
		assert.Error(t, err)
		assert.Nil(t, pair)
		assert.Equal(t, 1, closeCount, "writer should have been closed")
	})
}

func TestSQLPair_Close(t *testing.T) {
	t.Run("successful close", func(t *testing.T) {
		cfg := validCfg
		pair, err := NewSQLPair(context.Background(), cfg, cfg)
		require.NoError(t, err)

		err = pair.Close()
		assert.NoError(t, err)
	})

	t.Run("failed writer close", func(t *testing.T) {
		writerCfg := validCfg
		writerCfg.DSN = "close_fail"
		writerCfg.Warmup = true
		writerCfg.MaxIdleConnections = 1

		readerCfg := validCfg
		readerCfg.Warmup = true
		readerCfg.MaxIdleConnections = 1

		pair, err := NewSQLPair(context.Background(), writerCfg, readerCfg)
		require.NoError(t, err)

		err = pair.Close()
		assert.ErrorContains(t, err, "failed to close writer")
	})

	t.Run("failed reader close", func(t *testing.T) {
		writerCfg := validCfg
		writerCfg.Warmup = true
		writerCfg.MaxIdleConnections = 1

		readerCfg := validCfg
		readerCfg.DSN = "close_fail"
		readerCfg.Warmup = true
		readerCfg.MaxIdleConnections = 1

		pair, err := NewSQLPair(context.Background(), writerCfg, readerCfg)
		require.NoError(t, err)

		err = pair.Close()
		assert.ErrorContains(t, err, "failed to close reader")
	})

	t.Run("both fail to close", func(t *testing.T) {
		cfg := validCfg
		cfg.DSN = "close_fail"
		cfg.Warmup = true
		cfg.MaxIdleConnections = 1

		pair, err := NewSQLPair(context.Background(), cfg, cfg)
		require.NoError(t, err)

		err = pair.Close()
		assert.ErrorContains(t, err, "failed to close writer")
		assert.ErrorContains(t, err, "failed to close reader")
	})

	t.Run("nil connections", func(t *testing.T) {
		pair := &SQLPair{}
		err := pair.Close()
		assert.NoError(t, err)
	})
}

func TestSQLPair_Shutdown(t *testing.T) {
	cfg := validCfg
	pair, err := NewSQLPair(context.Background(), cfg, cfg)
	require.NoError(t, err)

	err = pair.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestSQLPair_Ping(t *testing.T) {
	t.Run("successful ping", func(t *testing.T) {
		cfg := validCfg
		pair, err := NewSQLPair(context.Background(), cfg, cfg)
		require.NoError(t, err)
		defer pair.Close()

		err = pair.Ping(context.Background())
		assert.NoError(t, err)
	})

	t.Run("writer ping fail", func(t *testing.T) {
		writerCfg := validCfg
		writerCfg.DSN = "ping_fail"
		readerCfg := validCfg

		pair, err := NewSQLPair(context.Background(), writerCfg, readerCfg)
		require.NoError(t, err)
		defer pair.Close()

		err = pair.Ping(context.Background())
		assert.ErrorContains(t, err, "writer ping failed")
	})

	t.Run("reader ping fail", func(t *testing.T) {
		writerCfg := validCfg
		readerCfg := validCfg
		readerCfg.DSN = "ping_fail"

		pair, err := NewSQLPair(context.Background(), writerCfg, readerCfg)
		require.NoError(t, err)
		defer pair.Close()

		err = pair.Ping(context.Background())
		assert.ErrorContains(t, err, "reader ping failed")
	})

	t.Run("both ping fail", func(t *testing.T) {
		cfg := validCfg
		cfg.DSN = "ping_fail"

		pair, err := NewSQLPair(context.Background(), cfg, cfg)
		require.NoError(t, err)
		defer pair.Close()

		err = pair.Ping(context.Background())
		assert.ErrorContains(t, err, "writer ping failed")
		assert.ErrorContains(t, err, "reader ping failed")
	})
}

func TestSQLPair_HealthCheck(t *testing.T) {
	t.Run("successful health check", func(t *testing.T) {
		cfg := validCfg
		pair, err := NewSQLPair(context.Background(), cfg, cfg)
		require.NoError(t, err)
		defer pair.Close()

		status := pair.HealthCheck(context.Background())
		assert.True(t, status.WriterAlive)
		assert.True(t, status.ReaderAlive)
		assert.Equal(t, "OK", status.Message)
		// mock driver stats are usually 0 unless we do work, but we verify they exist
		assert.GreaterOrEqual(t, status.OpenConns, 0)
		assert.GreaterOrEqual(t, status.IdleConns, 0)
	})

	t.Run("writer unhealthy", func(t *testing.T) {
		writerCfg := validCfg
		writerCfg.DSN = "ping_fail"
		readerCfg := validCfg

		pair, err := NewSQLPair(context.Background(), writerCfg, readerCfg)
		require.NoError(t, err)
		defer pair.Close()

		status := pair.HealthCheck(context.Background())
		assert.False(t, status.WriterAlive)
		assert.True(t, status.ReaderAlive)
		assert.Contains(t, status.Message, "writer")
		assert.NotEqual(t, "OK", status.Message)
	})

	t.Run("reader unhealthy", func(t *testing.T) {
		writerCfg := validCfg
		readerCfg := validCfg
		readerCfg.DSN = "ping_fail"

		pair, err := NewSQLPair(context.Background(), writerCfg, readerCfg)
		require.NoError(t, err)
		defer pair.Close()

		status := pair.HealthCheck(context.Background())
		assert.True(t, status.WriterAlive)
		assert.False(t, status.ReaderAlive)
		assert.Contains(t, status.Message, "reader")
	})

	t.Run("both unhealthy", func(t *testing.T) {
		cfg := validCfg
		cfg.DSN = "ping_fail"

		pair, err := NewSQLPair(context.Background(), cfg, cfg)
		require.NoError(t, err)
		defer pair.Close()

		status := pair.HealthCheck(context.Background())
		assert.False(t, status.WriterAlive)
		assert.False(t, status.ReaderAlive)
		assert.Contains(t, status.Message, "writer")
		assert.Contains(t, status.Message, "reader")
	})
}

func TestSQLPair_Stats(t *testing.T) {
	cfg := validCfg
	pair, err := NewSQLPair(context.Background(), cfg, cfg)
	require.NoError(t, err)
	defer pair.Close()

	stats := pair.Stats()
	assert.NotNil(t, stats.Writer)
	assert.NotNil(t, stats.Reader)
}
