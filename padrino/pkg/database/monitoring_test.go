package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLPair_Ping(t *testing.T) {
	t.Run("successful ping", func(t *testing.T) {
		cfg := validCfg
		pair, err := NewClient(context.Background(), WithWriter(cfg), WithReader(cfg))
		require.NoError(t, err)
		defer pair.Close()

		err = pair.Ping(context.Background())
		assert.NoError(t, err)
	})

	t.Run("writer ping fail", func(t *testing.T) {
		writerCfg := validCfg
		writerCfg.DSN = dsnPingFail
		readerCfg := validCfg

		pair, err := NewClient(context.Background(), WithWriter(writerCfg), WithReader(readerCfg))
		require.NoError(t, err)
		defer pair.Close()

		err = pair.Ping(context.Background())
		assert.ErrorContains(t, err, "writer ping failed")
	})

	t.Run("reader ping fail", func(t *testing.T) {
		writerCfg := validCfg
		readerCfg := validCfg
		readerCfg.DSN = dsnPingFail

		pair, err := NewClient(context.Background(), WithWriter(writerCfg), WithReader(readerCfg))
		require.NoError(t, err)
		defer pair.Close()

		err = pair.Ping(context.Background())
		assert.ErrorContains(t, err, "reader ping failed")
	})

	t.Run("both ping fail", func(t *testing.T) {
		cfg := validCfg
		cfg.DSN = dsnPingFail

		pair, err := NewClient(context.Background(), WithWriter(cfg), WithReader(cfg))
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
		pair, err := NewClient(context.Background(), WithWriter(cfg), WithReader(cfg))
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

		pair, err := NewClient(context.Background(), WithWriter(writerCfg), WithReader(readerCfg))
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

		pair, err := NewClient(context.Background(), WithWriter(writerCfg), WithReader(readerCfg))
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

		pair, err := NewClient(context.Background(), WithWriter(cfg), WithReader(cfg))
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
	pair, err := NewClient(context.Background(), WithWriter(cfg), WithReader(cfg))
	require.NoError(t, err)
	defer pair.Close()

	stats := pair.Stats()
	assert.NotNil(t, stats.Writer)
	assert.NotNil(t, stats.Reader)
}
