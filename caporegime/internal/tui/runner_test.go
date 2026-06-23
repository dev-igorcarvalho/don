package tui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	args := os.Args
	for i := range args {
		if args[i] == "--" {
			args = args[i+1:]
			break
		}
	}

	for _, arg := range args {
		fmt.Println(arg)
		time.Sleep(10 * time.Millisecond)
	}

	exitCode := 0
	if code := os.Getenv("GO_HELPER_EXIT_CODE"); code != "" {
		_, _ = fmt.Sscanf(code, "%d", &exitCode)
	}
	os.Exit(exitCode)
}

func mockExec(output []string, exitCode int) func(ctx context.Context, name string, arg ...string) *exec.Cmd {
	return func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		testArgs := []string{"-test.run=TestHelperProcess", "--"}
		testArgs = append(testArgs, output...)
		cmd := exec.CommandContext(ctx, os.Args[0], testArgs...)
		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1", fmt.Sprintf("GO_HELPER_EXIT_CODE=%d", exitCode))
		return cmd
	}
}

func TestRunner_StartAndRead(t *testing.T) {
	runner := NewRunner("dummy.go")
	runner.execCommand = mockExec([]string{"line 1", "line 2"}, 0)

	runner.Start(context.Background())

	var logs []string
	var finished bool
	var processErr error

	for msg := range runner.Channel() {
		switch m := msg.(type) {
		case LogLineMsg:
			logs = append(logs, string(m))
		case ProcessFinishedMsg:
			finished = true
			processErr = m.Err
		}
	}

	if !finished {
		t.Error("expected to receive ProcessFinishedMsg")
	}
	if processErr != nil {
		t.Errorf("expected no process error, got %v", processErr)
	}

	if len(logs) != 2 || logs[0] != "line 1" || logs[1] != "line 2" {
		t.Errorf("unexpected logs: %v", logs)
	}
}

func TestRunner_Cancel(t *testing.T) {
	runner := NewRunner("dummy.go")
	runner.execCommand = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestHelperProcess", "--", "slow line")
		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1", "GO_HELPER_EXIT_CODE=0")
		return cmd
	}

	runner.Start(context.Background())
	runner.Stop()

	var finished bool
	var processErr error

	for msg := range runner.Channel() {
		if m, ok := msg.(ProcessFinishedMsg); ok {
			finished = true
			processErr = m.Err
		}
	}

	if !finished {
		t.Error("expected to receive ProcessFinishedMsg on cancellation")
	}
	if processErr == nil {
		t.Error("expected context canceled or exit status error on cancel")
	}
}
