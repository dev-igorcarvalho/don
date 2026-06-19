package primitives

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"testing"
)

type mockProvider struct {
	cmd  string
	args []string
}

func (m mockProvider) ResolveProviderCmdLine() (string, []string) {
	return m.cmd, m.args
}

type testResult struct {
	Status string `json:"status"`
}

type testResultWithFailure struct {
	Status string `json:"status"`
	Err    string `json:"error,omitempty"`
}

func (t *testResultWithFailure) Failure() error {
	if t.Err != "" {
		return errors.New(t.Err)
	}
	return nil
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
		agent   Agent[any]
		wantErr bool
		errMsg  string
	}{
		{
			name: "missing provider",
			agent: Agent[any]{
				Name:   "test",
				Model:  "model",
				Prompt: "prompt",
			},
			wantErr: true,
			errMsg:  "agent provider is required",
		},
		{
			name: "missing name",
			agent: Agent[any]{
				Provider: ClaudeProvider{},
				Model:    "model",
				Prompt:   "prompt",
			},
			wantErr: true,
			errMsg:  "agent name is required",
		},
		{
			name: "missing model",
			agent: Agent[any]{
				Provider: ClaudeProvider{},
				Name:     "test",
				Prompt:   "prompt",
			},
			wantErr: true,
			errMsg:  "agent model is required",
		},
		{
			name: "missing prompt",
			agent: Agent[any]{
				Provider: ClaudeProvider{},
				Name:     "test",
				Model:    "model",
			},
			wantErr: true,
			errMsg:  "agent prompt is required",
		},
		{
			name: "valid agent",
			agent: Agent[any]{
				Provider: ClaudeProvider{},
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
			want:   "direct prompt",
		},
		{
			name:   "file prompt",
			prompt: tmpFile.Name(),
			want:   content,
		},
		{
			name:    "missing file prompt",
			prompt:  "nonexistent.md",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent[any]{Prompt: tt.prompt}
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if reflect.TypeOf(tt.want) == reflect.TypeOf(&testResult{}) {
				a := &Agent[testResult]{Name: "test"}
				got, err := a.parseResult(tt.out)
				if (err != nil) != tt.wantErr {
					t.Errorf("parseResult() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
					t.Errorf("parseResult() got = %+v, want %+v", got, tt.want)
				}
			} else if reflect.TypeOf(tt.want) == reflect.TypeOf(&testResultWithFailure{}) {
				a := &Agent[testResultWithFailure]{Name: "test"}
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
		agent        *Agent[any]
		expectedArgs []string
		output       string
		exitCode     int
		wantErr      bool
	}{
		{
			name: "success with model and system",
			agent: &Agent[any]{
				Name:     "test",
				Provider: mockProvider{cmd: "test-cmd", args: []string{"base1"}},
				Model:    "gpt-4",
				System:   "you are a bot",
				Prompt:   "hello",
			},
			expectedArgs: []string{"base1", "--model", "gpt-4", "--system-prompt", "you are a bot", "hello"},
			output:       "success output",
			exitCode:     0,
			wantErr:      false,
		},
		{
			name: "exit error",
			agent: &Agent[any]{
				Name:     "test",
				Provider: mockProvider{cmd: "test-cmd", args: []string{}},
				Prompt:   "prompt",
			},
			output:   "error output",
			exitCode: 1,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execCommandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
				if tt.expectedArgs != nil {
					providerCmd, _ := tt.agent.Provider.ResolveProviderCmdLine()
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeCalled = false
			afterCalled = false
			execCommandContext = mockExec(tt.output, tt.exitCode)

			got, err := tt.agent.Run(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Status != "ok" {
					t.Errorf("Run() got status = %v, want ok", got.Status)
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
