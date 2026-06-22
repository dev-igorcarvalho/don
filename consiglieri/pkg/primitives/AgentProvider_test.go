package primitives

import (
	"reflect"
	"testing"
)

func TestResolveProviderCmdLine(t *testing.T) {
	prompt := "test-prompt"
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
			wantArgs: []string{prompt, "--output-format", "json", "--dangerously-skip-permissions"},
		},
		{
			name:     "GeminiProvider",
			provider: GeminiProvider{},
			wantCmd:  "gemini",
			wantArgs: []string{"-p", prompt, "--output-format", "json", "--approval-mode", "auto_edit"},
		},
		{
			name:     "AgyProvider",
			provider: AgyProvider{},
			wantCmd:  "agy",
			wantArgs: []string{"-p", prompt, "--dangerously-skip-permissions"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmd, gotArgs := tt.provider.ResolveProviderCmdLine(prompt)
			if gotCmd != tt.wantCmd {
				t.Errorf("%s.ResolveProviderCmdLine() gotCmd = %v, want %v", tt.name, gotCmd, tt.wantCmd)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("%s.ResolveProviderCmdLine() gotArgs = %v, want %v", tt.name, gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestProviderParse(t *testing.T) {
	type testStruct struct {
		Val string `json:"val"`
	}

	// 1. ClaudeProvider.Parse
	t.Run("ClaudeProvider_Parse", func(t *testing.T) {
		p := ClaudeProvider{}
		var res testStruct
		err := p.Parse([]byte(`{"val":"hello"}`), &res)
		if err != nil {
			t.Fatalf("ClaudeProvider.Parse failed: %v", err)
		}
		if res.Val != "hello" {
			t.Errorf("expected hello, got %s", res.Val)
		}
	})

	// 2. GeminiProvider.Parse
	t.Run("GeminiProvider_Parse", func(t *testing.T) {
		p := GeminiProvider{}
		var res testStruct
		err := p.Parse([]byte(`{"val":"world"}`), &res)
		if err != nil {
			t.Fatalf("GeminiProvider.Parse failed: %v", err)
		}
		if res.Val != "world" {
			t.Errorf("expected world, got %s", res.Val)
		}
	})

	// 3. AgyProvider.Parse
	t.Run("AgyProvider_Parse", func(t *testing.T) {
		p := AgyProvider{}

		// String case
		var str string
		err := p.Parse([]byte("raw text"), &str)
		if err != nil {
			t.Fatalf("AgyProvider.Parse string failed: %v", err)
		}
		if str != "raw text" {
			t.Errorf("expected raw text, got %s", str)
		}

		// Any case
		var anyVal any
		err = p.Parse([]byte("raw text any"), &anyVal)
		if err != nil {
			t.Fatalf("AgyProvider.Parse any failed: %v", err)
		}
		if anyVal != "raw text any" {
			t.Errorf("expected raw text any, got %v", anyVal)
		}

		// XML case
		var res FoundationModelResponse
		xmlData := []byte(`<model_response><reasoning_process>step1;step2</reasoning_process><result>success</result></model_response>`)
		err = p.Parse(xmlData, &res)
		if err != nil {
			t.Fatalf("AgyProvider.Parse XML failed: %v", err)
		}
		if res.ReasoningProcess != "step1;step2" {
			t.Errorf("expected step1;step2, got %s", res.ReasoningProcess)
		}
		if res.Result() != "success" {
			t.Errorf("expected success, got %s", res.Result())
		}

		// Default case (struct/unknown target type)
		var resStruct testStruct
		err = p.Parse([]byte(`{"val":"agy"}`), &resStruct)
		if err == nil {
			t.Fatal("AgyProvider.Parse expected error for unknown target type, got nil")
		}
	})
}
