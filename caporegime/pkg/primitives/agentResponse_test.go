package primitives

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClaudeResult_Failure(t *testing.T) {
	tests := []struct {
		name    string
		r       *ClaudeResult
		wantErr bool
		errText string
	}{
		{
			name: "no error",
			r: &ClaudeResult{
				IsError:   false,
				RawResult: "success message",
			},
			wantErr: false,
		},
		{
			name: "has error",
			r: &ClaudeResult{
				IsError:   true,
				RawResult: "something went wrong",
			},
			wantErr: true,
			errText: "something went wrong",
		},
		{
			name: "has error with empty result",
			r: &ClaudeResult{
				IsError:   true,
				RawResult: "",
			},
			wantErr: true,
			errText: "",
		},
		{
			name:    "nil receiver",
			r:       nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.r.Failure()
			if (err != nil) != tt.wantErr {
				t.Errorf("ClaudeResult.Failure() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errText {
				t.Errorf("ClaudeResult.Failure() error = %v, wantErrText %v", err.Error(), tt.errText)
			}
		})
	}
}

func TestClaudeResult_Result(t *testing.T) {
	tests := []struct {
		name string
		r    *ClaudeResult
		want string
	}{
		{
			name: "success result",
			r: &ClaudeResult{
				RawResult: "hello world",
			},
			want: "hello world",
		},
		{
			name: "nil receiver",
			r:    nil,
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Result(); got != tt.want {
				t.Errorf("ClaudeResult.Result() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClaudeResult_PersistArtifact(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "claude_persist_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name         string
		r            *ClaudeResult
		ctx          context.Context
		artifactName string
		wantErr      bool
		wantPath     bool
	}{
		{
			name:         "nil receiver",
			r:            nil,
			ctx:          context.Background(),
			artifactName: "test",
			wantErr:      false,
			wantPath:     false,
		},
		{
			name: "missing artifact dir in context",
			r: &ClaudeResult{
				RawResult: "some content",
			},
			ctx:          context.Background(),
			artifactName: "test",
			wantErr:      false,
			wantPath:     false,
		},
		{
			name: "successful persist",
			r: &ClaudeResult{
				RawResult: "some content",
			},
			ctx:          context.WithValue(context.Background(), artifactDirKey{}, tmpDir),
			artifactName: "My - Cool - Artifact",
			wantErr:      false,
			wantPath:     true,
		},
		{
			name: "empty artifact name still persists",
			r: &ClaudeResult{
				RawResult: "content without a name",
			},
			ctx:          context.WithValue(context.Background(), artifactDirKey{}, tmpDir),
			artifactName: "",
			wantErr:      false,
			wantPath:     true,
		},
		{
			name: "nonexistent artifact directory returns error",
			r: &ClaudeResult{
				RawResult: "some content",
			},
			ctx:          context.WithValue(context.Background(), artifactDirKey{}, filepath.Join(tmpDir, "does-not-exist")),
			artifactName: "test",
			wantErr:      true,
			wantPath:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := tt.r.PersistArtifact(tt.ctx, tt.artifactName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClaudeResult.PersistArtifact() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantPath && path == "" {
				t.Error("expected non-empty path, got empty string")
			}
			if !tt.wantPath && path != "" {
				t.Errorf("expected empty path, got %v", path)
			}
			if tt.wantPath {
				// Verify file content
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("failed to read persisted file: %v", err)
				}
				if string(data) != tt.r.RawResult {
					t.Errorf("file content = %s, want %s", string(data), tt.r.RawResult)
				}
			}
		})
	}
}

func TestClaudeResult_UsageAndModelUsageFields(t *testing.T) {
	r := &ClaudeResult{
		IsError:   false,
		RawResult: "final output",
		Usage: ClaudeUsage{
			InputTokens:              100,
			CacheCreationInputTokens: 20,
			CacheReadInputTokens:     5,
			OutputTokens:             50,
			ServerToolUse: ClaudeServerToolUse{
				WebSearchRequests: 2,
				WebFetchRequests:  1,
			},
			ServiceTier: "standard",
			CacheCreation: ClaudeCacheCreation{
				Ephemeral1HInputTokens: 10,
				Ephemeral5MInputTokens: 15,
			},
			InferenceGeo: "us-east",
			Speed:        "fast",
		},
		ModelUsage: ClaudeModelUsage{
			ClaudeSonnet46: ClaudeSonnet46Usage{
				InputTokens:              100,
				OutputTokens:             50,
				CacheReadInputTokens:     5,
				CacheCreationInputTokens: 20,
				WebSearchRequests:        2,
				CostUSD:                  0.123,
				ContextWindow:            200000,
				MaxOutputTokens:          8192,
			},
		},
	}

	if err := r.Failure(); err != nil {
		t.Errorf("ClaudeResult.Failure() = %v, want nil", err)
	}
	if got := r.Result(); got != "final output" {
		t.Errorf("ClaudeResult.Result() = %q, want %q", got, "final output")
	}
	if r.Usage.InputTokens != 100 || r.Usage.OutputTokens != 50 {
		t.Errorf("Usage token fields not set correctly: %+v", r.Usage)
	}
	if r.Usage.ServerToolUse.WebSearchRequests != 2 || r.Usage.ServerToolUse.WebFetchRequests != 1 {
		t.Errorf("Usage.ServerToolUse fields not set correctly: %+v", r.Usage.ServerToolUse)
	}
	if r.Usage.CacheCreation.Ephemeral1HInputTokens != 10 || r.Usage.CacheCreation.Ephemeral5MInputTokens != 15 {
		t.Errorf("Usage.CacheCreation fields not set correctly: %+v", r.Usage.CacheCreation)
	}
	if r.ModelUsage.ClaudeSonnet46.CostUSD != 0.123 {
		t.Errorf("ModelUsage.ClaudeSonnet46.CostUSD = %v, want 0.123", r.ModelUsage.ClaudeSonnet46.CostUSD)
	}
	if r.ModelUsage.ClaudeSonnet46.ContextWindow != 200000 {
		t.Errorf("ModelUsage.ClaudeSonnet46.ContextWindow = %v, want 200000", r.ModelUsage.ClaudeSonnet46.ContextWindow)
	}
}

func TestFoundationModelResponse_Failure(t *testing.T) {
	tests := []struct {
		name string
		r    FoundationModelResponse
	}{
		{
			name: "empty response",
			r:    FoundationModelResponse{},
		},
		{
			name: "populated response",
			r:    FoundationModelResponse{ReasoningProcess: "thinking...", Response: "the answer"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.Failure(); err != nil {
				t.Errorf("FoundationModelResponse.Failure() = %v, want nil", err)
			}
		})
	}
}

func TestFoundationModelResponse_Result(t *testing.T) {
	tests := []struct {
		name string
		r    FoundationModelResponse
		want string
	}{
		{
			name: "empty result",
			r:    FoundationModelResponse{Response: ""},
			want: "",
		},
		{
			name: "whitespace result",
			r:    FoundationModelResponse{Response: "   "},
			want: "   ",
		},
		{
			name: "normal result",
			r:    FoundationModelResponse{Response: "final answer"},
			want: "final answer",
		},
		{
			name: "multiline result",
			r:    FoundationModelResponse{Response: "line one\nline two"},
			want: "line one\nline two",
		},
		{
			name: "reasoning process does not affect result",
			r:    FoundationModelResponse{ReasoningProcess: "some chain of thought", Response: "the result"},
			want: "the result",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Result(); got != tt.want {
				t.Errorf("FoundationModelResponse.Result() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFoundationModelResponse_PersistArtifact(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		r            FoundationModelResponse
		ctx          context.Context
		artifactName string
		wantErr      bool
		wantPath     bool
	}{
		{
			name:         "missing artifact dir in context",
			r:            FoundationModelResponse{Response: "some content"},
			ctx:          context.Background(),
			artifactName: "test",
			wantErr:      false,
			wantPath:     false,
		},
		{
			name:         "nil context",
			r:            FoundationModelResponse{Response: "some content"},
			ctx:          nil,
			artifactName: "test",
			wantErr:      false,
			wantPath:     false,
		},
		{
			name:         "successful persist",
			r:            FoundationModelResponse{Response: "final report content"},
			ctx:          context.WithValue(context.Background(), artifactDirKey{}, tmpDir),
			artifactName: "My Report",
			wantErr:      false,
			wantPath:     true,
		},
		{
			name:         "empty artifact name still persists",
			r:            FoundationModelResponse{Response: "content without a name"},
			ctx:          context.WithValue(context.Background(), artifactDirKey{}, tmpDir),
			artifactName: "",
			wantErr:      false,
			wantPath:     true,
		},
		{
			name:         "nonexistent artifact directory returns error",
			r:            FoundationModelResponse{Response: "some content"},
			ctx:          context.WithValue(context.Background(), artifactDirKey{}, filepath.Join(tmpDir, "does-not-exist")),
			artifactName: "test",
			wantErr:      true,
			wantPath:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := tt.r.PersistArtifact(tt.ctx, tt.artifactName)
			if (err != nil) != tt.wantErr {
				t.Errorf("FoundationModelResponse.PersistArtifact() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantPath && path == "" {
				t.Error("expected non-empty path, got empty string")
			}
			if !tt.wantPath && path != "" {
				t.Errorf("expected empty path, got %v", path)
			}
			if tt.wantPath {
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("failed to read persisted file: %v", err)
				}
				if string(data) != tt.r.Response {
					t.Errorf("file content = %s, want %s", string(data), tt.r.Response)
				}
			}
		})
	}
}

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
			name:         "spaces and hyphens are sanitized",
			dir:          "/tmp/artifacts",
			artifactName: "My - Cool Report",
			want:         filepath.Join("/tmp/artifacts", "my_cool_report.md"),
		},
		{
			name:         "slashes and backslashes are sanitized",
			dir:          "/tmp/artifacts",
			artifactName: `path/to\report`,
			want:         filepath.Join("/tmp/artifacts", "path_to_report.md"),
		},
		{
			name:         "mixed case is lowercased",
			dir:          "/tmp/artifacts",
			artifactName: "UPPER_Case_Name",
			want:         filepath.Join("/tmp/artifacts", "upper_case_name.md"),
		},
		{
			name:         "leading and trailing separators are trimmed",
			dir:          "/tmp/artifacts",
			artifactName: "  - leading and trailing -  ",
			want:         filepath.Join("/tmp/artifacts", "leading_and_trailing.md"),
		},
		{
			name:         "consecutive separators collapse to one underscore",
			dir:          "/tmp/artifacts",
			artifactName: "many---dashes   here",
			want:         filepath.Join("/tmp/artifacts", "many_dashes_here.md"),
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
		path, err := PersistArtifactToFile(nil, "test", "content")
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
		defer os.Chmod(roDir, 0o755) // restore so TempDir cleanup can remove the directory

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
