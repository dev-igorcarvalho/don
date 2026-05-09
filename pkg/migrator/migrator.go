// ---
// title: Migrator
// description: Handles database migrations using golang-migrate.
// last_updated: 2026-05-09
// type: Utility
// ---

// Package migrator provides utilities for database schema migrations and data seeding.
// It leverages the golang-migrate library for versioned migrations and supports multiple database drivers.
package migrator

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// DriverFactory defines a function type that initializes a migration driver for a specific database implementation.
type DriverFactory func(db *sql.DB) (database.Driver, error)

// MigrateUp applies all pending "up" migrations located in the default migrations directory (scripts/database/migrations).
// It initializes the migration engine with the provided database connection and driver factory.
// If no new migrations are found, it returns nil instead of an error.
func MigrateUp(db *sql.DB, dbName string, factory DriverFactory, migrationsFolder string) error {
	driver, err := factory(db)
	if err != nil {
		return fmt.Errorf("migrator: failed to initialize driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsFolder),
		dbName,
		driver,
	)
	if err != nil {
		return fmt.Errorf("migrator: failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrator: failed to apply migrations: %w", err)
	}

	return nil
}
