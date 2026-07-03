package primitives

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

const (
	// claudeCmd is the executable name for the Claude CLI backend.
	claudeCmd = "claude"
	// flagPrompt is the flag used to specify the prompt to the model CLI.
	flagPrompt = "-p"
	// flagOutputFormat is the flag used to request a specific output format from the CLI.
	flagOutputFormat = "--output-format"
	// formatJSON is the value indicating JSON output format.
	formatJSON = "json"
	// flagDangerouslySkipPerms is the CLI flag to skip interactive permission prompts.
	flagDangerouslySkipPerms = "--dangerously-skip-permissions"
)

// ClaudeJsonProvider implements the AgentProvider interface for the Claude CLI backend.
type ClaudeJsonProvider struct {
	// AdditionalArgs allows specifying custom flags for the claude command.
	AdditionalArgs []string
}

// ResolveProviderCmdLine returns the command name and arguments to execute the Claude CLI for a given prompt.
func (c ClaudeJsonProvider) ResolveProviderCmdLine(prompt string) (string, []string) {
	baseArgs := []string{flagPrompt, prompt, flagDangerouslySkipPerms, flagOutputFormat, flagDangerouslySkipPerms}
	return resolveCmdLine(claudeCmd, baseArgs, c.AdditionalArgs)
}

// Parse unmarshals the JSON raw output from the Claude CLI execution into the target object.
// It returns any JSON deserialization error encountered.
func (c ClaudeJsonProvider) Parse(out []byte, target any) error {
	return parseDefaultResponse(out, target)
}

// ClaudeDefaultProvider implements the AgentProvider interface for the Claude CLI backend.
type ClaudeDefaultProvider struct {
	// AdditionalArgs allows specifying custom flags for the claude command.
	AdditionalArgs []string
}

// ResolveProviderCmdLine returns the command name and arguments to execute the Claude CLI for a given prompt.
func (c ClaudeDefaultProvider) ResolveProviderCmdLine(prompt string) (string, []string) {
	baseArgs := []string{prompt, flagOutputFormat, formatJSON, flagDangerouslySkipPerms}
	return resolveCmdLine(claudeCmd, baseArgs, c.AdditionalArgs)
}

// Parse unmarshals the JSON raw output from the Claude CLI execution into the target object.
// It returns any JSON deserialization error encountered.
func (c ClaudeDefaultProvider) Parse(out []byte, target any) error {
	return json.Unmarshal(out, target)
}

// resolveCmdLine appends provider-specific extra flags after the required base
// arguments, giving both provider implementations a single place to compose
// their final CLI invocation.
func resolveCmdLine(cmd string, baseArgs, additionalArgs []string) (string, []string) {
	return cmd, append(baseArgs, additionalArgs...)
}

// parseDefaultResponse deserializes raw provider output into target. *string and *any
// targets receive the raw output verbatim; any other pointer type is unmarshaled as JSON,
// so new FoundationModelResult types work without adding a case here.
func parseDefaultResponse(out []byte, target any) error {
	if target == nil {
		return errors.New("parse target is nil")
	}
	switch t := target.(type) {
	case *string:
		return parseAsRawString(t, out)
	case *any:
		return parseAsRawAny(t, out)
	default:
		return parseAsJSON(target, out)
	}
}

// parseAsRawString stores out verbatim in t, without attempting any deserialization.
func parseAsRawString(t *string, out []byte) error {
	if t == nil {
		return errors.New("target string pointer is nil")
	}
	*t = string(out)
	return nil
}

// parseAsRawAny stores out verbatim in t, without attempting any deserialization.
func parseAsRawAny(t *any, out []byte) error {
	if t == nil {
		return errors.New("target any pointer is nil")
	}
	*t = string(out)
	return nil
}

// parseAsJSON unmarshals out as JSON into target, which must be a non-nil pointer.
func parseAsJSON(target any, out []byte) error {
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Pointer {
		return fmt.Errorf("unknown target type: %T", target)
	}
	if rv.IsNil() {
		return errors.New("target pointer is nil")
	}
	if err := json.Unmarshal(out, target); err != nil {
		return fmt.Errorf("failed unmarshal json into %T: %w", target, err)
	}
	return nil
}
