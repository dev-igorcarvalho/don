package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSQLPair(t *testing.T) {
	t.Run("successful pair creation", func(t *testing.T) {
		cfg := validCfg
		pair, err := NewClient(context.Background(), cfg, cfg)
		require.NoError(t, err)
		require.NotNil(t, pair)
		assert.NotNil(t, pair.Writer())
		assert.NotNil(t, pair.Reader())
		_ = pair.Close()
	})

	t.Run("failed writer creation", func(t *testing.T) {
		writerCfg := Config{}
		readerCfg := validCfg
		pair, err := NewClient(context.Background(), writerCfg, readerCfg)
		assert.Error(t, err)
		assert.Nil(t, pair)
		assert.Contains(t, err.Error(), "failed to initialize writer")
	})

	t.Run("failed reader creation", func(t *testing.T) {
		writerCfg := validCfg
		readerCfg := Config{}
		pair, err := NewClient(context.Background(), writerCfg, readerCfg)
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

		pair, err := NewClient(context.Background(), writerCfg, readerCfg)
		assert.Error(t, err)
		assert.Nil(t, pair)
		assert.Equal(t, 1, closeCount, "writer should have been closed")
	})
}

func TestSQLPair_Close(t *testing.T) {
	t.Run("successful close", func(t *testing.T) {
		cfg := validCfg
		pair, err := NewClient(context.Background(), cfg, cfg)
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

		pair, err := NewClient(context.Background(), writerCfg, readerCfg)
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

		pair, err := NewClient(context.Background(), writerCfg, readerCfg)
		require.NoError(t, err)

		err = pair.Close()
		assert.ErrorContains(t, err, "failed to close reader")
	})

	t.Run("both fail to close", func(t *testing.T) {
		cfg := validCfg
		cfg.DSN = dsnCloseFail
		cfg.Warmup = true
		cfg.MaxIdleConnections = 1

		pair, err := NewClient(context.Background(), cfg, cfg)
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
	pair, err := NewClient(context.Background(), cfg, cfg)
	require.NoError(t, err)

	err = pair.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestSQLPair_Ping(t *testing.T) {
	t.Run("successful ping", func(t *testing.T) {
		cfg := validCfg
		pair, err := NewClient(context.Background(), cfg, cfg)
		require.NoError(t, err)
		defer pair.Close()

		err = pair.Ping(context.Background())
		assert.NoError(t, err)
	})

	t.Run("writer ping fail", func(t *testing.T) {
		writerCfg := validCfg
		writerCfg.DSN = dsnPingFail
		readerCfg := validCfg

		pair, err := NewClient(context.Background(), writerCfg, readerCfg)
		require.NoError(t, err)
		defer pair.Close()

		err = pair.Ping(context.Background())
		assert.ErrorContains(t, err, "writer ping failed")
	})

	t.Run("reader ping fail", func(t *testing.T) {
		writerCfg := validCfg
		readerCfg := validCfg
		readerCfg.DSN = dsnPingFail

		pair, err := NewClient(context.Background(), writerCfg, readerCfg)
		require.NoError(t, err)
		defer pair.Close()

		err = pair.Ping(context.Background())
		assert.ErrorContains(t, err, "reader ping failed")
	})

	t.Run("both ping fail", func(t *testing.T) {
		cfg := validCfg
		cfg.DSN = dsnPingFail

		pair, err := NewClient(context.Background(), cfg, cfg)
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
		pair, err := NewClient(context.Background(), cfg, cfg)
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
		writerCfg.DSN = dsnPingFail
		readerCfg := validCfg

		pair, err := NewClient(context.Background(), writerCfg, readerCfg)
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
		readerCfg.DSN = dsnPingFail

		pair, err := NewClient(context.Background(), writerCfg, readerCfg)
		require.NoError(t, err)
		defer pair.Close()

		status := pair.HealthCheck(context.Background())
		assert.True(t, status.WriterAlive)
		assert.False(t, status.ReaderAlive)
		assert.Contains(t, status.Message, "reader")
	})

	t.Run("both unhealthy", func(t *testing.T) {
		cfg := validCfg
		cfg.DSN = dsnPingFail

		pair, err := NewClient(context.Background(), cfg, cfg)
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
	pair, err := NewClient(context.Background(), cfg, cfg)
	require.NoError(t, err)
	defer pair.Close()

	stats := pair.Stats()
	assert.NotNil(t, stats.Writer)
	assert.NotNil(t, stats.Reader)
}
