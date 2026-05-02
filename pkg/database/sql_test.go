package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	cfg := Config{
		Driver:          "postgres",
		DSN:             "postgres://user:pass@localhost:5432/db",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 10 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		Warmup:          true,
		ConnectTimeout:  2 * time.Second,
	}

	assert.Equal(t, "postgres", cfg.Driver)
	assert.Equal(t, 10, cfg.MaxOpenConns)
	assert.Equal(t, 5, cfg.MaxIdleConns)
	assert.Equal(t, 10*time.Minute, cfg.ConnMaxLifetime)
	assert.Equal(t, 5*time.Minute, cfg.ConnMaxIdleTime)
	assert.True(t, cfg.Warmup)
	assert.Equal(t, 2*time.Second, cfg.ConnectTimeout)
}

func TestNewSQL_Error(t *testing.T) {
	cfg := Config{
		Driver: "nonexistent",
		DSN:    "some-dsn",
	}

	db, err := NewSQL(cfg)
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "sql: unknown driver")
}

func TestNewSQLPair_Error(t *testing.T) {
	cfg := Config{
		Driver: "nonexistent",
		DSN:    "some-dsn",
	}

	pair, err := NewSQLPair(cfg, cfg)
	assert.Error(t, err)
	assert.Nil(t, pair)
}
