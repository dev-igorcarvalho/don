// ---
// title: DSN Builder
// description: Provides factory functions to construct Data Source Name (DSN) strings for various database drivers.
// last_updated: 2026-05-09
// type: Utility
// ---

package dsn

import (
	"fmt"
	"net/url"
)

// Postgres constructs a DSN string for the PostgreSQL driver.
func Postgres(host, port, user, password, dbName, sslMode string) string {
	if sslMode == "" {
		sslMode = "disable"
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbName, sslMode)
}

// MySQL constructs a DSN string for the MySQL driver.
func MySQL(host, port, user, password, dbName string) string {
	// Format: [username[:password]@][protocol[(address)]]/dbname
	auth := ""
	if user != "" {
		auth = user
		if password != "" {
			auth += ":" + password
		}
		auth += "@"
	}
	return fmt.Sprintf("%stcp(%s:%s)/%s?parseTime=true", auth, host, port, dbName)
}

// SQLite constructs a DSN string for the SQLite3 driver.
func SQLite(filePath string) string {
	// Simple path is usually enough for sqlite3 driver,
	// but adding cache=shared is common for performance.
	return fmt.Sprintf("%s?cache=shared&_journal_mode=WAL", filePath)
}

// PostgresURL constructs a Postgres DSN using the URL format.
func PostgresURL(host, port, user, password, dbName, sslMode string) string {
	if sslMode == "" {
		sslMode = "disable"
	}
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   fmt.Sprintf("%s:%s", host, port),
		Path:   dbName,
	}
	q := u.Query()
	q.Set("sslmode", sslMode)
	u.RawQuery = q.Encode()
	return u.String()
}
