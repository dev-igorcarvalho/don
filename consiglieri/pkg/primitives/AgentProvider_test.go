package primitives

import (
	"reflect"
	"testing"
)

func TestResolveProviderCmdLine(t *testing.T) {
	tests := []struct {
		name     string
		provider AgentProvider
		wantCmd  string
		wantArgs []string
	}{
		{
			name:     "ClaudeProvider",
			provider: ClaudeProvider{},
			wantCmd:  "claude",
			wantArgs: []string{"-p", "--output-format", "json", "--dangerously-skip-permissions"},
		},
		{
			name:     "GeminiProvider",
			provider: GeminiProvider{},
			wantCmd:  "gemini",
			wantArgs: []string{"-p", "--output-format", "json", "--approval-mode", "auto_edit"},
		},
		{
			name:     "AgyProvider",
			provider: AgyProvider{},
			wantCmd:  "agy",
			wantArgs: []string{"-p", "--output-format", "json", "--dangerously-skip-permissions"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmd, gotArgs := tt.provider.ResolveProviderCmdLine()
			if gotCmd != tt.wantCmd {
				t.Errorf("%s.ResolveProviderCmdLine() gotCmd = %v, want %v", tt.name, gotCmd, tt.wantCmd)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("%s.ResolveProviderCmdLine() gotArgs = %v, want %v", tt.name, gotArgs, tt.wantArgs)
			}
		})
	}
}
