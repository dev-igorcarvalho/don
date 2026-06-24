package primitives

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/dev-igorcarvalho/don/caporegime/pkg/utils"
)

type mockProvider struct {
	cmd  string
	args []string
}

func (m mockProvider) ResolveProviderCmdLine(prompt string) (string, []string) {
	return m.cmd, append(m.args, prompt)
}

func (m mockProvider) Parse(out []byte, target any) error {
	if err := json.Unmarshal(out, target); err == nil {
		return nil
	}
	switch t := target.(type) {
	case *string:
		*t = string(out)
		return nil
	case *any:
		*t = string(out)
		return nil
	case *testStringResult:
		*t = testStringResult(out)
		return nil
	case *testAnyResult:
		t.Val = string(out)
		return nil
	default:
		return json.Unmarshal(out, target)
	}
}

type testResult struct {
	Status string `json:"status"`
}

func (r testResult) Result() string {
	return r.Status
}

func (r testResult) PersistArtifact(ctx context.Context, artifactName string) (string, error) {
	dir, ok := ArtifactDir(ctx)
	if !ok || dir == "" {
		return "", nil
	}
	if dir == "/nonexistent-dir-for-test" {
		return "", errors.New("persist error")
	}
	path := filepath.Join(dir, "test_agent_123.md")
	if err := os.WriteFile(path, []byte(r.Result()), 0644); err != nil {
		return "", err
	}
	return path, nil
}

type testResultWithFailure struct {
	Status string `json:"status"`
	Err    string `json:"error,omitempty"`
}

func (t testResultWithFailure) Failure() error {
	if t.Err != "" {
		return errors.New(t.Err)
	}
	return nil
}

func (t testResultWithFailure) Result() string {
	return t.Status
}

func (t testResultWithFailure) PersistArtifact(ctx context.Context, artifactName string) (string, error) {
	return "", nil
}

type testStringResult string

func (r testStringResult) Result() string {
	return string(r)
}

func (r testStringResult) PersistArtifact(ctx context.Context, artifactName string) (string, error) {
	return "", nil
}

type testAnyResult struct {
	Val any
}

func (r testAnyResult) Result() string {
	return fmt.Sprintf("%v", r.Val)
}

func (r testAnyResult) PersistArtifact(ctx context.Context, artifactName string) (string, error) {
	return "", nil
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// The output is passed as the first argument after "--"
	args := os.Args
	for i := range args {
		if args[i] == "--" {
			args = args[i+1:]
			break
		}
	}

	if len(args) > 0 {
		fmt.Fprint(os.Stdout, args[0])
	}

	exitCode := 0
	if code := os.Getenv("GO_HELPER_EXIT_CODE"); code != "" {
		_, _ = fmt.Sscanf(code, "%d", &exitCode)
	}
	os.Exit(exitCode)
}

func mockExec(output string, exitCode int) func(ctx context.Context, name string, arg ...string) *exec.Cmd {
	return func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestHelperProcess", "--", output)
		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1", fmt.Sprintf("GO_HELPER_EXIT_CODE=%d", exitCode))
		return cmd
	}
}

