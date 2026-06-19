package migrator

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFactories(t *testing.T) {
	t.Parallel()

	t.Run("SQLite3Factory initializes driver with real db", func(t *testing.T) {
		t.Parallel()
		db, err := sql.Open("sqlite3", ":memory:")
		require.NoError(t, err)
		defer db.Close()

		driver, err := SQLite3Factory(db)
		assert.NoError(t, err)
		assert.NotNil(t, driver)
	})

	t.Run("PostgresFactory returns error with closed db", func(t *testing.T) {
		t.Parallel()
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		db.Close()

		_, err = PostgresFactory(db)
		assert.Error(t, err)
	})

	t.Run("MySQLFactory returns error with closed db", func(t *testing.T) {
		t.Parallel()
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		db.Close()

		_, err = MySQLFactory(db)
		assert.Error(t, err)
	})
}
