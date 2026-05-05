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

const (
	dsnCloseFail = "close_fail"
	dsnPingFail  = "ping_fail"
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
	if m.name == dsnCloseFail {
		return errors.New("close failed")
	}
	return nil
}
func (m *mockConn) Begin() (driver.Tx, error) { return nil, nil }

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

	t.Run("successful connection with default warmup timeout", func(t *testing.T) {
		cfg := validCfg
		cfg.Warmup = true
		cfg.ConnectTimeout = 0

		db, err := newSQL(context.Background(), cfg)
		require.NoError(t, err)
		require.NotNil(t, db)
		_ = db.Close()
	})
}
