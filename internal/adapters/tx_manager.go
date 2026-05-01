package adapters

import (
	"context"
	"database/sql"
	"fmt"
)

// TransactionManager defines the interface for managing database transactions
type TransactionManager interface {
	Atomic(ctx context.Context, fn func(ctx context.Context) error) error
}

type sqlTransactionManager struct {
	db *sql.DB
}

// NewTransactionManager creates a new SQL transaction manager
func NewTransactionManager(db *sql.DB) TransactionManager {
	return &sqlTransactionManager{db: db}
}

// Atomic executes a function within a database transaction
func (m *sqlTransactionManager) Atomic(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
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
