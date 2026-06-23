package primitives

import (
	"context"
	"don_consiglieri/pkg/utils"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// AgentResponse represents the outcome of an agent's execution.
// It contains the model's parsed response and the file path to any persisted artifact.
type AgentResponse struct {
	// ModelResponse is the parsed result from the foundation model.
	ModelResponse FoundationModelResult
	// ArtifactPath is the file path where the agent's output artifact was persisted.
	// It is empty if no artifact was persisted.
	ArtifactPath string
}

// FoundationModelResponse represents a structured XML or JSON response returned by a foundation model.
// It includes both the reasoning process and the final result output.
type FoundationModelResponse struct {
	// XMLName is the XML element name, mapped to model_response.
	XMLName xml.Name `xml:"model_response" json:"-"`
	// ReasoningProcess is the step-by-step thinking or chain of thought of the model.
	ReasoningProcess string `xml:"reasoning_process" json:"reasoning_process"`
	// Response is the final output result of the model.
	Response string `xml:"result" json:"result"`
}

// Failure returns nil to indicate that a FoundationModelResponse does not represent a logical failure.
func (r FoundationModelResponse) Failure() error {
	return nil
}

// Result returns the final output response of the foundation model.
func (r FoundationModelResponse) Result() string {
	return r.Response
}

// PersistArtifact writes the response result to a markdown file in the current session's artifact directory.
// It returns the file path of the saved artifact and any file system error encountered.
func (r FoundationModelResponse) PersistArtifact(ctx context.Context, artifactName string) (string, error) {
	return PersistArtifactToFile(ctx, artifactName, r.Result())
}

// PersistArtifactToFile writes the artifact content to a file in the artifact directory specified in the context.
// It returns the absolute path of the written file and any file system error encountered.
// If the artifact directory is not configured in the context, it returns an empty path and a nil error.
func PersistArtifactToFile(ctx context.Context, artifactName, content string) (string, error) {
	dir, ok := ArtifactDir(ctx)
	if !ok || dir == "" {
		return "", nil
	}
	filename := fmt.Sprintf("%s.md", utils.SanitizeName(artifactName))
	path := filepath.Join(dir, filename)
	Logger(ctx).Info("persisting agent artifact", "name", artifactName, "path", path)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("persist agent artifact: %w", err)
	}
	return path, nil
}

