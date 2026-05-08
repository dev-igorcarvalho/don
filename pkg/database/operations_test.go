package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Operations(t *testing.T) {
	cfg := validCfg
	ctx := context.Background()

	t.Run("ExecContext", func(t *testing.T) {
		client, err := NewClient(ctx, WithWriter(cfg))
		require.NoError(t, err)
		defer client.Close()

		res, err := client.ExecContext(ctx, "INSERT INTO table VALUES (1)")
		assert.NoError(t, err)
		assert.NotNil(t, res)
		affected, _ := res.RowsAffected()
		assert.Equal(t, int64(1), affected)
	})

	t.Run("QueryContext Routing", func(t *testing.T) {
		t.Run("uses writer when no reader", func(t *testing.T) {
			client, _ := NewClient(ctx, WithWriter(cfg))
			defer client.Close()

			rows, err := client.QueryContext(ctx, "SELECT 1")
			assert.NoError(t, err)
			assert.NotNil(t, rows)
			rows.Close()
		})

		t.Run("uses reader when available", func(t *testing.T) {
			// Writer fails, Reader succeeds.
			// If it uses reader, Query should succeed.
			writerCfg := cfg
			writerCfg.DSN = dsnQueryFail
			readerCfg := cfg

			client, _ := NewClient(ctx, WithWriter(writerCfg), WithReader(readerCfg))
			defer client.Close()

			rows, err := client.QueryContext(ctx, "SELECT 1")
			assert.NoError(t, err, "should use reader which is successful")
			assert.NotNil(t, rows)
			rows.Close()
		})

		t.Run("uses writer when forced", func(t *testing.T) {
			// Reader fails, Writer succeeds.
			// If it forces writer, Query should succeed.
			writerCfg := cfg
			readerCfg := cfg
			readerCfg.DSN = dsnQueryFail

			client, _ := NewClient(ctx, WithWriter(writerCfg), WithReader(readerCfg))
			defer client.Close()

			writerCtx := WithWriterContext(ctx)
			rows, err := client.QueryContext(writerCtx, "SELECT 1")
			assert.NoError(t, err, "should use writer because it is forced")
			assert.NotNil(t, rows)
			rows.Close()
		})
	})

	t.Run("QueryRowContext Routing", func(t *testing.T) {
		t.Run("uses reader when available", func(t *testing.T) {
			writerCfg := cfg
			writerCfg.DSN = dsnQueryFail
			readerCfg := cfg

			client, _ := NewClient(ctx, WithWriter(writerCfg), WithReader(readerCfg))
			defer client.Close()

			row := client.QueryRowContext(ctx, "SELECT 1")
			assert.NotNil(t, row)
			var val int
			err := row.Scan(&val) // In mock, Scan doesn't do much but it triggers QueryRow if not deferred
			assert.NoError(t, err)
		})

		t.Run("uses writer when forced", func(t *testing.T) {
			writerCfg := cfg
			readerCfg := cfg
			readerCfg.DSN = dsnQueryFail

			client, _ := NewClient(ctx, WithWriter(writerCfg), WithReader(readerCfg))
			defer client.Close()

			writerCtx := WithWriterContext(ctx)
			row := client.QueryRowContext(writerCtx, "SELECT 1")
			assert.NotNil(t, row)
		})
	})

	t.Run("BeginTx", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			client, _ := NewClient(ctx, WithWriter(cfg))
			defer client.Close()

			tx, err := client.BeginTx(ctx, nil)
			assert.NoError(t, err)
			assert.NotNil(t, tx)
			_ = tx.Rollback()
		})

		t.Run("failure", func(t *testing.T) {
			writerCfg := cfg
			writerCfg.DSN = dsnBeginFail
			client, _ := NewClient(ctx, WithWriter(writerCfg))
			defer client.Close()

			tx, err := client.BeginTx(ctx, nil)
			assert.Error(t, err)
			assert.Nil(t, tx)
		})
	})

	t.Run("InTransaction", func(t *testing.T) {
		client, _ := NewClient(ctx, WithWriter(cfg))
		defer client.Close()

		t.Run("success", func(t *testing.T) {
			err := client.InTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
				return nil
			})
			assert.NoError(t, err)
		})

		t.Run("callback error triggers rollback", func(t *testing.T) {
			expectedErr := errors.New("something went wrong")
			err := client.InTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
				return expectedErr
			})
			assert.ErrorIs(t, err, expectedErr)
		})

		t.Run("panic triggers rollback", func(t *testing.T) {
			err := client.InTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
				panic("kaboom")
			})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "panic in transaction: kaboom")
		})

		t.Run("begin failure", func(t *testing.T) {
			writerCfg := cfg
			writerCfg.DSN = dsnBeginFail
			clientFail, _ := NewClient(ctx, WithWriter(writerCfg))
			defer clientFail.Close()

			err := clientFail.InTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
				return nil
			})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "begin failed")
		})

		t.Run("commit failure", func(t *testing.T) {
			writerCfg := cfg
			writerCfg.DSN = dsnCommitFail
			clientFail, _ := NewClient(ctx, WithWriter(writerCfg))
			defer clientFail.Close()

			err := clientFail.InTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
				return nil
			})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to commit transaction")
		})
	})

	t.Run("ensureContextWithTimeout", func(t *testing.T) {
		client, _ := NewClient(ctx, WithWriter(cfg))
		defer client.Close()

		t.Run("no timeout when duration <= 0", func(t *testing.T) {
			newCtx, cancel := client.ensureContextWithTimeout(ctx, 0)
			assert.Equal(t, ctx, newCtx)
			assert.Nil(t, cancel)
		})

		t.Run("no timeout when deadline already set", func(t *testing.T) {
			deadlineCtx, cancel := context.WithDeadline(ctx, time.Now().Add(1*time.Hour))
			defer cancel()

			newCtx, newCancel := client.ensureContextWithTimeout(deadlineCtx, 1*time.Minute)
			assert.Equal(t, deadlineCtx, newCtx)
			assert.Nil(t, newCancel)
		})

		t.Run("applies timeout when none exists", func(t *testing.T) {
			newCtx, cancel := client.ensureContextWithTimeout(ctx, 1*time.Minute)
			assert.NotNil(t, cancel)
			defer cancel()

			deadline, ok := newCtx.Deadline()
			assert.True(t, ok)
			assert.WithinDuration(t, time.Now().Add(1*time.Minute), deadline, 1*time.Second)
		})
	})
}