func TestAgent_isValid(t *testing.T) {
	tests := []struct {
		name    string
		agent   Agent[FoundationModelResponse]
		wantErr bool
		errMsg  string
	}{
		{
			name: "missing provider",
			agent: Agent[FoundationModelResponse]{
				Name:   "test",
				Model:  "model",
				Prompt: "prompt",
			},
			wantErr: true,
			errMsg:  "agent provider is required",
		},
		{
			name: "missing name",
			agent: Agent[FoundationModelResponse]{
				Provider: ClaudeDefaultProvider{},
				Model:    "model",
				Prompt:   "prompt",
			},
			wantErr: true,
			errMsg:  "agent name is required",
		},
		{
			name: "missing model",
			agent: Agent[FoundationModelResponse]{
				Provider: ClaudeDefaultProvider{},
				Name:     "test",
				Prompt:   "prompt",
			},
			wantErr: true,
			errMsg:  "agent model is required",
		},
		{
			name: "missing prompt",
			agent: Agent[FoundationModelResponse]{
				Provider: ClaudeDefaultProvider{},
				Name:     "test",
				Model:    "model",
			},
			wantErr: true,
			errMsg:  "agent prompt is required",
		},
		{
			name: "valid agent",
			agent: Agent[FoundationModelResponse]{
				Provider: ClaudeDefaultProvider{},
				Name:     "test",
				Model:    "model",
				Prompt:   "prompt",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.agent.isValid()
			if (err != nil) != tt.wantErr {
				t.Errorf("isValid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("isValid() error msg = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestAgent_resolvePrompt(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_prompt*.md")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	content := "file content"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	tests := []struct {
		name    string
		prompt  string
		want    string
		wantErr bool
	}{
		{
			name:   "plain prompt",
			prompt: "direct prompt",
			want:   fmt.Sprintf("%s \n %s", AgentResponseFormatEnforcerXml, "direct prompt"),
		},
		{
			name:   "file prompt",
			prompt: tmpFile.Name(),
			want:   fmt.Sprintf("%s \n %s", AgentResponseFormatEnforcerXml, content),
		},
		{
			name:    "missing file prompt",
			prompt:  "nonexistent.md",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent[FoundationModelResponse]{Prompt: tt.prompt}
			got, err := a.resolvePrompt()
			if (err != nil) != tt.wantErr {
				t.Errorf("resolvePrompt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("resolvePrompt() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgent_parseResult(t *testing.T) {
	tests := []struct {
		name    string
		out     []byte
		want    any
		wantErr bool
	}{
		{
			name: "valid json",
			out:  []byte(`{"status": "ok"}`),
			want: &testResult{Status: "ok"},
		},
		{
			name:    "invalid json",
			out:     []byte(`invalid`),
			wantErr: true,
		},
		{
			name: "failure checker success",
			out:  []byte(`{"status": "ok"}`),
			want: &testResultWithFailure{Status: "ok"},
		},
		{
			name:    "failure checker failure",
			out:     []byte(`{"status": "error", "error": "something went wrong"}`),
			want:    &testResultWithFailure{Status: "error", Err: "something went wrong"},
			wantErr: true,
		},
		{
			name: "fallback string",
			out:  []byte(`raw string output`),
			want: testStringResult("raw string output"),
		},
		{
			name: "fallback any",
			out:  []byte(`raw any output`),
			want: &testAnyResult{Val: "raw any output"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.want == nil {
				return
			}
			switch reflect.TypeOf(tt.want) {
			case reflect.TypeOf(&testResult{}):
				a := &Agent[testResult]{Name: "test", Provider: mockProvider{}}
				got, err := a.parseResult(tt.out)
				if (err != nil) != tt.wantErr {
					t.Errorf("parseResult() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
					t.Errorf("parseResult() got = %+v, want %+v", got, tt.want)
				}
			case reflect.TypeOf(&testResultWithFailure{}):
				a := &Agent[testResultWithFailure]{Name: "test", Provider: mockProvider{}}
				got, err := a.parseResult(tt.out)
				if (err != nil) != tt.wantErr {
					t.Errorf("parseResult() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
					t.Errorf("parseResult() got = %+v, want %+v", got, tt.want)
				}
			case reflect.TypeOf(testStringResult("")):
				a := &Agent[testStringResult]{Name: "test", Provider: mockProvider{}}
				got, err := a.parseResult(tt.out)
				if (err != nil) != tt.wantErr {
					t.Errorf("parseResult() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && *got != tt.want.(testStringResult) {
					t.Errorf("parseResult() got = %v, want %v", *got, tt.want)
				}
			case reflect.TypeOf(&testAnyResult{}):
				a := &Agent[testAnyResult]{Name: "test", Provider: mockProvider{}}
				got, err := a.parseResult(tt.out)
				if (err != nil) != tt.wantErr {
					t.Errorf("parseResult() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
					t.Errorf("parseResult() got = %+v, want %+v", got, tt.want)
				}
			}
		})
	}
}

func TestAgent_execute(t *testing.T) {
	origExec := execCommandContext
	defer func() { execCommandContext = origExec }()

	tests := []struct {
		name         string
		agent        *Agent[FoundationModelResponse]
		expectedArgs []string
		output       string
		exitCode     int
		wantErr      bool
	}{
		{
			name: "success with model and system",
			agent: &Agent[FoundationModelResponse]{
				Name:     "test",
				Provider: mockProvider{cmd: "test-cmd", args: []string{"base1"}},
				Model:    "gpt-4",
				System:   "you are a bot",
				Prompt:   "hello",
			},
			expectedArgs: []string{"base1", fmt.Sprintf("%s \n %s", AgentResponseFormatEnforcerXml, "hello"), "--model", "gpt-4", "--system-prompt", "you are a bot"},
			output:       "success output",
			exitCode:     0,
			wantErr:      false,
		},
		{
			name: "exit error",
			agent: &Agent[FoundationModelResponse]{
				Name:     "test",
				Provider: mockProvider{cmd: "test-cmd", args: []string{}},
				Prompt:   "prompt",
			},
			output:   "error output",
			exitCode: 1,
			wantErr:  true,
		},
		{
			name: "non exit error (binary not found)",
			agent: &Agent[FoundationModelResponse]{
				Name:     "test",
				Provider: mockProvider{cmd: "nonexistentbinaryhere", args: []string{}},
				Prompt:   "prompt",
			},
			wantErr: true,
		},
		{
			name: "resolve prompt error",
			agent: &Agent[FoundationModelResponse]{
				Name:     "test",
				Provider: mockProvider{cmd: "test-cmd", args: []string{}},
				Prompt:   "nonexistent.md",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execCommandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
				if tt.name == "non exit error (binary not found)" {
					return exec.CommandContext(ctx, "nonexistentbinaryhere")
				}
				if tt.expectedArgs != nil {
					providerCmd, _ := tt.agent.Provider.ResolveProviderCmdLine(tt.agent.Prompt)
					if name != providerCmd {
						t.Errorf("execute() name = %v, want %v", name, providerCmd)
					}
					if !reflect.DeepEqual(arg, tt.expectedArgs) {
						t.Errorf("execute() args = %v, want %v", arg, tt.expectedArgs)
					}
				}
				return mockExec(tt.output, tt.exitCode)(ctx, name, arg...)
			}

			got, err := tt.agent.execute(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && string(got) != tt.output {
				t.Errorf("execute() got = %v, want %v", string(got), tt.output)
			}
		})
	}
}

func TestAgent_Run(t *testing.T) {
	origExec := execCommandContext
	defer func() { execCommandContext = origExec }()

	beforeCalled := false
	afterCalled := false

	tests := []struct {
		name     string
		agent    Agent[testResult]
		output   string
		exitCode int
		wantErr  bool
	}{
		{
			name: "full success",
			agent: Agent[testResult]{
				Name:     "test",
				Provider: mockProvider{cmd: "test"},
				Model:    "model",
				Prompt:   "prompt",
				Before: func(ctx context.Context) error {
					beforeCalled = true
					return nil
				},
				After: func(ctx context.Context) error {
					afterCalled = true
					return nil
				},
			},
			output:   `{"status": "ok"}`,
			exitCode: 0,
			wantErr:  false,
		},
		{
			name: "validation error (missing prompt)",
			agent: Agent[testResult]{
				Name:     "test",
				Provider: mockProvider{cmd: "test"},
				Model:    "model",
			},
			wantErr: true,
		},
		{
			name: "before error",
			agent: Agent[testResult]{
				Name:     "test",
				Provider: mockProvider{cmd: "test"},
				Model:    "model",
				Prompt:   "prompt",
				Before: func(ctx context.Context) error {
					return errors.New("before failed")
				},
			},
			wantErr: true,
		},
		{
			name: "execute error",
			agent: Agent[testResult]{
				Name:     "test",
				Provider: mockProvider{cmd: "test"},
				Model:    "model",
				Prompt:   "prompt",
			},
			output:   `failed execute`,
			exitCode: 1,
			wantErr:  true,
		},
		{
			name: "persist error",
			agent: Agent[testResult]{
				Name:     "test",
				Provider: mockProvider{cmd: "test"},
				Model:    "model",
				Prompt:   "prompt",
			},
			output:   `{"status": "ok"}`,
			exitCode: 0,
			wantErr:  true,
		},
		{
			name: "after error",
			agent: Agent[testResult]{
				Name:     "test",
				Provider: mockProvider{cmd: "test"},
				Model:    "model",
				Prompt:   "prompt",
				After: func(ctx context.Context) error {
					return errors.New("after failed")
				},
			},
			output:   `{"status": "ok"}`,
			exitCode: 0,
			wantErr:  true,
		},
		{
			name: "parseDefaultResponse error",
			agent: Agent[testResult]{
				Name:     "test",
				Provider: mockProvider{cmd: "test"},
				Model:    "model",
				Prompt:   "prompt",
			},
			output:   `malformed json`,
			exitCode: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeCalled = false
			afterCalled = false
			execCommandContext = mockExec(tt.output, tt.exitCode)

			ctx := context.Background()
			if tt.name == "persist error" {
				ctx = context.WithValue(ctx, artifactDirKey{}, "/nonexistent-dir-for-test")
			}

			got, err := tt.agent.Run(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				testRes, ok := got.ModelResponse.(testResult)
				if !ok || testRes.Status != "ok" {
					t.Errorf("Run() got ModelResponse = %v, want testResult with status ok", got.ModelResponse)
				}
				if !beforeCalled {
					t.Error("before hook was not called")
				}
				if !afterCalled {
					t.Error("after hook was not called")
				}
			}
		})
	}
}

func TestAgent_Run_Persist(t *testing.T) {
	origExec := execCommandContext
	defer func() { execCommandContext = origExec }()

	tmpDir, err := os.MkdirTemp("", "agent_persist_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	execCommandContext = mockExec(`{"status": "ok"}`, 0)

	a := Agent[testResult]{
		Name:     "Test Agent 123",
		Provider: mockProvider{cmd: "test"},
		Model:    "model",
		Prompt:   "prompt",
	}

	ctx := context.WithValue(context.Background(), artifactDirKey{}, tmpDir)
	_, err = a.Run(ctx)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "test_agent_123.md")
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read persisted file: %v", err)
	}

	if string(data) != `ok` {
		t.Errorf("persisted data got = %s, want %s", string(data), `ok`)
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain name",
			input:    "simpleName",
			expected: "simplename",
		},
		{
			name:     "spaces and hyphens",
			input:    "My - Agent - 123",
			expected: "my_agent_123",
		},
		{
			name:     "consecutive delimiters",
			input:    "///some--name\\\\   with spaces",
			expected: "some_name_with_spaces",
		},
		{
			name:     "leading and trailing",
			input:    "  -foo-bar-  ",
			expected: "foo_bar",
		},
		{
			name:     "complex pattern",
			input:    "agent/test\\run-1",
			expected: "agent_test_run_1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.SanitizeName(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizeName(%q) got %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
