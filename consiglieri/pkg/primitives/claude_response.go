package primitives

import "errors"

type ClaudeResult struct {
	Type            string      `json:"type"`
	Subtype         string      `json:"subtype"`
	IsError         bool        `json:"is_error"`
	ApiErrorStatus  interface{} `json:"api_error_status"`
	DurationMs      int         `json:"duration_ms"`
	DurationApiMs   int         `json:"duration_api_ms"`
	TtftMs          int         `json:"ttft_ms"`
	TtftStreamMs    int         `json:"ttft_stream_ms"`
	TimeToRequestMs int         `json:"time_to_request_ms"`
	NumTurns        int         `json:"num_turns"`
	Result          string      `json:"result"`
	StopReason      string      `json:"stop_reason"`
	SessionId       string      `json:"session_id"`
	TotalCostUsd    float64     `json:"total_cost_usd"`
	Usage           struct {
		InputTokens              int `json:"input_tokens"`
		CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
		CacheReadInputTokens     int `json:"cache_read_input_tokens"`
		OutputTokens             int `json:"output_tokens"`
		ServerToolUse            struct {
			WebSearchRequests int `json:"web_search_requests"`
			WebFetchRequests  int `json:"web_fetch_requests"`
		} `json:"server_tool_use"`
		ServiceTier   string `json:"service_tier"`
		CacheCreation struct {
			Ephemeral1HInputTokens int `json:"ephemeral_1h_input_tokens"`
			Ephemeral5MInputTokens int `json:"ephemeral_5m_input_tokens"`
		} `json:"cache_creation"`
		InferenceGeo string        `json:"inference_geo"`
		Iterations   []interface{} `json:"iterations"`
		Speed        string        `json:"speed"`
	} `json:"usage"`
	ModelUsage struct {
		ClaudeSonnet46 struct {
			InputTokens              int     `json:"inputTokens"`
			OutputTokens             int     `json:"outputTokens"`
			CacheReadInputTokens     int     `json:"cacheReadInputTokens"`
			CacheCreationInputTokens int     `json:"cacheCreationInputTokens"`
			WebSearchRequests        int     `json:"webSearchRequests"`
			CostUSD                  float64 `json:"costUSD"`
			ContextWindow            int     `json:"contextWindow"`
			MaxOutputTokens          int     `json:"maxOutputTokens"`
		} `json:"claude-sonnet-4-6"`
	} `json:"modelUsage"`
	PermissionDenials []interface{} `json:"permission_denials"`
	TerminalReason    string        `json:"terminal_reason"`
	FastModeState     string        `json:"fast_mode_state"`
	Uuid              string        `json:"uuid"`
}

// TODO repensar sobre isso
func (r *ClaudeResult) Failure() error {
	if r == nil {
		return nil
	}
	if r.IsError {
		return errors.New(r.Result)
	}
	return nil
}
