package primitives

import (
	"encoding/json"
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
	flagPermissionMode       = "--permission-mode"
	PermissionModeDontAsk    = "dontAsk"
)

// ClaudeJsonProvider implements the AgentProvider interface for the Claude CLI backend.
type ClaudeJsonProvider struct {
	// AdditionalArgs allows specifying custom flags for the claude command.
	AdditionalArgs []string
}

// ResolveProviderCmdLine returns the command name and arguments to execute the Claude CLI for a given prompt.
func (c ClaudeJsonProvider) ResolveProviderCmdLine(prompt string) (string, []string) {
	baseArgs := []string{flagPrompt, prompt, flagPermissionMode, PermissionModeDontAsk, flagOutputFormat, formatJSON}
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
