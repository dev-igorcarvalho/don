package migrator

import (
	"context"
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failReadFileFS struct {
	fstest.MapFS
}

func (f failReadFileFS) ReadFile(name string) ([]byte, error) {
	return nil, errors.New("read error")
}

func TestRunSeeds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		seedsFS   fs.FS
		seedsPath string
		mockSQL   func(mock sqlmock.Sqlmock)
		ctx       func() (context.Context, context.CancelFunc)
		wantErr   string
	}{
		{
			name: "success: executes multiple files in order",
			seedsFS: fstest.MapFS{
				"seeds/0002_second.sql": &fstest.MapFile{Data: []byte("INSERT INTO t2 VALUES (1);")},
				"seeds/0001_first.sql":  &fstest.MapFile{Data: []byte("INSERT INTO t1 VALUES (1);")},
				"seeds/not_sql.txt":     &fstest.MapFile{Data: []byte("SHOULD BE IGNORED")},
			},
			seedsPath: "seeds",
			mockSQL: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO t1 VALUES (1);").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO t2 VALUES (1);").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: "",
		},
		{
			name:      "error: failed to read directory",
			seedsFS:   fstest.MapFS{},
			seedsPath: "non_existent",
			mockSQL:   func(mock sqlmock.Sqlmock) {},
			wantErr:   "seeder: failed to read directory",
		},
		{
			name: "error: failed to read file",
			seedsFS: failReadFileFS{
				MapFS: fstest.MapFS{
					"seeds/0001_first.sql": &fstest.MapFile{Data: []byte("SELECT 1")},
				},
			},
			seedsPath: "seeds",
			mockSQL:   func(mock sqlmock.Sqlmock) {},
			wantErr:   "seeder: failed to read 0001_first.sql: read error",
		},
		{
			name: "error: execution failed",
			seedsFS: fstest.MapFS{
				"seeds/0001_first.sql": &fstest.MapFile{Data: []byte("INVALID SQL")},
			},
			seedsPath: "seeds",
			mockSQL: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INVALID SQL").WillReturnError(errors.New("db error"))
			},
			wantErr: "seeder: execution failed for 0001_first.sql: db error",
		},
		{
			name: "error: context cancelled before execution",
			seedsFS: fstest.MapFS{
				"seeds/0001_first.sql": &fstest.MapFile{Data: []byte("SELECT 1")},
			},
			seedsPath: "seeds",
			mockSQL:   func(mock sqlmock.Sqlmock) {},
			ctx: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx, cancel
			},
			wantErr: "seeder: context cancelled",
		},
		{
			name: "success: no sql files found",
			seedsFS: fstest.MapFS{
				"seeds/empty": &fstest.MapFile{Data: []byte("")},
			},
			seedsPath: "seeds",
			mockSQL:   func(mock sqlmock.Sqlmock) {},
			wantErr:   "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer db.Close()

			if tt.mockSQL != nil {
				tt.mockSQL(mock)
			}

			ctx := context.Background()
			if tt.ctx != nil {
				var cancel context.CancelFunc
				ctx, cancel = tt.ctx()
				defer cancel()
			}

			err = RunSeeds(ctx, db, tt.seedsFS, tt.seedsPath)

			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
