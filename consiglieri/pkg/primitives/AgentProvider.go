package primitives

type ClaudeProvider struct{}

func (c ClaudeProvider) ResolveProviderCmdLine() (string, []string) {
	return "claude", claudeCommandLineArgs
}

type GeminiProvider struct{}

func (g GeminiProvider) ResolveProviderCmdLine() (string, []string) {
	return "gemini", geminiCommandLineArgs
}

type AgyProvider struct{}

func (a AgyProvider) ResolveProviderCmdLine() (string, []string) {
	return "agy", agyCommandLineArgs
}

// Maybe using --output-format stream-json to record in real time each interaction to show on the gui
// TODO: need to review these cmd line args
var claudeCommandLineArgs = []string{"-p", "--output-format", "json", "--dangerously-skip-permissions"}
var geminiCommandLineArgs = []string{"-p", "--output-format", "json", "--approval-mode", "auto_edit"}
var agyCommandLineArgs = []string{"-p", "--output-format", "json", "--dangerously-skip-permissions"}
