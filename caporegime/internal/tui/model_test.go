package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func TestModelTransitions(t *testing.T) {
	tempDir := t.TempDir()

	filePath := filepath.Join(tempDir, "hello.go")
	binDir := filepath.Join(filepath.Dir(tempDir), "bin")
	binaryPath := filepath.Join(binDir, "hello")

	err := os.MkdirAll(binDir, 0755)
	if err != nil {
		t.Fatalf("failed to create bin dir: %v", err)
	}

	err = os.WriteFile(filePath, []byte("// name: Hello Workflow\n// description: Hello"), 0644)
	if err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}
	err = os.WriteFile(binaryPath, []byte("compiled binary"), 0755)
	if err != nil {
		t.Fatalf("failed to write binary file: %v", err)
	}

	// Make binary newer than source so compilation is skipped
	now := time.Now()
	_ = os.Chtimes(filePath, now.Add(-10*time.Second), now.Add(-10*time.Second))
	_ = os.Chtimes(binaryPath, now, now)

	// Prepare mock items
	items := []list.Item{
		WorkflowItem{
			name:        "Hello Workflow",
			filePath:    filePath,
			binaryPath:  binaryPath,
			description: "Hello",
			buildStatus: BuildStatusSuccess,
		},
	}

	model := NewMainModel(tempDir, items)

	// Verify initial state
	if model.running {
		t.Error("expected running to be false initially")
	}
	if len(model.logLines) != 0 {
		t.Error("expected logLines to be empty initially")
	}

	// Send enter key message
	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := newModel.(MainModel)

	if !m.running {
		t.Error("expected model to start running after pressing enter")
	}
	if cmd == nil {
		t.Error("expected command batch to be returned")
	}

	// Send log line message
	newModel, _ = m.Update(LogLineMsg("Step 1"))
	m = newModel.(MainModel)
	if len(m.logLines) != 2 { // first line is spawning process, second is Step 1
		t.Errorf("expected 2 log lines, got %d", len(m.logLines))
	}
	if m.logLines[1] != "Step 1" {
		t.Errorf("expected last log line to be 'Step 1', got %s", m.logLines[1])
	}

	// Send process finished message
	newModel, _ = m.Update(ProcessFinishedMsg{Err: nil})
	m = newModel.(MainModel)
	if m.running {
		t.Error("expected running to be false after ProcessFinishedMsg")
	}

	// Send esc key to clear
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = newModel.(MainModel)
	if len(m.logLines) != 0 {
		t.Error("expected logs to be cleared after pressing esc")
	}
}

func TestModelTransitionsNeedsCompile(t *testing.T) {
	tempDir := t.TempDir()

	filePath := filepath.Join(tempDir, "hello.go")
	binDir := filepath.Join(filepath.Dir(tempDir), "bin")
	binaryPath := filepath.Join(binDir, "hello")

	err := os.MkdirAll(binDir, 0755)
	if err != nil {
		t.Fatalf("failed to create bin dir: %v", err)
	}

	err = os.WriteFile(filePath, []byte("// name: Hello Workflow\n// description: Hello"), 0644)
	if err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}
	err = os.WriteFile(binaryPath, []byte("compiled binary"), 0755)
	if err != nil {
		t.Fatalf("failed to write binary file: %v", err)
	}

	// Make source newer than binary so compilation is required
	now := time.Now()
	_ = os.Chtimes(filePath, now, now)
	_ = os.Chtimes(binaryPath, now.Add(-10*time.Second), now.Add(-10*time.Second))

	// Prepare mock items
	items := []list.Item{
		WorkflowItem{
			name:        "Hello Workflow",
			filePath:    filePath,
			binaryPath:  binaryPath,
			description: "Hello",
			buildStatus: BuildStatusSuccess,
		},
	}

	model := NewMainModel(tempDir, items)

	// Send enter key message -> should trigger compilation
	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := newModel.(MainModel)

	if !m.running {
		t.Error("expected model to start running (transition to compiling view) after pressing enter")
	}
	if m.autoRunFilePath != filePath {
		t.Errorf("expected autoRunFilePath to be %s, got %s", filePath, m.autoRunFilePath)
	}
	if cmd == nil {
		t.Error("expected compilation command to be returned")
	}

	// Send WorkflowBuildFinishedMsg (success)
	finishedMsg := WorkflowBuildFinishedMsg{
		filePath:   filePath,
		binaryPath: binaryPath,
		err:        nil,
	}
	newModel, runCmd := m.Update(finishedMsg)
	m = newModel.(MainModel)

	if !m.running {
		t.Error("expected model to remain running to spawn process")
	}
	if m.autoRunFilePath != "" {
		t.Error("expected autoRunFilePath to be cleared after compilation finished")
	}
	if runCmd == nil {
		t.Error("expected runner check activity command to be returned")
	}

	// Mock incoming activity to ensure runner runs (simulate process exit)
	newModel, _ = m.Update(ProcessFinishedMsg{Err: nil})
	m = newModel.(MainModel)
	if m.running {
		t.Error("expected running to be false after ProcessFinishedMsg")
	}
}

