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
	Name string

	// Before runs before fn. If it returns an error, fn and After are skipped.
	Before func(ctx context.Context) error

	// After runs after fn regardless of outcome.
	// It is for side effects only (logging, DB persistence, metrics).
	After func(ctx context.Context) error

	fn func(ctx context.Context) error // Injected from outside so you can code any standard Go func calling agents or other artifacts
}

// NewPipeline creates a Pipeline with the given name and business-logic func.
// The fn closure should capture any dependencies (agents, DB, logger, etc.).
func NewPipeline(name string, fn func(ctx context.Context) error) *Pipeline {
	return &Pipeline{Name: name, fn: fn}
}

func (p *Pipeline) isValid() error {
	if p.Name == "" {
		return errors.New("pipeline name is required")
	}
	if p.fn == nil {
		return errors.New("pipeline fn is required")
	}
	return nil
}

// Run enforces lifecycle concerns around the user-provided fn.
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

	finalErr := fnErr
	if finalErr == nil {
		finalErr = afterErr
	}

	p.logResult(ctx, finalErr)
	return finalErr
}

func (p *Pipeline) runBefore(ctx context.Context) error {
	if p.Before == nil {
		return nil
	}
	if err := p.Before(ctx); err != nil {
		return fmt.Errorf("%s: before: %w", p.Name, err)
	}
	return nil
}

func (p *Pipeline) runFn(ctx context.Context) error {
	if err := p.fn(ctx); err != nil {
		return fmt.Errorf("%s: %w", p.Name, err)
	}
	return nil
}

func (p *Pipeline) runAfter(ctx context.Context) error {
	if p.After == nil {
		return nil
	}
	if err := p.After(ctx); err != nil {
		return fmt.Errorf("%s: after: %w", p.Name, err)
	}
	return nil
}

func (p *Pipeline) logResult(ctx context.Context, err error) {
	log := Logger(ctx)
	if err != nil {
		log.Error("pipeline failed", "name", p.Name, "error", err)
	} else {
		log.Info("pipeline done", "name", p.Name)
	}
}
