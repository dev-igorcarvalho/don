package primitives

import (
	"reflect"
	"strings"
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
			name:     "ClaudeJsonProvider",
			provider: ClaudeJsonProvider{},
			wantCmd:  "claude",
			wantArgs: []string{"-p", prompt, "--dangerously-skip-permissions", "--output-format", "--dangerously-skip-permissions"},
		},
		{
			name:     "ClaudeJsonProvider with AdditionalArgs",
			provider: ClaudeJsonProvider{AdditionalArgs: []string{"--foo", "bar"}},
			wantCmd:  "claude",
			wantArgs: []string{"-p", prompt, "--dangerously-skip-permissions", "--output-format", "--dangerously-skip-permissions", "--foo", "bar"},
		},
		{
			name:     "ClaudeJsonProvider with empty AdditionalArgs",
			provider: ClaudeJsonProvider{AdditionalArgs: []string{}},
			wantCmd:  "claude",
			wantArgs: []string{"-p", prompt, "--dangerously-skip-permissions", "--output-format", "--dangerously-skip-permissions"},
		},
		{
			name:     "ClaudeDefaultProvider",
			provider: ClaudeDefaultProvider{},
			wantCmd:  "claude",
			wantArgs: []string{prompt, "--output-format", "json", "--dangerously-skip-permissions"},
		},
		{
			name:     "ClaudeDefaultProvider with AdditionalArgs",
			provider: ClaudeDefaultProvider{AdditionalArgs: []string{"--foo", "bar"}},
			wantCmd:  "claude",
			wantArgs: []string{prompt, "--output-format", "json", "--dangerously-skip-permissions", "--foo", "bar"},
		},
		{
			name:     "ClaudeDefaultProvider with multiple AdditionalArgs",
			provider: ClaudeDefaultProvider{AdditionalArgs: []string{"--model", "sonnet", "--verbose"}},
			wantCmd:  "claude",
			wantArgs: []string{prompt, "--output-format", "json", "--dangerously-skip-permissions", "--model", "sonnet", "--verbose"},
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

// TestResolveProviderCmdLineEmptyPrompt verifies both providers handle an empty prompt string
// without panicking, still producing the fixed structural flags around the empty value.
func TestResolveProviderCmdLineEmptyPrompt(t *testing.T) {
	t.Run("ClaudeJsonProvider empty prompt", func(t *testing.T) {
		gotCmd, gotArgs := ClaudeJsonProvider{}.ResolveProviderCmdLine("")
		wantArgs := []string{"-p", "", "--dangerously-skip-permissions", "--output-format", "--dangerously-skip-permissions"}
		if gotCmd != "claude" || !reflect.DeepEqual(gotArgs, wantArgs) {
			t.Errorf("got (%v, %v), want (%v, %v)", gotCmd, gotArgs, "claude", wantArgs)
		}
	})

	t.Run("ClaudeDefaultProvider empty prompt", func(t *testing.T) {
		gotCmd, gotArgs := ClaudeDefaultProvider{}.ResolveProviderCmdLine("")
		wantArgs := []string{"", "--output-format", "json", "--dangerously-skip-permissions"}
		if gotCmd != "claude" || !reflect.DeepEqual(gotArgs, wantArgs) {
			t.Errorf("got (%v, %v), want (%v, %v)", gotCmd, gotArgs, "claude", wantArgs)
		}
	})
}

// TestResolveCmdLine directly exercises the shared helper used by both providers to compose
// their final CLI invocation, independent of any specific provider's base flags.
func TestResolveCmdLine(t *testing.T) {
	tests := []struct {
		name           string
		cmd            string
		baseArgs       []string
		additionalArgs []string
		wantCmd        string
		wantArgs       []string
	}{
		{
			name:           "nil additional args",
			cmd:            "mycmd",
			baseArgs:       []string{"a", "b"},
			additionalArgs: nil,
			wantCmd:        "mycmd",
			wantArgs:       []string{"a", "b"},
		},
		{
			name:           "empty base and additional args",
			cmd:            "mycmd",
			baseArgs:       []string{},
			additionalArgs: []string{},
			wantCmd:        "mycmd",
			wantArgs:       []string{},
		},
		{
			name:           "merges base and additional",
			cmd:            "mycmd",
			baseArgs:       []string{"a"},
			additionalArgs: []string{"b", "c"},
			wantCmd:        "mycmd",
			wantArgs:       []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmd, gotArgs := resolveCmdLine(tt.cmd, tt.baseArgs, tt.additionalArgs)
			if gotCmd != tt.wantCmd {
				t.Errorf("resolveCmdLine() gotCmd = %v, want %v", gotCmd, tt.wantCmd)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("resolveCmdLine() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

// testStruct is a small JSON-tagged struct used across Parse tests to exercise
// the default (arbitrary struct) dispatch branch of parseDefaultResponse/parseAsJSON.
type testStruct struct {
	Val string `json:"val"`
}

func TestClaudeJsonProviderParse(t *testing.T) {
	p := ClaudeJsonProvider{}

	t.Run("string target receives raw bytes verbatim", func(t *testing.T) {
		var str string
		err := p.Parse([]byte("raw text"), &str)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		if str != "raw text" {
			t.Errorf("expected %q, got %q", "raw text", str)
		}
	})

	t.Run("any target receives raw bytes verbatim as string", func(t *testing.T) {
		var anyVal any
		err := p.Parse([]byte("raw text any"), &anyVal)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		if anyVal != "raw text any" {
			t.Errorf("expected %q, got %v", "raw text any", anyVal)
		}
	})

	t.Run("struct target is JSON-unmarshaled", func(t *testing.T) {
		var res FoundationModelResponse
		err := p.Parse([]byte(`{"reasoning_process":"step1;step2","result":"success"}`), &res)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		if res.ReasoningProcess != "step1;step2" {
			t.Errorf("expected step1;step2, got %s", res.ReasoningProcess)
		}
		if res.Result() != "success" {
			t.Errorf("expected success, got %s", res.Result())
		}
	})

	t.Run("struct target with malformed JSON errors", func(t *testing.T) {
		var res testStruct
		err := p.Parse([]byte(`not json`), &res)
		if err == nil {
			t.Fatal("expected error for malformed JSON, got nil")
		}
		if !strings.Contains(err.Error(), "failed unmarshal json into *primitives.testStruct") {
			t.Errorf("expected error to mention target type, got: %v", err)
		}
	})

	t.Run("non-pointer struct target errors as unknown", func(t *testing.T) {
		err := p.Parse([]byte(`{"val":"x"}`), testStruct{})
		if err == nil {
			t.Fatal("expected error for non-pointer target, got nil")
		}
		if !strings.Contains(err.Error(), "unknown target type: primitives.testStruct") {
			t.Errorf("expected 'unknown target type' error, got: %v", err)
		}
	})

	t.Run("nil struct pointer target errors", func(t *testing.T) {
		var nilRes *testStruct
		err := p.Parse([]byte(`{"val":"x"}`), nilRes)
		if err == nil || err.Error() != "target pointer is nil" {
			t.Errorf("expected 'target pointer is nil', got %v", err)
		}
	})

	t.Run("nil target interface errors", func(t *testing.T) {
		err := p.Parse([]byte("raw text"), nil)
		if err == nil || err.Error() != "parse target is nil" {
			t.Errorf("expected 'parse target is nil', got %v", err)
		}
	})

	t.Run("typed nil string pointer errors", func(t *testing.T) {
		var nilStr *string
		err := p.Parse([]byte("raw text"), nilStr)
		if err == nil || err.Error() != "target string pointer is nil" {
			t.Errorf("expected 'target string pointer is nil', got %v", err)
		}
	})

	t.Run("typed nil any pointer errors", func(t *testing.T) {
		var nilAny *any
		err := p.Parse([]byte("raw text"), nilAny)
		if err == nil || err.Error() != "target any pointer is nil" {
			t.Errorf("expected 'target any pointer is nil', got %v", err)
		}
	})
}

func TestClaudeDefaultProviderParse(t *testing.T) {
	p := ClaudeDefaultProvider{}

	t.Run("struct target is JSON-unmarshaled", func(t *testing.T) {
		var res testStruct
		err := p.Parse([]byte(`{"val":"hello"}`), &res)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		if res.Val != "hello" {
			t.Errorf("expected hello, got %s", res.Val)
		}
	})

	t.Run("FoundationModelResponse JSON round trip", func(t *testing.T) {
		var res FoundationModelResponse
		err := p.Parse([]byte(`{"reasoning_process":"think","result":"done"}`), &res)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		if res.ReasoningProcess != "think" || res.Result() != "done" {
			t.Errorf("unexpected result: %+v", res)
		}
	})

	t.Run("malformed JSON errors", func(t *testing.T) {
		var res testStruct
		err := p.Parse([]byte(`not json`), &res)
		if err == nil {
			t.Fatal("expected error for malformed JSON, got nil")
		}
	})

	t.Run("non-JSON raw text into string pointer errors, unlike ClaudeJsonProvider", func(t *testing.T) {
		// ClaudeDefaultProvider.Parse always JSON-unmarshals directly; it does not
		// special-case *string/*any the way ClaudeJsonProvider (via parseDefaultResponse) does.
		var str string
		err := p.Parse([]byte("raw text"), &str)
		if err == nil {
			t.Fatal("expected error for non-JSON input into *string, got nil")
		}
	})

	t.Run("nil target errors", func(t *testing.T) {
		var res *testStruct
		err := p.Parse([]byte(`{"val":"x"}`), res)
		if err == nil {
			t.Fatal("expected error for nil pointer target, got nil")
		}
	})
}

// TestParseDefaultResponse directly exercises the three-way dispatch performed by
// parseDefaultResponse, independent of any specific provider.
func TestParseDefaultResponse(t *testing.T) {
	t.Run("nil target interface", func(t *testing.T) {
		err := parseDefaultResponse([]byte("data"), nil)
		if err == nil || err.Error() != "parse target is nil" {
			t.Errorf("expected 'parse target is nil', got %v", err)
		}
	})

	t.Run("dispatches *string to raw string branch", func(t *testing.T) {
		var s string
		err := parseDefaultResponse([]byte("hello world"), &s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s != "hello world" {
			t.Errorf("expected %q, got %q", "hello world", s)
		}
	})

	t.Run("dispatches *any to raw any branch", func(t *testing.T) {
		var a any
		err := parseDefaultResponse([]byte("hello any"), &a)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a != "hello any" {
			t.Errorf("expected %q, got %v", "hello any", a)
		}
	})

	t.Run("dispatches arbitrary struct pointer to JSON branch", func(t *testing.T) {
		var res testStruct
		err := parseDefaultResponse([]byte(`{"val":"json-branch"}`), &res)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Val != "json-branch" {
			t.Errorf("expected json-branch, got %s", res.Val)
		}
	})

	t.Run("JSON branch propagates unmarshal errors", func(t *testing.T) {
		var res testStruct
		err := parseDefaultResponse([]byte(`{invalid`), &res)
		if err == nil {
			t.Fatal("expected error for malformed JSON, got nil")
		}
		if !strings.Contains(err.Error(), "failed unmarshal json into *primitives.testStruct") {
			t.Errorf("expected wrapped unmarshal error, got: %v", err)
		}
	})
}

func TestParseAsRawString(t *testing.T) {
	t.Run("nil pointer errors", func(t *testing.T) {
		err := parseAsRawString(nil, []byte("x"))
		if err == nil || err.Error() != "target string pointer is nil" {
			t.Errorf("expected 'target string pointer is nil', got %v", err)
		}
	})

	t.Run("stores bytes as string verbatim", func(t *testing.T) {
		var s string
		err := parseAsRawString(&s, []byte("verbatim content"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s != "verbatim content" {
			t.Errorf("expected %q, got %q", "verbatim content", s)
		}
	})

	t.Run("empty bytes yields empty string", func(t *testing.T) {
		var s string
		err := parseAsRawString(&s, []byte{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s != "" {
			t.Errorf("expected empty string, got %q", s)
		}
	})
}

func TestParseAsRawAny(t *testing.T) {
	t.Run("nil pointer errors", func(t *testing.T) {
		err := parseAsRawAny(nil, []byte("x"))
		if err == nil || err.Error() != "target any pointer is nil" {
			t.Errorf("expected 'target any pointer is nil', got %v", err)
		}
	})

	t.Run("stores bytes as string verbatim", func(t *testing.T) {
		var a any
		err := parseAsRawAny(&a, []byte("verbatim any content"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a != "verbatim any content" {
			t.Errorf("expected %q, got %v", "verbatim any content", a)
		}
	})
}

func TestParseAsJSON(t *testing.T) {
	t.Run("non-pointer target errors as unknown", func(t *testing.T) {
		err := parseAsJSON(testStruct{}, []byte(`{"val":"x"}`))
		if err == nil {
			t.Fatal("expected error for non-pointer target, got nil")
		}
		if !strings.Contains(err.Error(), "unknown target type: primitives.testStruct") {
			t.Errorf("expected 'unknown target type' error, got: %v", err)
		}
	})

	t.Run("nil pointer target errors", func(t *testing.T) {
		var t1 *testStruct
		err := parseAsJSON(t1, []byte(`{"val":"x"}`))
		if err == nil || err.Error() != "target pointer is nil" {
			t.Errorf("expected 'target pointer is nil', got %v", err)
		}
	})

	t.Run("valid JSON unmarshals successfully", func(t *testing.T) {
		var res testStruct
		err := parseAsJSON(&res, []byte(`{"val":"parsed"}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Val != "parsed" {
			t.Errorf("expected parsed, got %s", res.Val)
		}
	})

	t.Run("malformed JSON returns wrapped error", func(t *testing.T) {
		var res testStruct
		err := parseAsJSON(&res, []byte(`{"val":`))
		if err == nil {
			t.Fatal("expected error for malformed JSON, got nil")
		}
		if !strings.Contains(err.Error(), "failed unmarshal json into *primitives.testStruct") {
			t.Errorf("expected wrapped error message, got: %v", err)
		}
	})
}
