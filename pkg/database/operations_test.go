package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Operations(t *testing.T) {
	cfg := validCfg
	client, err := NewClient(context.Background(), WithWriter(cfg), WithReader(cfg))
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()

	t.Run("ExecContext", func(t *testing.T) {
		res, err := client.ExecContext(ctx, "INSERT INTO table VALUES (1)")
		assert.NoError(t, err)
		assert.NotNil(t, res)
		affected, _ := res.RowsAffected()
		assert.Equal(t, int64(1), affected)
	})

	t.Run("QueryContext - uses reader", func(t *testing.T) {
		rows, err := client.QueryContext(ctx, "SELECT 1")
		assert.NoError(t, err)
		assert.NotNil(t, rows)
		rows.Close()
	})

	t.Run("QueryRowContext - uses reader", func(t *testing.T) {
		row := client.QueryRowContext(ctx, "SELECT 1")
		assert.NotNil(t, row)
	})

	t.Run("QueryContext - fallback to writer", func(t *testing.T) {
		clientWriterOnly, err := NewClient(ctx, WithWriter(cfg))
		require.NoError(t, err)
		defer clientWriterOnly.Close()

		rows, err := clientWriterOnly.QueryContext(ctx, "SELECT 1")
		assert.NoError(t, err)
		assert.NotNil(t, rows)
		rows.Close()
	})

	t.Run("QueryRowContext - fallback to writer", func(t *testing.T) {
		clientWriterOnly, err := NewClient(ctx, WithWriter(cfg))
		require.NoError(t, err)
		defer clientWriterOnly.Close()

		row := clientWriterOnly.QueryRowContext(ctx, "SELECT 1")
		assert.NotNil(t, row)
	})

	t.Run("ExecContext failure", func(t *testing.T) {
		failCfg := cfg
		failCfg.DSN = dsnExecFail
		clientFail, err := NewClient(ctx, WithWriter(failCfg))
		require.NoError(t, err)
		defer clientFail.Close()

		_, err = clientFail.ExecContext(ctx, "INSERT INTO table VALUES (1)")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exec failed")
	})

	t.Run("QueryContext failure", func(t *testing.T) {
		failCfg := cfg
		failCfg.DSN = dsnQueryFail
		clientFail, err := NewClient(ctx, WithWriter(failCfg))
		require.NoError(t, err)
		defer clientFail.Close()

		_, err = clientFail.QueryContext(ctx, "SELECT 1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "query failed")
	})
}
