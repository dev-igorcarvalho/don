package dsn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostgres(t *testing.T) {
	t.Parallel()

	expected := "host=localhost port=5432 user=user password=pass dbname=db sslmode=disable"
	result := Postgres("localhost", "5432", "user", "pass", "db", "")
	assert.Equal(t, expected, result)

	expectedWithSSL := "host=localhost port=5432 user=user password=pass dbname=db sslmode=require"
	resultWithSSL := Postgres("localhost", "5432", "user", "pass", "db", "require")
	assert.Equal(t, expectedWithSSL, resultWithSSL)
}

func TestMySQL(t *testing.T) {
	t.Parallel()

	expected := "user:pass@tcp(localhost:3306)/db?parseTime=true"
	result := MySQL("localhost", "3306", "user", "pass", "db")
	assert.Equal(t, expected, result)

	expectedNoPass := "user@tcp(localhost:3306)/db?parseTime=true"
	resultNoPass := MySQL("localhost", "3306", "user", "", "db")
	assert.Equal(t, expectedNoPass, resultNoPass)
}

func TestSQLite(t *testing.T) {
	t.Parallel()

	expected := "/tmp/test.db?cache=shared&_journal_mode=WAL"
	result := SQLite("/tmp/test.db")
	assert.Equal(t, expected, result)
}

func TestPostgresURL(t *testing.T) {
	t.Parallel()

	// url.URL will escape special characters in password
	expected := "postgres://user:pass@localhost:5432/db?sslmode=disable"
	result := PostgresURL("localhost", "5432", "user", "pass", "db", "")
	assert.Equal(t, expected, result)
}
