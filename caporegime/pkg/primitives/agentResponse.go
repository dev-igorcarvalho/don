package primitives

// AgentResponse represents the outcome of an agent's execution.
// It contains the model's parsed response and the file path to any persisted artifact.
type AgentResponse struct {
	// ModelResponse is the parsed result from the foundation model.
	ModelResponse FoundationModelResult
	// ArtifactPath is the file path where the agent's output artifact was persisted.
	// It is empty if no artifact was persisted.
	ArtifactPath string
}
