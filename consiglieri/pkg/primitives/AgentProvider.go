package primitives

import (
	"encoding/json"
	"encoding/xml"
	"errors"
)

type ClaudeProvider struct{}

func (c ClaudeProvider) ResolveProviderCmdLine(prompt string) (string, []string) {
	return "claude", []string{prompt, "--output-format", "json", "--dangerously-skip-permissions"}
}

func (c ClaudeProvider) Parse(out []byte, target any) error {
	return json.Unmarshal(out, target)
}

type AgyProvider struct{}

func (a AgyProvider) ResolveProviderCmdLine(prompt string) (string, []string) {
	return "agy", []string{"-p", prompt, "--dangerously-skip-permissions"}
}

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
		return errors.New("unknown target type")
	}
}

type GeminiProvider struct{}

func (g GeminiProvider) ResolveProviderCmdLine(prompt string) (string, []string) {
	return "gemini", []string{"-p", prompt, "--output-format", "json", "--approval-mode", "auto_edit"}
}

func (g GeminiProvider) Parse(out []byte, target any) error {
	return json.Unmarshal(out, target)
}

// Maybe using --output-format stream-json to record in real time each interaction to show on the gui
// TODO: need to review these cmd line args
// var claudeCommandLineArgs = []string{"--output-format", "json", "--dangerously-skip-permissions"}
