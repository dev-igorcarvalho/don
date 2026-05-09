// ---
// title: Database Operations
// description: Provides SQL operations (Query, Exec, Transactions) with automatic read/write splitting, default timeouts, and forced master reads.
// last_updated: 2026-05-09
// type: Implementation
// ---

package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// TxFunc is a function that can be executed within a transaction.
type TxFunc func(ctx context.Context, tx *sql.Tx) error

// ExecContext executes a query without returning any rows.
// It always uses the writer connection pool.
func (p *Client) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	ctx, cancel := p.ensureContextWithTimeout(ctx, p.writerTimeout)
	if cancel != nil {
		defer cancel()
	}

	return p.writer.ExecContext(ctx, query, args...)
}

// QueryContext executes a query that returns rows, typically a SELECT.
// It uses the reader connection pool if available, otherwise it falls back to the writer.
// If the context has the writer forced signal, it uses the writer pool.
func (p *Client) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	useWriter := p.reader == nil || isWriterForced(ctx)

	var timeout time.Duration
	if useWriter {
		timeout = p.writerTimeout
	} else {
		timeout = p.readerTimeout
	}

	// Note: We don't defer cancel() here because sql.Rows needs the context to remain active
	// while the caller iterates over the results. The context will be cleaned up when the
	// timeout expires or the parent context is canceled.
	ctx, _ = p.ensureContextWithTimeout(ctx, timeout)

	if useWriter {
		return p.writer.QueryContext(ctx, query, args...)
	}
	return p.reader.QueryContext(ctx, query, args...)
}

// QueryRowContext executes a query that is expected to return at most one row.
// It uses the reader connection pool if available, otherwise it falls back to the writer.
// If the context has the writer forced signal, it uses the writer pool.
func (p *Client) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	useWriter := p.reader == nil || isWriterForced(ctx)

	var timeout time.Duration
	if useWriter {
		timeout = p.writerTimeout
	} else {
		timeout = p.readerTimeout
	}

	// Note: We don't defer cancel() here because sql.Row's Scan needs the context.
	ctx, _ = p.ensureContextWithTimeout(ctx, timeout)

	if useWriter {
		return p.writer.QueryRowContext(ctx, query, args...)
	}
	return p.reader.QueryRowContext(ctx, query, args...)
}

// BeginTx starts a new transaction on the writer connection.
func (p *Client) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	// Note: We don't defer cancel() here because the transaction needs the context
	// to remain active until it is committed or rolled back.
	ctx, _ = p.ensureContextWithTimeout(ctx, p.writerTimeout)

	return p.writer.BeginTx(ctx, opts)
}

// InTransaction executes the provided function within a transaction.
// Lifecycle Management:
//
//  1. Starts a new transaction on the writer connection using BeginTx.
//
//  2. Executes the callback function 'fn'.
//     Example usage:
//     err := db.InTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
//     // 'tx' is the active transaction. Use it for all operations.
//     res, err := tx.ExecContext(ctx, "UPDATE users SET active = ? WHERE id = ?", true, 1)
//     if err != nil {
//     return fmt.Errorf("failed to update user: %w", err) // triggers Rollback
//     }
//
//     if _, err := tx.ExecContext(ctx, "INSERT INTO logs..."); err != nil {
//     return fmt.Errorf("failed to log action: %w", err) // triggers Rollback
//     }
//
//     return nil // triggers Commit
//     })
//
// 3. Atomicity:
//   - If 'fn' returns an error, the transaction is automatically rolled back.
//   - If 'fn' panics, the panic is captured, the transaction is rolled back, and the panic is returned as an error.
//   - If 'fn' succeeds (returns nil), the transaction is committed.
//   - If the commit itself fails, that error is returned.
func (p *Client) InTransaction(ctx context.Context, opts *sql.TxOptions, fn TxFunc) (err error) {
	// todo revisar essa implementação, ainda nao achei 100%
	ctx, cancel := p.ensureContextWithTimeout(ctx, p.writerTimeout)
	if cancel != nil {
		defer cancel()
	}

	tx, err := p.writer.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			err = fmt.Errorf("panic in transaction: %v", r)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			if cmErr := tx.Commit(); cmErr != nil {
				err = fmt.Errorf("failed to commit transaction: %w", cmErr)
			}
		}
	}()

	err = fn(ctx, tx)
	return err
}

// ensureContextWithTimeout ensures the context has a timeout if one is provided and no deadline is already set.
func (p *Client) ensureContextWithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return ctx, nil
	}

	if _, ok := ctx.Deadline(); ok {
		return ctx, nil
	}

	return context.WithTimeout(ctx, timeout)
}
