package primitives

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dev-igorcarvalho/don/caporegime/pkg/utils"
)

const (
	// artifactFileExtension is the extension applied to every persisted artifact file.
	artifactFileExtension = ".md"
	// artifactFilePermissions is the file mode used when writing artifact files.
	artifactFilePermissions = 0o644
)

// artifactFilePath builds the destination path for an artifact within dir,
// sanitizing artifactName so it is safe to use as a file name.
func artifactFilePath(dir, artifactName string) string {
	filename := utils.SanitizeName(artifactName) + artifactFileExtension
	return filepath.Join(dir, filename)
}

// PersistArtifactToFile writes the artifact content to a file in the artifact directory specified in the context.
// It returns the absolute path of the written file and any file system error encountered.
// If the artifact directory is not configured in the context, it returns an empty path and a nil error.
func PersistArtifactToFile(ctx context.Context, artifactName, content string) (string, error) {
	dir, ok := ArtifactDir(ctx)
	if !ok || dir == "" {
		return "", nil
	}
	path := artifactFilePath(dir, artifactName)
	Logger(ctx).Info("persisting agent artifact", "name", artifactName, "path", path)
	if err := os.WriteFile(path, []byte(content), artifactFilePermissions); err != nil {
		return "", fmt.Errorf("persist agent artifact: %w", err)
	}
	return path, nil
}
