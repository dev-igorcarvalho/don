package primitives

import (
	"context"
	"errors"
	"fmt"
)

type Workflow interface {
	Run(ctx context.Context) error
}

// Orchestrator runs a sequence of Workflows in the order they were given.
// It stops at the first error. It satisfies the Workflow interface itself,
// so orchestrators can be nested or passed anywhere a Workflow is expected.
type Orchestrator struct {
	Name      string
	workflows []Workflow
}

// NewOrchestrator creates an Orchestrator with the given workflows.
func NewOrchestrator(name string, workflows ...Workflow) *Orchestrator {
	return &Orchestrator{Name: name, workflows: workflows}
}

// AddWorkflow adds a workflow to the orchestrator.
func (o *Orchestrator) AddWorkflow(w Workflow) {
	o.workflows = append(o.workflows, w)
}

// AddAgent wraps an agent in a Pipeline and adds it to the orchestrator.
func (o *Orchestrator) AddAgent(name string, runFn func(ctx context.Context) error) {
	o.AddWorkflow(NewPipeline(name, runFn))
}

func (o *Orchestrator) isValid() error {
	if o.Name == "" {
		return errors.New("orchestrator name is required")
	}
	if len(o.workflows) == 0 {
		return errors.New("orchestrator requires at least one workflow")
	}
	return nil
}

// Run sets up a session, injects it into the context, then executes
// each workflow in order, stopping at the first error.
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

func (o *Orchestrator) runWorkflows(ctx context.Context) error {
	sessionID, _ := SessionID(ctx)
	log := Logger(ctx)
	log.Info("orchestrator starting", "name", o.Name, "workflows", len(o.workflows), "session", sessionID)

	for step, w := range o.workflows {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("%s: step %d: context done: %w", o.Name, step+1, err)
		}
		if err := w.Run(ctx); err != nil {
			return fmt.Errorf("%s: step %d: %w", o.Name, step+1, err)
		}
	}

	log.Info("orchestrator done", "name", o.Name, "session", sessionID)
	return nil
}