// ClaudeResult represents the detailed metadata, performance metrics, and output from a Claude execution.
type ClaudeResult struct {
	// Type specifies the execution type of the Claude process.
	Type string `json:"type"`
	// Subtype specifies the sub-execution type or variant.
	Subtype string `json:"subtype"`
	// IsError indicates whether the Claude process execution resulted in an error.
	IsError bool `json:"is_error"`
	// ApiErrorStatus holds any error details returned by the Claude API.
	ApiErrorStatus interface{} `json:"api_error_status"`
	// DurationMs is the total execution duration in milliseconds.
	DurationMs int `json:"duration_ms"`
	// DurationApiMs is the time spent waiting for the API response in milliseconds.
	DurationApiMs int `json:"duration_api_ms"`
	// TtftMs is the time to first token in milliseconds.
	TtftMs int `json:"ttft_ms"`
	// TtftStreamMs is the time to first streamed token in milliseconds.
	TtftStreamMs int `json:"ttft_stream_ms"`
	// TimeToRequestMs is the time taken to initiate the request in milliseconds.
	TimeToRequestMs int `json:"time_to_request_ms"`
	// NumTurns is the number of interaction turns in the conversation.
	NumTurns int `json:"num_turns"`
	// RawResult is the raw string output returned by the Claude model.
	RawResult string `json:"result"`
	// StopReason indicates why the model execution stopped (e.g. stop sequence, max tokens).
	StopReason string `json:"stop_reason"`
	// SessionId is the identifier of the Claude execution session.
	SessionId string `json:"session_id"`
	// TotalCostUsd is the total cost of the execution in USD.
	TotalCostUsd float64 `json:"total_cost_usd"`
	// Usage contains token consumption statistics for the execution.
	Usage struct {
		// InputTokens is the number of input tokens consumed.
		InputTokens int `json:"input_tokens"`
		// CacheCreationInputTokens is the number of input tokens written to cache.
		CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
		// CacheReadInputTokens is the number of cached input tokens read.
		CacheReadInputTokens int `json:"cache_read_input_tokens"`
		// OutputTokens is the number of output tokens generated.
		OutputTokens int `json:"output_tokens"`
		// ServerToolUse tracks web requests made by the model's server tools.
		ServerToolUse struct {
			// WebSearchRequests is the count of web searches executed.
			WebSearchRequests int `json:"web_search_requests"`
			// WebFetchRequests is the count of web fetches executed.
			WebFetchRequests int `json:"web_fetch_requests"`
		} `json:"server_tool_use"`
		// ServiceTier is the pricing/service tier used.
		ServiceTier string `json:"service_tier"`
		// CacheCreation tracks cache write details.
		CacheCreation struct {
			// Ephemeral1HInputTokens is the number of ephemeral 1-hour cached tokens created.
			Ephemeral1HInputTokens int `json:"ephemeral_1h_input_tokens"`
			// Ephemeral5MInputTokens is the number of ephemeral 5-minute cached tokens created.
			Ephemeral5MInputTokens int `json:"ephemeral_5m_input_tokens"`
		} `json:"cache_creation"`
		// InferenceGeo is the geographical region where inference was processed.
		InferenceGeo string `json:"inference_geo"`
		// Iterations holds details about agent execution iterations.
		Iterations []interface{} `json:"iterations"`
		// Speed is the relative execution speed tier or mode.
		Speed string `json:"speed"`
	} `json:"usage"`
	// ModelUsage tracks specific model usage and cost details.
	ModelUsage struct {
		// ClaudeSonnet46 tracks input, output, cache tokens and costs specific to Claude Sonnet 4.6.
		ClaudeSonnet46 struct {
			// InputTokens is the number of input tokens consumed.
			InputTokens int `json:"inputTokens"`
			// OutputTokens is the number of output tokens generated.
			OutputTokens int `json:"outputTokens"`
			// CacheReadInputTokens is the number of cache read tokens.
			CacheReadInputTokens int `json:"cacheReadInputTokens"`
			// CacheCreationInputTokens is the number of cache creation tokens.
			CacheCreationInputTokens int `json:"cacheCreationInputTokens"`
			// WebSearchRequests is the count of web search requests.
			WebSearchRequests int `json:"webSearchRequests"`
			// CostUSD is the cost in USD.
			CostUSD float64 `json:"costUSD"`
			// ContextWindow is the size of the context window.
			ContextWindow int `json:"contextWindow"`
			// MaxOutputTokens is the maximum output tokens limit.
			MaxOutputTokens int `json:"maxOutputTokens"`
		} `json:"claude-sonnet-4-6"`
	} `json:"modelUsage"`
	// PermissionDenials logs any user permission denials during the execution.
	PermissionDenials []interface{} `json:"permission_denials"`
	// TerminalReason explains why the execution terminated.
	TerminalReason string `json:"terminal_reason"`
	// FastModeState reports the status of fast execution mode.
	FastModeState string `json:"fast_mode_state"`
	// Uuid is a unique identifier for this result entry.
	Uuid string `json:"uuid"`
}

// Failure returns an error if the Claude execution resulted in a failure.
// If the ClaudeResult is nil or there was no error, it returns nil.
func (r *ClaudeResult) Failure() error {
	if r == nil {
		return nil
	}
	if r.IsError {
		return errors.New(r.RawResult)
	}
	return nil
}

// Result returns the raw output string from the Claude model.
// If the ClaudeResult is nil, it returns an empty string.
func (r *ClaudeResult) Result() string {
	if r == nil {
		return ""
	}
	return r.RawResult
}

// PersistArtifact writes the Claude result content to a file in the artifact directory.
// It returns the file path of the persisted artifact and any filesystem error encountered.
// If the ClaudeResult is nil, it returns an empty string and a nil error.
func (r *ClaudeResult) PersistArtifact(ctx context.Context, artifactName string) (string, error) {
	if r == nil {
		return "", nil
	}
	return PersistArtifactToFile(ctx, artifactName, r.RawResult)
}
