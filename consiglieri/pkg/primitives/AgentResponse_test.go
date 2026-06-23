package primitives

import (
	"context"
	"os"
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
