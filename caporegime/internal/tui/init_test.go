package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitializeWorkspace(t *testing.T) {
	t.Parallel()

	t.Run("creates directory and writes default workflow", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		targetDir := filepath.Join(tempDir, "workflows")

		err := InitializeWorkspace(targetDir)
		require.NoError(t, err)

		// Assert directory exists
		info, err := os.Stat(targetDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())

		// Assert hello.go exists
		helloPath := filepath.Join(targetDir, "hello.go")
		_, err = os.Stat(helloPath)
		assert.NoError(t, err)

		// Assert content is what we expect
		content, err := os.ReadFile(helloPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "Hello Workflow")
	})

	t.Run("does not overwrite existing go workflows", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		targetDir := filepath.Join(tempDir, "workflows")

		err := os.MkdirAll(targetDir, 0755)
		require.NoError(t, err)

		existingPath := filepath.Join(targetDir, "custom.go")
		err = os.WriteFile(existingPath, []byte("// custom workflow"), 0644)
		require.NoError(t, err)

		err = InitializeWorkspace(targetDir)
		require.NoError(t, err)

		// Assert hello.go does not exist because custom.go was already there
		helloPath := filepath.Join(targetDir, "hello.go")
		_, err = os.Stat(helloPath)
		assert.True(t, os.IsNotExist(err))
	})
}
