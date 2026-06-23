package utils

import "testing"

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain name",
			input:    "simpleName",
			expected: "simplename",
		},
		{
			name:     "spaces and hyphens",
			input:    "My - Agent - 123",
			expected: "my_agent_123",
		},
		{
			name:     "consecutive delimiters",
			input:    "///some--name\\\\   with spaces",
			expected: "some_name_with_spaces",
		},
		{
			name:     "leading and trailing",
			input:    "  -foo-bar-  ",
			expected: "foo_bar",
		},
		{
			name:     "complex pattern",
			input:    "agent/test\\run-1",
			expected: "agent_test_run_1",
		},
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
		{
			name:     "only delimiters",
			input:    "---///\\\\\\   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeName(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizeName(%q) got %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
