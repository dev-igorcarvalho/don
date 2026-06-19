// ---
// title: Migration Driver Factories
// description: Provides factory functions for initializing migration drivers for different databases.
// last_updated: 2026-05-09
// type: Utility
// ---

package migrator

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
)

// PostgresFactory initializes a migration driver for PostgreSQL.
// It uses the standard postgres driver instance with default configuration.
func PostgresFactory(db *sql.DB) (database.Driver, error) {
	return postgres.WithInstance(db, &postgres.Config{})
}

// MySQLFactory initializes a migration driver for MySQL.
// It uses the standard mysql driver instance with default configuration.
func MySQLFactory(db *sql.DB) (database.Driver, error) {
	return mysql.WithInstance(db, &mysql.Config{})
}

// SQLite3Factory initializes a migration driver for SQLite3.
// It uses the standard sqlite3 driver instance with default configuration.
func SQLite3Factory(db *sql.DB) (database.Driver, error) {
	return sqlite3.WithInstance(db, &sqlite3.Config{})
}
