package primitives

import (
	"context"
	"encoding/xml"
	"os"
	"path/filepath"
	"testing"
)

// FoundationModelResponse represents a structured XML or JSON response returned by a foundation model.
// It includes both the reasoning process and the final result output.
//
// It is a test-only helper implementing FoundationModelResult, standing in for a concrete
// foundation-model result type so Agent[T] generics can be exercised in tests.
type FoundationModelResponse struct {
	// XMLName is the XML element name, mapped to model_response.
	XMLName xml.Name `xml:"model_response" json:"-"`
	// ReasoningProcess is the step-by-step thinking or chain of thought of the model.
	ReasoningProcess string `xml:"reasoning_process" json:"reasoning_process"`
	// Response is the final output result of the model.
	Response string `xml:"result" json:"result"`
}

// Failure returns nil to indicate that a FoundationModelResponse does not represent a logical failure.
func (r FoundationModelResponse) Failure() error {
	return nil
}

// Result returns the final output response of the foundation model.
func (r FoundationModelResponse) Result() string {
	return r.Response
}

// PersistArtifact writes the response result to a markdown file in the current session's artifact directory.
func (r FoundationModelResponse) PersistArtifact(ctx context.Context, artifactName string) (string, error) {
	return PersistArtifactToFile(ctx, artifactName, r.Result())
}

func TestClaudeResponse_UsageAndModelUsageFields(t *testing.T) {
	r := &ClaudeResponse{
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
		t.Errorf("ClaudeResponse.Failure() = %v, want nil", err)
	}
	if got := r.Result(); got != "final output" {
		t.Errorf("ClaudeResponse.Result() = %q, want %q", got, "final output")
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
