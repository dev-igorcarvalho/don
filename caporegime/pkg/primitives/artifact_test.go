package primitives

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArtifactFilePath(t *testing.T) {
	tests := []struct {
		name         string
		dir          string
		artifactName string
		want         string
	}{
		{
			name:         "simple name",
			dir:          "/tmp/artifacts",
			artifactName: "report",
			want:         filepath.Join("/tmp/artifacts", "report.md"),
		},
		{
			name:         "name with spaces and case",
			dir:          "/tmp/artifacts",
			artifactName: "My Artifact",
			want:         filepath.Join("/tmp/artifacts", "my_artifact.md"),
		},
		{
			name:         "name with slashes and hyphens",
			dir:          "/tmp/artifacts",
			artifactName: "sub/dir-name",
			want:         filepath.Join("/tmp/artifacts", "sub_dir_name.md"),
		},
		{
			name:         "empty artifact name",
			dir:          "/tmp/artifacts",
			artifactName: "",
			want:         filepath.Join("/tmp/artifacts", ".md"),
		},
		{
			name:         "empty dir",
			dir:          "",
			artifactName: "report",
			want:         "report.md",
		},
		{
			name:         "relative dir",
			dir:          "relative/dir",
			artifactName: "report",
			want:         filepath.Join("relative/dir", "report.md"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := artifactFilePath(tt.dir, tt.artifactName); got != tt.want {
				t.Errorf("artifactFilePath(%q, %q) = %q, want %q", tt.dir, tt.artifactName, got, tt.want)
			}
		})
	}
}

func TestPersistArtifactToFile(t *testing.T) {
	t.Run("nil context returns empty result", func(t *testing.T) {
		path, err := PersistArtifactToFile(context.TODO(), "test", "content")
		if err != nil {
			t.Fatalf("PersistArtifactToFile() error = %v, want nil", err)
		}
		if path != "" {
			t.Errorf("PersistArtifactToFile() path = %q, want empty", path)
		}
	})

	t.Run("missing artifact dir in context returns empty result", func(t *testing.T) {
		path, err := PersistArtifactToFile(context.Background(), "test", "content")
		if err != nil {
			t.Fatalf("PersistArtifactToFile() error = %v, want nil", err)
		}
		if path != "" {
			t.Errorf("PersistArtifactToFile() path = %q, want empty", path)
		}
	})

	t.Run("empty artifact dir in context returns empty result", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), artifactDirKey{}, "")
		path, err := PersistArtifactToFile(ctx, "test", "content")
		if err != nil {
			t.Fatalf("PersistArtifactToFile() error = %v, want nil", err)
		}
		if path != "" {
			t.Errorf("PersistArtifactToFile() path = %q, want empty", path)
		}
	})

	t.Run("successful persist writes content and returns path", func(t *testing.T) {
		tmpDir := t.TempDir()
		ctx := context.WithValue(context.Background(), artifactDirKey{}, tmpDir)
		path, err := PersistArtifactToFile(ctx, "my artifact", "the file content")
		if err != nil {
			t.Fatalf("PersistArtifactToFile() error = %v, want nil", err)
		}
		wantPath := filepath.Join(tmpDir, "my_artifact.md")
		if path != wantPath {
			t.Errorf("PersistArtifactToFile() path = %q, want %q", path, wantPath)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read persisted file: %v", err)
		}
		if string(data) != "the file content" {
			t.Errorf("file content = %q, want %q", string(data), "the file content")
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("failed to stat persisted file: %v", err)
		}
		if perm := info.Mode().Perm(); perm != artifactFilePermissions {
			t.Errorf("file permissions = %v, want %v", perm, os.FileMode(artifactFilePermissions))
		}
	})

	t.Run("empty content is written as empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		ctx := context.WithValue(context.Background(), artifactDirKey{}, tmpDir)
		path, err := PersistArtifactToFile(ctx, "empty", "")
		if err != nil {
			t.Fatalf("PersistArtifactToFile() error = %v, want nil", err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read persisted file: %v", err)
		}
		if len(data) != 0 {
			t.Errorf("file content = %q, want empty", string(data))
		}
	})

	t.Run("overwrites existing file at same path", func(t *testing.T) {
		tmpDir := t.TempDir()
		ctx := context.WithValue(context.Background(), artifactDirKey{}, tmpDir)
		if _, err := PersistArtifactToFile(ctx, "dup", "first"); err != nil {
			t.Fatalf("first PersistArtifactToFile() error = %v, want nil", err)
		}
		path, err := PersistArtifactToFile(ctx, "dup", "second")
		if err != nil {
			t.Fatalf("second PersistArtifactToFile() error = %v, want nil", err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read persisted file: %v", err)
		}
		if string(data) != "second" {
			t.Errorf("file content = %q, want %q", string(data), "second")
		}
	})

	t.Run("nonexistent directory returns wrapped error", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), artifactDirKey{}, filepath.Join(t.TempDir(), "does-not-exist"))
		path, err := PersistArtifactToFile(ctx, "test", "content")
		if err == nil {
			t.Fatal("PersistArtifactToFile() error = nil, want error")
		}
		if !strings.Contains(err.Error(), "persist agent artifact:") {
			t.Errorf("PersistArtifactToFile() error = %v, want prefix %q", err, "persist agent artifact:")
		}
		if path != "" {
			t.Errorf("PersistArtifactToFile() path = %q, want empty on error", path)
		}
	})

	t.Run("unwritable directory returns wrapped error", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("skipping permission test when running as root")
		}
		roDir := t.TempDir()
		if err := os.Chmod(roDir, 0o555); err != nil {
			t.Fatalf("failed to chmod dir: %v", err)
		}
		defer func() {
			if err := os.Chmod(roDir, 0o755); err != nil {
				t.Fatalf("failed to restore dir permissions: %v", err)
			}
		}() // restore so TempDir cleanup can remove the directory

		ctx := context.WithValue(context.Background(), artifactDirKey{}, roDir)
		path, err := PersistArtifactToFile(ctx, "test", "content")
		if err == nil {
			t.Fatal("PersistArtifactToFile() error = nil, want error")
		}
		if !strings.Contains(err.Error(), "persist agent artifact:") {
			t.Errorf("PersistArtifactToFile() error = %v, want prefix %q", err, "persist agent artifact:")
		}
		if path != "" {
			t.Errorf("PersistArtifactToFile() path = %q, want empty on error", path)
		}
	})
}
