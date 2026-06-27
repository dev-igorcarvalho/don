// Package primitives defines the core agentic orchestration structures,
// including Agents, Providers, Pipelines, Orchestrators, and Sessions.
package primitives

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// execCommandContext is a package-level variable wrapping exec.CommandContext.
// It is used for mocking command execution during testing.
var execCommandContext = exec.CommandContext

// AgentProvider defines the contract for interacting with LLM CLI providers.
// It resolves the necessary command line to run the provider and parses its output.
type AgentProvider interface {
	// ResolveProviderCmdLine resolves the command name and arguments needed to execute
	// the provider for a given prompt. It returns the command path and args slice.
	ResolveProviderCmdLine(prompt string) (string, []string)

	// Parse deserializes the raw command output bytes into the target structure.
	// It returns an error if deserialization fails.
	Parse(out []byte, target any) error
}

// FailureChecker is an interface that result types can implement to indicate
// if the LLM returned a logical error (such as a safety denial or an API error).
type FailureChecker interface {
	// Failure returns an error representing the logical failure of the provider, or nil if no failure occurred.
	Failure() error
}

// FoundationModelResult represents the parsed result returned by a foundation model provider.
type FoundationModelResult interface {
	// Result returns the core text response or result content of the model execution.
	Result() string

	// PersistArtifact stores the execution output to the session's workspace.
	// It returns the absolute path of the saved artifact or an error if persistence fails.
	PersistArtifact(ctx context.Context, artifactName string) (string, error)
}

// Agent represents a reusable execution unit that runs prompts using an AgentProvider.
// It manages configuration, system prompts, validation, execution hooks (Before/After),
// and parsing of the results.
type Agent[T FoundationModelResult] struct {
	// Name is the unique name of the agent.
	Name string
	// Provider is the backend provider responsible for executing the prompt.
	Provider AgentProvider
	// Description is a human-readable description of what the agent does.
	Description string
	// Model is the specific LLM model version/identifier to use.
	Model string
	// System is the optional system prompt or instructions for the agent.
	System string
	// Prompt is either the raw prompt content or a filepath ending in .md containing the prompt.
	Prompt string
	// Before is an optional hook executed prior to running the prompt.
	Before func(ctx context.Context) error
	// After is an optional hook executed after the prompt is run and before the results are parsed.
	After func(ctx context.Context) error
}

// isValid verifies that all required fields on the agent are present and valid.
// It returns an error if any required field is missing.
func (a *Agent[T]) isValid() error {
	if a == nil {
		return errors.New("agent is nil")
	}
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

// Run executes the full lifecycle of the agent: validation, executing the Before hook,
// running the prompt via the provider, executing the After hook, parsing the output,
// and persisting the parsed result.
// It returns the AgentResponse containing the persisted path and model response, or an error if any stage fails.
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

// runBefore executes the Before hook of the agent if it is defined.
// It returns an error if the hook execution fails.
func (a *Agent[T]) runBefore(ctx context.Context) error {
	if a.Before == nil {
		return nil
	}
	if err := a.Before(ctx); err != nil {
		return fmt.Errorf("%s: before: %w", a.Name, err)
	}
	return nil
}

// runAfter executes the After hook of the agent if it is defined.
// It returns an error if the hook execution fails.
func (a *Agent[T]) runAfter(ctx context.Context) error {
	if a.After == nil {
		return nil
	}
	if err := a.After(ctx); err != nil {
		return fmt.Errorf("%s: after: %w", a.Name, err)
	}
	return nil
}

// execute resolves the prompt, triggers the provider command context, and executes it.
// It returns the raw command output bytes or an error if execution fails.
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

// parseResult deserializes the raw command output bytes into the target foundation model result type.
// If the result type implements FailureChecker, it also checks for semantic/logical errors.
// It returns a pointer to the parsed result or an error if parsing or semantic checking fails.
func (a *Agent[T]) parseResult(out []byte) (*T, error) {
	var r T
	if err := a.Provider.Parse(out, &r); err != nil {
		return nil, fmt.Errorf("%s: parseDefaultResponse: %w\nraw output: %s", a.Name, err, string(out))
	}

	// If the result type implements FailureChecker, check for logical errors.
	if checker, ok := any(&r).(FailureChecker); ok {
		if err := checker.Failure(); err != nil {
			return nil, fmt.Errorf("%s: %w", a.Name, err)
		}
	}
	return &r, nil
}

// readPromptContent reads the raw prompt string, either directly from the Prompt field or from a file if the field ends in .md.
// It returns the prompt contents or an error if a file read fails.
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

// resolvePrompt retrieves the prompt content and prefixes it with the XML format enforcer string.
// It returns the resolved prompt string or an error if reading the content fails.
func (a *Agent[T]) resolvePrompt() (string, error) {
	content, err := a.readPromptContent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s \n %s", AgentResponseFormatEnforcerXml, content), nil
}

// resolveArgs appends the optional model and system prompt flags to the provider's base command arguments.
// It returns the complete slice of arguments.
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

// persist saves the model response as an artifact via the response's PersistArtifact method.
// It returns the absolute path of the persisted artifact or an error if persistence fails.
func (a *Agent[T]) persist(ctx context.Context, out T) (string, error) {
	path, err := out.PersistArtifact(ctx, a.Name)
	if err != nil {
		return "", err
	}
	return path, nil
}
