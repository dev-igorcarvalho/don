// ---
// title: Seeder
// description: Executes SQL seed scripts from a filesystem in alphabetical order.
// last_updated: 2026-05-09
// type: Utility
// ---

package migrator

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
)

// RunSeeds discovers and executes SQL seed scripts from the provided filesystem (fs.FS) at the specified path.
// It executes scripts in alphabetical order by filename. Only files with the .sql extension are processed.
// Each script is executed as a single batch using ExecContext.
func RunSeeds(ctx context.Context, db *sql.DB, seedsFS fs.FS, seedsPath string) error {
	entries, err := fs.ReadDir(seedsFS, seedsPath)
	if err != nil {
		return fmt.Errorf("seeder: failed to read directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("seeder: context cancelled: %w", err)
		}

		content, err := fs.ReadFile(seedsFS, filepath.Join(seedsPath, file))
		if err != nil {
			return fmt.Errorf("seeder: failed to read %s: %w", file, err)
		}

		if _, err := db.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("seeder: execution failed for %s: %w", file, err)
		}
	}
	return nil
}
