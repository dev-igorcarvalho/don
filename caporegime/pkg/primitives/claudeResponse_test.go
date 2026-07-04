package primitives

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClaudeResponse_Failure(t *testing.T) {
	tests := []struct {
		name    string
		r       *ClaudeResponse
		wantErr string // empty means want nil error
	}{
		{
			name:    "no error returns nil",
			r:       &ClaudeResponse{IsError: false, RawResult: "all good"},
			wantErr: "",
		},
		{
			name:    "error true returns error wrapping raw result",
			r:       &ClaudeResponse{IsError: true, RawResult: "something broke"},
			wantErr: "something broke",
		},
		{
			name:    "error true with empty raw result returns non-nil error with empty message",
			r:       &ClaudeResponse{IsError: true, RawResult: ""},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.r.Failure()

			// Distinguish "no error expected" (IsError false) from
			// "error expected but with an empty message" via the IsError flag.
			wantNilErr := !tt.r.IsError
			if wantNilErr {
				if err != nil {
					t.Errorf("Failure() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatal("Failure() = nil, want non-nil error")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("Failure() error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestClaudeResponse_Result(t *testing.T) {
	tests := []struct {
		name string
		r    *ClaudeResponse
		want string
	}{
		{
			name: "returns raw result",
			r:    &ClaudeResponse{RawResult: "the model output"},
			want: "the model output",
		},
		{
			name: "empty raw result returns empty string",
			r:    &ClaudeResponse{RawResult: ""},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Result(); got != tt.want {
				t.Errorf("Result() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClaudeResponse_PersistArtifact(t *testing.T) {
	t.Run("missing artifact dir in context returns empty result", func(t *testing.T) {
		r := &ClaudeResponse{RawResult: "content"}
		path, err := r.PersistArtifact(context.Background(), "artifact")
		if err != nil {
			t.Fatalf("PersistArtifact() error = %v, want nil", err)
		}
		if path != "" {
			t.Errorf("PersistArtifact() path = %q, want empty", path)
		}
	})

	t.Run("successful persist writes raw result and returns path", func(t *testing.T) {
		tmpDir := t.TempDir()
		ctx := context.WithValue(context.Background(), artifactDirKey{}, tmpDir)
		r := &ClaudeResponse{RawResult: "the file content"}

		path, err := r.PersistArtifact(ctx, "my artifact")
		if err != nil {
			t.Fatalf("PersistArtifact() error = %v, want nil", err)
		}
		wantPath := filepath.Join(tmpDir, "my_artifact.md")
		if path != wantPath {
			t.Errorf("PersistArtifact() path = %q, want %q", path, wantPath)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read persisted file: %v", err)
		}
		if string(data) != "the file content" {
			t.Errorf("file content = %q, want %q", string(data), "the file content")
		}
	})

	t.Run("nonexistent directory returns wrapped error", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), artifactDirKey{}, filepath.Join(t.TempDir(), "does-not-exist"))
		r := &ClaudeResponse{RawResult: "content"}

		path, err := r.PersistArtifact(ctx, "artifact")
		if err == nil {
			t.Fatal("PersistArtifact() error = nil, want error")
		}
		if !strings.Contains(err.Error(), "persist agent artifact:") {
			t.Errorf("PersistArtifact() error = %v, want prefix %q", err, "persist agent artifact:")
		}
		if path != "" {
			t.Errorf("PersistArtifact() path = %q, want empty on error", path)
		}
	})
}
