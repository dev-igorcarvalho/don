package migrator

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDriver implements database.Driver for testing purposes.
type mockDriver struct {
	errOnRun        error
	errOnSetVersion error
	version         int
	dirty           bool
}

func (m *mockDriver) Open(url string) (database.Driver, error) { return m, nil }
func (m *mockDriver) Close() error                             { return nil }
func (m *mockDriver) Lock() error                              { return nil }
func (m *mockDriver) Unlock() error                            { return nil }
func (m *mockDriver) Run(migration io.Reader) error            { return m.errOnRun }
func (m *mockDriver) SetVersion(version int, dirty bool) error {
	m.version = version
	m.dirty = dirty
	return m.errOnSetVersion
}
func (m *mockDriver) Version() (version int, dirty bool, err error) { return m.version, m.dirty, nil }
func (m *mockDriver) Drop() error                                   { return nil }

func TestMigrateUp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		factory DriverFactory
		wantErr string
	}{
		{
			name: "success: migrations applied",
			factory: func(db *sql.DB) (database.Driver, error) {
				return &mockDriver{version: -1}, nil
			},
			wantErr: "",
		},
		{
			name: "success: no changes needed",
			factory: func(db *sql.DB) (database.Driver, error) {
				return &mockDriver{version: 1}, nil
			},
			wantErr: "",
		},
		{
			name: "error: factory failure",
			factory: func(db *sql.DB) (database.Driver, error) {
				return nil, errors.New("factory error")
			},
			wantErr: "migrator: failed to initialize driver: factory error",
		},
		{
			name: "error: migrate instance failure (invalid path)",
			factory: func(db *sql.DB) (database.Driver, error) {
				return &mockDriver{version: -1}, nil
			},
		},
		{
			name: "error: apply failure",
			factory: func(db *sql.DB) (database.Driver, error) {
				return &mockDriver{version: -1, errOnRun: errors.New("db error")}, nil
			},
			wantErr: "migrator: failed to apply migrations: db error",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a temporary directory for dummy migrations for each subtest
			tmpDir, err := os.MkdirTemp("", "migrations-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			// Create a dummy migration file with standard naming
			err = os.WriteFile(fmt.Sprintf("%s/000001_init.up.sql", tmpDir), []byte("CREATE TABLE test (id INT);"), 0644)
			require.NoError(t, err)

			db, _, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			folder := tmpDir
			if tt.name == "error: migrate instance failure (invalid path)" {
				folder = "/non/existent/path"
			}

			err = MigrateUp(db, "testdb", tt.factory, folder)

			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else if tt.name == "error: migrate instance failure (invalid path)" {
				assert.ErrorContains(t, err, "migrator: failed to create migrate instance")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
