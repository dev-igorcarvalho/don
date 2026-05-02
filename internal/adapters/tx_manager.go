// ---
// title: SQL Transaction Manager
// description: Implements atomicity for database operations using the Writer connection from SQLPair.
// last_updated: 2026-05-02
// type: Adapter
// ---

package adapters

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dev-igorcarvalho/don/pkg/database"
)

// TransactionManager defines the interface for managing database transactions
type TransactionManager interface {
	Atomic(ctx context.Context, fn func(ctx context.Context) error) error
}

type sqlTransactionManager struct {
	sqlPair *database.SQLPair
}

// NewTransactionManager creates a new SQL transaction manager
func NewTransactionManager(sqlPair *database.SQLPair) TransactionManager {
	return &sqlTransactionManager{sqlPair: sqlPair}
}

// Atomic executes a function within a database transaction
func (m *sqlTransactionManager) Atomic(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.sqlPair.Writer.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Use a custom context that carries the transaction
	txCtx := context.WithValue(ctx, txKey{}, tx)

	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to rollback transaction after error (%v): %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

type txKey struct{}

// GetTx retrieves the transaction from the context if it exists
func GetTx(ctx context.Context) *sql.Tx {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}
	return nil
}
