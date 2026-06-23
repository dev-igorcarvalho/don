package primitives

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var execCommandContext = exec.CommandContext

type AgentProvider interface {
	ResolveProviderCmdLine(prompt string) (string, []string)
	Parse(out []byte, target any) error
}

// FailureChecker is an interface that result types can implement to indicate
// if the LLM returned a logical error (e.g. safety denial, API error).
type FailureChecker interface {
	Failure() error
}

type FoundationModelResult interface {
	Result() string
	PersistArtifact(ctx context.Context, artifactName string) (string, error)
}

// Agent is a reusable unit that runs prompts via an AgentProvider.
type Agent[T FoundationModelResult] struct {
	Name        string
	Provider    AgentProvider
	Description string
	Model       string
	System      string
	Prompt      string
	Before      func(ctx context.Context) error
	After       func(ctx context.Context) error
}

func (a *Agent[T]) isValid() error {
	if a.Provider == nil {
		return errors.New("agent provider is required")
	}
	if a.Name == "" {
		return errors.New("agent name is required")
	}
	if a.Model == "" {
		return errors.New("agent model is required")
	}
	if a.Prompt == "" {
		return errors.New("agent prompt is required")
	}
	return nil
}

// Run executes the agent lifecycle: validate, before hook, execute, after hook, parse.
func (a *Agent[T]) Run(ctx context.Context) (*AgentResponse, error) {
	if err := a.isValid(); err != nil {
		return nil, err
	}
	log := Logger(ctx)
	log.Info("agent starting", "name", a.Name, "model", a.Model)
	if err := a.runBefore(ctx); err != nil {
		return nil, err
	}
	out, err := a.execute(ctx)
	if err != nil {
		log.Error("agent failed", "name", a.Name, "error", err)
		return nil, err
	}
	if err := a.runAfter(ctx); err != nil {
		return nil, err
	}
	result, err := a.parseResult(out)
	if err != nil {
		log.Error("agent failed", "name", a.Name, "error", err)
		return nil, err
	}
	path, err := a.persist(ctx, *result)
	if err != nil {
		log.Error("failed to persist agent output", "name", a.Name, "error", err)
		return nil, err
	}
	log.Info("agent done", "name", a.Name)
	return &AgentResponse{ArtifactPath: path, ModelResponse: *result}, nil
}

func (a *Agent[T]) runBefore(ctx context.Context) error {
	if a.Before == nil {
		return nil
	}
	if err := a.Before(ctx); err != nil {
		return fmt.Errorf("%s: before: %w", a.Name, err)
	}
	return nil
}

func (a *Agent[T]) runAfter(ctx context.Context) error {
	if a.After == nil {
		return nil
	}
	if err := a.After(ctx); err != nil {
		return fmt.Errorf("%s: after: %w", a.Name, err)
	}
	return nil
}

func (a *Agent[T]) execute(ctx context.Context) ([]byte, error) {
	prompt, err := a.resolvePrompt()
	if err != nil {
		return nil, err
	}
	providerCmd, providerArgs := a.Provider.ResolveProviderCmdLine(prompt)
	out, err := execCommandContext(ctx, providerCmd, a.resolveArgs(providerArgs)...).Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("%s: %s exited %d: %s", a.Name, providerCmd, exitErr.ExitCode(), string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("%s: exec: %w", a.Name, err)
	}
	return out, nil
}

func (a *Agent[T]) parseResult(out []byte) (*T, error) {
	var r T
	if err := a.Provider.Parse(out, &r); err != nil {
		return nil, fmt.Errorf("%s: parse: %w\nraw output: %s", a.Name, err, string(out))
	}

	// If the result type implements FailureChecker, check for logical errors.
	if checker, ok := any(&r).(FailureChecker); ok {
		if err := checker.Failure(); err != nil {
			return nil, fmt.Errorf("%s: %w", a.Name, err)
		}
	}
	return &r, nil
}

// readPromptContent reads the raw prompt string, either directly from Prompt or from a file if Prompt ends in .md.
func (a *Agent[T]) readPromptContent() (string, error) {
	if !strings.HasSuffix(a.Prompt, ".md") {
		return a.Prompt, nil
	}
	data, err := os.ReadFile(a.Prompt)
	if err != nil {
		return "", fmt.Errorf("prompt file %q: %w", a.Prompt, err)
	}
	return string(data), nil
}

// resolvePrompt returns Agent.Prompt wrapped with formatting enforcer, or the file contents wrapped if Prompt ends in .md.
func (a *Agent[T]) resolvePrompt() (string, error) {
	content, err := a.readPromptContent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s \n %s", AgentResponseFormatEnforcerXml, content), nil
}

func (a *Agent[T]) resolveArgs(baseArgs []string) []string {
	args := append([]string(nil), baseArgs...)
	if a.Model != "" {
		args = append(args, "--model", a.Model)
	}
	if a.System != "" {
		args = append(args, "--system-prompt", a.System)
	}
	return args
}

func (a *Agent[T]) persist(ctx context.Context, out T) (string, error) {
	path, err := out.PersistArtifact(ctx, a.Name)
	if err != nil {
		return "", err
	}
	return path, nil
}
