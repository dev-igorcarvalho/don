package primitives

import (
	"context"
	"errors"
	"fmt"
)

// Pipeline wraps user-defined business logic with enforced cross-cutting concerns:
// context pre-check, optional Before/After hooks, and error wrapping.
// It satisfies the Workflow interface.
type Pipeline struct {
	// Name is the unique name of the pipeline, used for execution logging and error context.
	Name string

	// Before runs before the core pipeline function. If it returns an error, the core function and After are skipped.
	Before func(ctx context.Context) error

	// After runs after the core pipeline function, regardless of whether the function succeeded or failed.
	// It is intended for side effects such as logging, database persistence, or metrics collection.
	After func(ctx context.Context) error

	// fn is the core business-logic function of the pipeline. It is injected from the outside
	// to execute standard Go code, call agents, or perform other tasks.
	fn func(ctx context.Context) error
}

// NewPipeline creates and returns a new Pipeline with the specified name and core business-logic function.
// The caller should use the fn closure to capture any external dependencies, such as agents, databases, or loggers.
// It returns a pointer to the initialized Pipeline.
func NewPipeline(name string, fn func(ctx context.Context) error) *Pipeline {
	return &Pipeline{Name: name, fn: fn}
}

// isValid validates that the Pipeline has a non-empty name and a non-nil core function.
// It returns an error if any of these validation checks fail.
func (p *Pipeline) isValid() error {
	if p.Name == "" {
		return errors.New("pipeline name is required")
	}
	if p.fn == nil {
		return errors.New("pipeline fn is required")
	}
	return nil
}

// Run executes the Pipeline lifecycle, enforcing cross-cutting concerns around the core function.
// It first checks if the context is already cancelled, validates the pipeline structure,
// runs the Before hook (if present), executes the core function, and runs the After hook (if present).
// It returns the error from the core function or hooks, and wraps context execution errors.
func (p *Pipeline) Run(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("%s: context already done: %w", p.Name, err)
	}

	if err := p.isValid(); err != nil {
		return err
	}
	Logger(ctx).Info("pipeline starting", "name", p.Name)
	if err := p.runBefore(ctx); err != nil {
		return err
	}

	fnErr := p.runFn(ctx)
	afterErr := p.runAfter(ctx)

	err := fnErr
	if err == nil {
		err = afterErr
	} else if afterErr != nil {
		Logger(ctx).Error("pipeline after-hook failed", "name", p.Name, "error", afterErr)
	}

	p.logResult(ctx, err)
	return err
}

// runBefore executes the Before hook if one is registered.
// It wraps any returned hook error with the pipeline's name for context.
func (p *Pipeline) runBefore(ctx context.Context) error {
	if p.Before == nil {
		return nil
	}
	if err := p.Before(ctx); err != nil {
		return fmt.Errorf("%s: before: %w", p.Name, err)
	}
	return nil
}

// runFn executes the core pipeline function.
// It wraps any returned execution error with the pipeline's name.
func (p *Pipeline) runFn(ctx context.Context) error {
	if err := p.fn(ctx); err != nil {
		return fmt.Errorf("%s: %w", p.Name, err)
	}
	return nil
}

// runAfter executes the After hook if one is registered.
// It wraps any returned hook error with the pipeline's name for context.
func (p *Pipeline) runAfter(ctx context.Context) error {
	if p.After == nil {
		return nil
	}
	if err := p.After(ctx); err != nil {
		return fmt.Errorf("%s: after: %w", p.Name, err)
	}
	return nil
}

// logResult logs the outcome of the pipeline execution using the logger found in the context.
// It logs an error level message if the pipeline failed, or an info level message if it succeeded.
func (p *Pipeline) logResult(ctx context.Context, err error) {
	log := Logger(ctx)
	if err != nil {
		log.Error("pipeline failed", "name", p.Name, "error", err)
	} else {
		log.Info("pipeline done", "name", p.Name)
	}
}
