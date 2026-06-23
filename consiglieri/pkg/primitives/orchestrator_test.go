package primitives

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type mockWorkflow struct {
	runCalled bool
	runErr    error
}

func (m *mockWorkflow) Run(ctx context.Context) error {
	m.runCalled = true
	return m.runErr
}

func TestOrchestrator_AddWorkflow(t *testing.T) {
	o := NewOrchestrator("test-orch")
	if len(o.Workflows()) != 0 {
		t.Errorf("expected 0 workflows, got %d", len(o.Workflows()))
	}

	w := &mockWorkflow{}
	o.AddWorkflow(w)

	if len(o.Workflows()) != 1 {
		t.Errorf("expected 1 workflow, got %d", len(o.Workflows()))
	}
	if o.Workflows()[0] != w {
		t.Error("workflow mismatch")
	}
}

func TestOrchestrator_AddAgent(t *testing.T) {
	o := NewOrchestrator("test-orch")
	o.AddAgent("test-agent", func(ctx context.Context) error {
		return nil
	})

	if len(o.Workflows()) != 1 {
		t.Errorf("expected 1 workflow, got %d", len(o.Workflows()))
	}

	p, ok := o.Workflows()[0].(*Pipeline)
	if !ok {
		t.Fatal("expected workflow to be a Pipeline")
	}
	if p.Name != "test-agent" {
		t.Errorf("expected pipeline name 'test-agent', got %s", p.Name)
	}
}

func TestOrchestrator_Run(t *testing.T) {
	defer os.RemoveAll(".agentic")

	t.Run("invalid orchestrator - no name", func(t *testing.T) {
		o := NewOrchestrator("", &mockWorkflow{})
		err := o.Run(context.Background())
		if !errors.Is(err, ErrOrchestratorNameRequired) {
			t.Errorf("expected ErrOrchestratorNameRequired error, got %v", err)
		}
	})

	t.Run("invalid orchestrator - no workflows", func(t *testing.T) {
		o := NewOrchestrator("test")
		err := o.Run(context.Background())
		if !errors.Is(err, ErrOrchestratorWorkflowRequired) {
			t.Errorf("expected ErrOrchestratorWorkflowRequired error, got %v", err)
		}
	})

	t.Run("successful run", func(t *testing.T) {
		m1 := &mockWorkflow{}
		m2 := &mockWorkflow{}
		o := NewOrchestrator("success-orch", m1, m2)

		err := o.Run(context.Background())
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if !m1.runCalled {
			t.Error("m1 was not called")
		}
		if !m2.runCalled {
			t.Error("m2 was not called")
		}
	})

	t.Run("stops at first error", func(t *testing.T) {
		m1 := &mockWorkflow{runErr: errors.New("m1 failed")}
		m2 := &mockWorkflow{}
		o := NewOrchestrator("fail-orch", m1, m2)

		err := o.Run(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "m1 failed") {
			t.Errorf("expected 'm1 failed' in error, got %v", err)
		}

		if !m1.runCalled {
			t.Error("m1 was not called")
		}
		if m2.runCalled {
			t.Error("m2 should not have been called")
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		m1 := &mockWorkflow{}
		o := NewOrchestrator("cancel-orch", m1)

		err := o.Run(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "context done") {
			t.Errorf("expected 'context done' error, got %v", err)
		}
	})

	t.Run("session init error", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "session_base_fail")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		err = os.WriteFile(filepath.Join(tmpDir, ".agentic"), []byte("blocker"), 0644)
		if err != nil {
			t.Fatal(err)
		}

		oldWd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		err = os.Chdir(tmpDir)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			_ = os.Chdir(oldWd)
		}()

		o := NewOrchestrator("session-err-orch", &mockWorkflow{})
		err = o.Run(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "session init:") {
			t.Errorf("expected session init error, got %v", err)
		}
	})
}
