package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	dsnCloseFail = "close_fail"
	dsnPingFail  = "ping_fail"
	dsnExecFail  = "exec_fail"
	dsnQueryFail = "query_fail"
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

func (m *mockConn) Prepare(query string) (driver.Stmt, error) { return &mockStmt{}, nil }
func (m *mockConn) Close() error {
	closeCount++
	if m.name == dsnCloseFail {
		return errors.New("close failed")
	}
	return nil
}
func (m *mockConn) Begin() (driver.Tx, error) { return nil, nil }

// Exec support for mockConn
func (m *mockConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	if m.name == dsnExecFail {
		return nil, errors.New("exec failed")
	}
	return driver.RowsAffected(1), nil
}

// Query support for mockConn
func (m *mockConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	if m.name == dsnQueryFail {
		return nil, errors.New("query failed")
	}
	return &mockRows{}, nil
}

type mockStmt struct{}

func (s *mockStmt) Close() error                                    { return nil }
func (s *mockStmt) NumInput() int                                   { return -1 }
func (s *mockStmt) Exec(args []driver.Value) (driver.Result, error) { return nil, nil }
func (s *mockStmt) Query(args []driver.Value) (driver.Rows, error)  { return nil, nil }

type mockRows struct{}

func (r *mockRows) Columns() []string              { return []string{"col1"} }
func (r *mockRows) Close() error                   { return nil }
func (r *mockRows) Next(dest []driver.Value) error { return io.EOF }

// Ping support for mockConn
func (m *mockConn) Ping(ctx context.Context) error {
	if m.name == dsnPingFail {
		return errors.New("ping failed")
	}
	return nil
}

