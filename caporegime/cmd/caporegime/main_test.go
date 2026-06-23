package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupWorkflows(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		giveSetup func(t *testing.T, baseDir string) string
		wantCount int
		wantErr   string
	}{
		{
			name: "successful workflow discovery",
			giveSetup: func(t *testing.T, baseDir string) string {
				dir := filepath.Join(baseDir, "success")
				require.NoError(t, os.MkdirAll(dir, 0755))

				content := `package main
// Name: Test Workflow
// Description: A valid workflow description.
func main() {}
`
				err := os.WriteFile(filepath.Join(dir, "valid.go"), []byte(content), 0644)
				require.NoError(t, err)
				return dir
			},
			wantCount: 1,
		},
		{
			name: "no workflows found - directory empty",
			giveSetup: func(t *testing.T, baseDir string) string {
				dir := filepath.Join(baseDir, "empty")
				require.NoError(t, os.MkdirAll(dir, 0755))
				return dir
			},
			wantErr: "no workflows found in",
		},
		{
			name: "no workflows found - directory does not exist",
			giveSetup: func(t *testing.T, baseDir string) string {
				return filepath.Join(baseDir, "nonexistent")
			},
			wantErr: "no workflows found in",
		},
		{
			name: "no workflows found - only non-go files",
			giveSetup: func(t *testing.T, baseDir string) string {
				dir := filepath.Join(baseDir, "nongo")
				require.NoError(t, os.MkdirAll(dir, 0755))
				err := os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# Readme"), 0644)
				require.NoError(t, err)
				return dir
			},
			wantErr: "no workflows found in",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			baseDir := t.TempDir()
			dir := tt.giveSetup(t, baseDir)

			items, err := setupWorkflows(dir)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Nil(t, items)
			} else {
				require.NoError(t, err)
				assert.Len(t, items, tt.wantCount)
			}
		})
	}
}