func TestModelTransitionsCompileFailure(t *testing.T) {
	tempDir := t.TempDir()

	filePath := filepath.Join(tempDir, "hello.go")
	binaryPath := filepath.Join(tempDir, "hello")

	err := os.WriteFile(filePath, []byte("// name: Hello Workflow\n// description: Hello"), 0644)
	if err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	items := []list.Item{
		WorkflowItem{
			name:        "Hello Workflow",
			filePath:    filePath,
			binaryPath:  binaryPath,
			description: "Hello",
			buildStatus: BuildStatusSuccess,
		},
	}

	model := NewMainModel(tempDir, items)

	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := newModel.(MainModel)

	// Send failed compilation message
	finishedMsg := WorkflowBuildFinishedMsg{
		filePath: filePath,
		buildLog: "syntax error on line 42",
		err:      fmt.Errorf("exit status 2"),
	}
	newModel, cmd := m.Update(finishedMsg)
	m = newModel.(MainModel)

	if m.running {
		t.Error("expected running to be false after failed compilation")
	}
	if m.autoRunFilePath != "" {
		t.Error("expected autoRunFilePath to be cleared")
	}
	if cmd != nil {
		t.Error("expected command to be nil after compile failure")
	}

	// Check that log contains the compilation error
	logStr := strings.Join(m.logLines, "\n")
	if !strings.Contains(logStr, "syntax error on line 42") {
		t.Errorf("expected logs to contain syntax error, got: %s", logStr)
	}
}

func TestWorkspaceInitTransitions(t *testing.T) {
	tempDir := t.TempDir()

	// Create a dummy .go file so DiscoverWorkflows finds it
	dummyFile := filepath.Join(tempDir, "workflow.go")
	err := os.WriteFile(dummyFile, []byte("// name: Dummy\n// description: Test"), 0644)
	if err != nil {
		t.Fatalf("failed to write dummy file: %v", err)
	}

	// Start with no items, meaning we start in viewInit state
	model := NewMainModel(tempDir, nil)
	if model.state != viewInit {
		t.Errorf("expected state to be viewInit, got %v", model.state)
	}

	// Send successful workspaceInitMsg
	newModel, cmd := model.Update(workspaceInitMsg{err: nil})
	m := newModel.(MainModel)

	if m.state != viewInitSuccess {
		t.Errorf("expected state to transition to viewInitSuccess, got %v", m.state)
	}
	if cmd == nil {
		t.Error("expected cmd to be returned for background compilation dispatch")
	}

	// Send enter key to transition from viewInitSuccess to viewDashboard
	newModel, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(MainModel)

	if m.state != viewDashboard {
		t.Errorf("expected state to transition to viewDashboard, got %v", m.state)
	}
	if cmd != nil {
		t.Errorf("expected cmd to be nil, got %v", cmd)
	}
}