func init() {
	sql.Register("mock", &mockDriver{})
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
		db, err := newSQL(context.Background(), cfg)
		assert.Nil(t, db)
		assert.ErrorIs(t, err, ErrInvalidDriver)
	})

	t.Run("invalid driver", func(t *testing.T) {
		cfg := validCfg
		cfg.Driver = "invalid"

		db, err := newSQL(context.Background(), cfg)
		assert.Nil(t, db)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sql: unknown driver")
	})

	t.Run("successful connection without warmup", func(t *testing.T) {
		cfg := validCfg
		db, err := newSQL(context.Background(), cfg)
		require.NoError(t, err)
		require.NotNil(t, db)
		_ = db.Close()
	})

	t.Run("successful connection with warmup", func(t *testing.T) {
		cfg := validCfg
		cfg.Warmup = true

		db, err := newSQL(context.Background(), cfg)
		require.NoError(t, err)
		require.NotNil(t, db)
		_ = db.Close()
	})

	t.Run("failed ping with warmup", func(t *testing.T) {
		cfg := validCfg
		cfg.DSN = dsnPingFail
		cfg.Warmup = true

		db, err := newSQL(context.Background(), cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to ping database")
		assert.Nil(t, db)
	})

	t.Run("successful connection with default timeout", func(t *testing.T) {
		cfg := validCfg
		cfg.Warmup = true
		cfg.ConnectTimeout = 0

		db, err := newSQL(context.Background(), cfg)
		require.NoError(t, err)
		require.NotNil(t, db)
		_ = db.Close()
	})

	t.Run("warmup full retry exhaustion", func(t *testing.T) {
		// Save and restore
		oldRetries := WarmupMaxRetries
		oldBase := WarmupBaseDelay
		oldMax := WarmupMaxDelay
		defer func() {
			WarmupMaxRetries = oldRetries
			WarmupBaseDelay = oldBase
			WarmupMaxDelay = oldMax
		}()

		WarmupMaxRetries = 2
		WarmupBaseDelay = 1 * time.Millisecond
		WarmupMaxDelay = 2 * time.Millisecond

		cfg := validCfg
		cfg.DSN = dsnPingFail
		cfg.Warmup = true

		db, err := newSQL(context.Background(), cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to ping database after 3 attempts")
		assert.Nil(t, db)
	})

	t.Run("warmup timeout", func(t *testing.T) {
		cfg := validCfg
		cfg.DSN = dsnPingFail
		cfg.Warmup = true
		cfg.ConnectTimeout = 10 * time.Millisecond

		db, err := newSQL(context.Background(), cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout")
		assert.Nil(t, db)
	})
}

func TestNewSQLPair(t *testing.T) {
	t.Run("successful pair creation", func(t *testing.T) {
		cfg := validCfg
		pair, err := NewClient(context.Background(), WithWriter(cfg), WithReader(cfg))
		require.NoError(t, err)
		require.NotNil(t, pair)
		assert.NotNil(t, pair.Writer())
		assert.NotNil(t, pair.Reader())
		_ = pair.Close()
	})

	t.Run("failed writer creation", func(t *testing.T) {
		writerCfg := Config{}
		readerCfg := validCfg
		pair, err := NewClient(context.Background(), WithWriter(writerCfg), WithReader(readerCfg))
		assert.Error(t, err)
		assert.Nil(t, pair)
		assert.Contains(t, err.Error(), "failed to initialize writer")
	})

	t.Run("successful writer only creation", func(t *testing.T) {
		cfg := validCfg
		pair, err := NewClient(context.Background(), WithWriter(cfg))
		require.NoError(t, err)
		require.NotNil(t, pair)
		assert.NotNil(t, pair.Writer())
		assert.Nil(t, pair.Reader())
		_ = pair.Close()
	})

	t.Run("failed no options", func(t *testing.T) {
		pair, err := NewClient(context.Background())
		assert.ErrorIs(t, err, ErrNoOptions)
		assert.Nil(t, pair)
	})

	t.Run("failed no writer", func(t *testing.T) {
		cfg := validCfg
		pair, err := NewClient(context.Background(), WithReader(cfg))
		assert.ErrorIs(t, err, ErrWriterRequired)
		assert.Nil(t, pair)
	})

	t.Run("failed reader creation", func(t *testing.T) {
		writerCfg := validCfg
		readerCfg := Config{}
		pair, err := NewClient(context.Background(), WithWriter(writerCfg), WithReader(readerCfg))
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

		pair, err := NewClient(context.Background(), WithWriter(writerCfg), WithReader(readerCfg))
		assert.Error(t, err)
		assert.Nil(t, pair)
		assert.Equal(t, 1, closeCount, "writer should have been closed")
	})
}

func TestSQLPair_Close(t *testing.T) {
	t.Run("successful close", func(t *testing.T) {
		cfg := validCfg
		pair, err := NewClient(context.Background(), WithWriter(cfg), WithReader(cfg))
		require.NoError(t, err)

		err = pair.Close()
		assert.NoError(t, err)
	})

	t.Run("failed writer close", func(t *testing.T) {
		writerCfg := validCfg
		writerCfg.DSN = dsnCloseFail
		writerCfg.Warmup = true
		writerCfg.MaxIdleConnections = 1

		readerCfg := validCfg
		readerCfg.Warmup = true
		readerCfg.MaxIdleConnections = 1

		pair, err := NewClient(context.Background(), WithWriter(writerCfg), WithReader(readerCfg))
		require.NoError(t, err)

		err = pair.Close()
		assert.ErrorContains(t, err, "failed to close writer")
	})

	t.Run("failed reader close", func(t *testing.T) {
		writerCfg := validCfg
		writerCfg.Warmup = true
		writerCfg.MaxIdleConnections = 1

		readerCfg := validCfg
		readerCfg.DSN = dsnCloseFail
		readerCfg.Warmup = true
		readerCfg.MaxIdleConnections = 1

		pair, err := NewClient(context.Background(), WithWriter(writerCfg), WithReader(readerCfg))
		require.NoError(t, err)

		err = pair.Close()
		assert.ErrorContains(t, err, "failed to close reader")
	})

	t.Run("both fail to close", func(t *testing.T) {
		cfg := validCfg
		cfg.DSN = dsnCloseFail
		cfg.Warmup = true
		cfg.MaxIdleConnections = 1

		pair, err := NewClient(context.Background(), WithWriter(cfg), WithReader(cfg))
		require.NoError(t, err)

		err = pair.Close()
		assert.ErrorContains(t, err, "failed to close writer")
		assert.ErrorContains(t, err, "failed to close reader")
	})

	t.Run("nil connections", func(t *testing.T) {
		pair := &Client{}
		err := pair.Close()
		assert.NoError(t, err)
	})
}

func TestSQLPair_Shutdown(t *testing.T) {
	cfg := validCfg
	pair, err := NewClient(context.Background(), WithWriter(cfg), WithReader(cfg))
	require.NoError(t, err)

	err = pair.Shutdown(context.Background())
	assert.NoError(t, err)
}
