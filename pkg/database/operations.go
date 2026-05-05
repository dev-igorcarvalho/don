// ---
// title: Database Operations
// description: Provides SQL operations (Query, Exec, Transactions) with automatic read/write splitting.
// last_updated: 2026-05-05
// type: Implementation
// ---

package database

import (
	"context"
	"database/sql"
)

// ExecContext executes a query without returning any rows.
// It always uses the writer connection pool.
func (p *Client) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return p.writer.ExecContext(ctx, query, args...)
}

// QueryContext executes a query that returns rows, typically a SELECT.
// It uses the reader connection pool if available, otherwise it falls back to the writer.
func (p *Client) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if p.reader != nil {
		return p.reader.QueryContext(ctx, query, args...)
	}
	return p.writer.QueryContext(ctx, query, args...)
}

// QueryRowContext executes a query that is expected to return at most one row.
// It uses the reader connection pool if available, otherwise it falls back to the writer.
func (p *Client) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	if p.reader != nil {
		return p.reader.QueryRowContext(ctx, query, args...)
	}
	return p.writer.QueryRowContext(ctx, query, args...)
}
