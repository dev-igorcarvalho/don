// ---
// title: Base Repository Implementation
// description: Provides a foundational structure for database repositories with support for read/write splitting.
// last_updated: 2026-05-02
// type: Adapter
// ---

package adapters

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/dev-igorcarvalho/don/pkg/database"
	"github.com/dev-igorcarvalho/don/pkg/logger"
)

// BaseRepository provides common functionality for database repositories
type BaseRepository struct {
	sqlPair      *database.SQLPair
	queryTimeout time.Duration
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(sqlPair *database.SQLPair, queryTimeout time.Duration) *BaseRepository {
	return &BaseRepository{
		sqlPair:      sqlPair,
		queryTimeout: queryTimeout,
	}
}

// ExecContext executes a query without returning any rows
func (r *BaseRepository) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	var result sql.Result
	var err error

	tx := GetTx(ctx)
	if tx != nil {
		result, err = tx.ExecContext(ctx, query, args...)
	} else {
		result, err = r.sqlPair.Writer.ExecContext(ctx, query, args...)
	}

	r.logQuery(ctx, query, time.Since(start), err)

	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return result, nil
}

// QueryContext executes a query that returns rows
func (r *BaseRepository) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	var rows *sql.Rows
	var err error

	tx := GetTx(ctx)
	if tx != nil {
		rows, err = tx.QueryContext(ctx, query, args...)
	} else {
		rows, err = r.sqlPair.Reader.QueryContext(ctx, query, args...)
	}

	r.logQuery(ctx, query, time.Since(start), err)

	if err != nil {
		return nil, fmt.Errorf("failed to query rows: %w", err)
	}

	return rows, nil
}

// QueryRowContext executes a query that is expected to return at most one row
func (r *BaseRepository) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	start := time.Now()
	// Note: We don't have an easy way to cancel QueryRowContext after it returns *sql.Row
	// until Scan is called, but we can still set a timeout for the call itself.
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	var row *sql.Row
	tx := GetTx(ctx)
	if tx != nil {
		row = tx.QueryRowContext(ctx, query, args...)
	} else {
		row = r.sqlPair.Reader.QueryRowContext(ctx, query, args...)
	}

	// We can't log the error here because it's only available after Scan()
	// But we can log that we initiated the query
	r.logQuery(ctx, query, time.Since(start), nil)

	return row
}

func (r *BaseRepository) logQuery(ctx context.Context, query string, duration time.Duration, err error) {
	attrs := []slog.Attr{
		slog.String("query", query),
		slog.Duration("duration", duration),
	}

	switch {
	case err != nil:
		attrs = append(attrs, slog.Any("error", err))
		logger.Error(ctx, "database query failed", attrs...)
	case duration > 200*time.Millisecond:
		logger.Warn(ctx, "slow database query detected", attrs...)
	default:
		logger.Debug(ctx, "database query executed", attrs...)
	}
}

// CheckAffectedRows validates that the expected number of rows were affected
func (r *BaseRepository) CheckAffectedRows(result sql.Result, expected int64) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if affected != expected {
		return fmt.Errorf("unexpected number of rows affected: expected %d, got %d", expected, affected)
	}

	return nil
}
