package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func TestModelTransitions(t *testing.T) {
	// Prepare mock items
	items := []list.Item{
		WorkflowItem{
			name:        "Hello Workflow",
			filePath:    "hello.go",
			binaryPath:  "hello",
			description: "Hello",
			buildStatus: BuildStatusSuccess,
		},
	}

	model := NewMainModel(".", items)

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
