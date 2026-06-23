package primitives

import (
	"context"
	"errors"
	"fmt"
)

var (
	// ErrOrchestratorNameRequired is returned when an orchestrator's name is empty.
	ErrOrchestratorNameRequired = errors.New("orchestrator name is required")
	// ErrOrchestratorWorkflowRequired is returned when an orchestrator has no workflows.
	ErrOrchestratorWorkflowRequired = errors.New("orchestrator requires at least one workflow")
)

// Workflow represents an executable unit of work or process in the system.
// Any component that implements this interface can be scheduled and executed
// sequentially by an Orchestrator.
type Workflow interface {
	// Run executes the workflow task using the provided context.
	// It returns an error if the execution fails or is interrupted.
	Run(ctx context.Context) error
}

// Orchestrator coordinates and executes a sequence of Workflow implementations in the order they were registered.
// It terminates execution at the first error encountered. Because Orchestrator implements the Workflow interface
// itself, orchestrators can be nested or passed anywhere a Workflow is expected.
type Orchestrator struct {
	// Name specifies the unique identifier for the orchestrator, used for session directory naming and logging.
	Name string
	// workflows holds the sequential list of Workflow tasks to be run.
	workflows []Workflow
}

// NewOrchestrator initializes and returns a new Orchestrator with the specified name and initial workflows.
// It duplicates the workflows slice to prevent subsequent external modifications.
func NewOrchestrator(name string, workflows ...Workflow) *Orchestrator {
	var ws []Workflow
	if len(workflows) > 0 {
		ws = make([]Workflow, len(workflows))
		copy(ws, workflows)
	}
	return &Orchestrator{Name: name, workflows: ws}
}

// AddWorkflow appends a new Workflow implementation to the list of tasks registered in the orchestrator.
func (o *Orchestrator) AddWorkflow(w Workflow) {
	o.workflows = append(o.workflows, w)
}

// AddAgent wraps a simple execution function in a Pipeline and appends it to the orchestrator's workflows.
// The name parameter identifies the created pipeline, and runFn specifies the functional task logic.
func (o *Orchestrator) AddAgent(name string, runFn func(ctx context.Context) error) {
	o.AddWorkflow(NewPipeline(name, runFn))
}

// Workflows returns a shallow copy of the Workflow list currently registered in the orchestrator.
// This prevents callers from mutating the orchestrator's internal slice. It returns nil if no workflows are registered.
func (o *Orchestrator) Workflows() []Workflow {
	if len(o.workflows) == 0 {
		return nil
	}
	cp := make([]Workflow, len(o.workflows))
	copy(cp, o.workflows)
	return cp
}

// isValid checks the configuration of the orchestrator, ensuring it has a non-empty name
// and contains at least one registered workflow task.
// It returns ErrOrchestratorNameRequired or ErrOrchestratorWorkflowRequired if validation fails, and nil otherwise.
func (o *Orchestrator) isValid() error {
	if o.Name == "" {
		return ErrOrchestratorNameRequired
	}
	if len(o.workflows) == 0 {
		return ErrOrchestratorWorkflowRequired
	}
	return nil
}

// Run validates the orchestrator configuration, initializes a new session, injects session details into the context,
// and executes each registered workflow sequentially. It terminates execution and returns the error on the first failure.
// It returns an error if validation fails, if session initialization fails, or if any workflow execution fails.
func (o *Orchestrator) Run(ctx context.Context) error {
	if err := o.isValid(); err != nil {
		return err
	}
	ctx, logFile, err := o.initSession(ctx)
	if err != nil {
		return fmt.Errorf("%s: session init: %w", o.Name, err)
	}
	defer logFile.Close()
	return o.runWorkflows(ctx)
}

// runWorkflows executes each registered workflow sequentially, logging start and completion times.
// It checks the context for cancellation before executing each step.
// It returns an error if the context is done or if any workflow execution fails.
func (o *Orchestrator) runWorkflows(ctx context.Context) error {
	sessionID, _ := SessionID(ctx)
	log := Logger(ctx)
	log.Info("orchestrator starting", "name", o.Name, "workflows", len(o.workflows), "session", sessionID)

	for step, w := range o.workflows {
		stepNum := step + 1
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("%s: step %d: context done: %w", o.Name, stepNum, err)
		}
		if err := w.Run(ctx); err != nil {
			return fmt.Errorf("%s: step %d: %w", o.Name, stepNum, err)
		}
	}

	log.Info("orchestrator done", "name", o.Name, "session", sessionID)
	return nil
}
