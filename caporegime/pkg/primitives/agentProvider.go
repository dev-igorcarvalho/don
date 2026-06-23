package primitives

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
)

const (
	// flagPrompt is the flag used to specify the prompt to the model CLI.
	flagPrompt = "-p"
	// flagOutputFormat is the flag used to request a specific output format from the CLI.
	flagOutputFormat = "--output-format"
	// formatJSON is the value indicating JSON output format.
	formatJSON = "json"
	// flagDangerouslySkipPerms is the CLI flag to skip interactive permission prompts.
	flagDangerouslySkipPerms = "--dangerously-skip-permissions"
	// flagApprovalMode is the flag used to set tool execution approval policy.
	flagApprovalMode = "--approval-mode"
	// approvalModeAutoEdit is the approval policy value that allows automatic file edits.
	approvalModeAutoEdit = "auto_edit"
)

// ClaudeProvider implements the AgentProvider interface for the Claude CLI backend.
type ClaudeProvider struct {
	// AdditionalArgs allows specifying custom flags for the claude command.
	AdditionalArgs []string
}

// NewClaudeProvider creates a new ClaudeProvider instance.
func NewClaudeProvider() ClaudeProvider {
	return ClaudeProvider{}
}

// ResolveProviderCmdLine returns the command name and arguments to execute the Claude CLI for a given prompt.
func (c ClaudeProvider) ResolveProviderCmdLine(prompt string) (string, []string) {
	args := []string{prompt, flagOutputFormat, formatJSON, flagDangerouslySkipPerms}
	args = append(args, c.AdditionalArgs...)
	return "claude", args
}

// Parse unmarshals the JSON raw output from the Claude CLI execution into the target object.
// It returns any JSON deserialization error encountered.
func (c ClaudeProvider) Parse(out []byte, target any) error {
	return json.Unmarshal(out, target)
}

// AgyProvider implements the AgentProvider interface for the Google Antigravity (Agy) CLI backend.
type AgyProvider struct {
	// AdditionalArgs allows specifying custom flags for the agy command.
	AdditionalArgs []string
}

// NewAgyProvider creates a new AgyProvider instance.
func NewAgyProvider() AgyProvider {
	return AgyProvider{}
}

// ResolveProviderCmdLine returns the command name and arguments to execute the Agy CLI for a given prompt.
func (a AgyProvider) ResolveProviderCmdLine(prompt string) (string, []string) {
	args := []string{flagPrompt, prompt, flagDangerouslySkipPerms}
	args = append(args, a.AdditionalArgs...)
	return "agy", args
}

// Parse decodes the raw output from the Agy CLI execution into the target object.
// Depending on the target type, it parses output as a raw string, an unmarshaled XML FoundationModelResponse,
// or returns an error if the target type is unknown.
func (a AgyProvider) Parse(out []byte, target any) error {
	switch t := target.(type) {
	case *string:
		*t = string(out)
		return nil
	case *any:
		*t = string(out)
		return nil
	case *FoundationModelResponse:
		return xml.Unmarshal(out, t)
	default:
		return fmt.Errorf("unknown target type: %T", target)
	}
}

// GeminiProvider implements the AgentProvider interface for the Gemini CLI backend.
type GeminiProvider struct {
	// AdditionalArgs allows specifying custom flags for the gemini command.
	AdditionalArgs []string
}

// NewGeminiProvider creates a new GeminiProvider instance.
func NewGeminiProvider() GeminiProvider {
	return GeminiProvider{}
}

// ResolveProviderCmdLine returns the command name and arguments to execute the Gemini CLI for a given prompt.
func (g GeminiProvider) ResolveProviderCmdLine(prompt string) (string, []string) {
	args := []string{flagPrompt, prompt, flagOutputFormat, formatJSON, flagApprovalMode, approvalModeAutoEdit}
	args = append(args, g.AdditionalArgs...)
	return "gemini", args
}

// Parse unmarshals the JSON raw output from the Gemini CLI execution into the target object.
// It returns any JSON deserialization error encountered.
func (g GeminiProvider) Parse(out []byte, target any) error {
	return json.Unmarshal(out, target)
}
