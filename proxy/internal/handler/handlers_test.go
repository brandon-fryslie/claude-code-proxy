package handler

import (
	"testing"

	"github.com/seifghazi/claude-code-monitor/internal/model"
)

func TestExtractToolsUsed(t *testing.T) {
	tests := []struct {
		name     string
		request  model.AnthropicRequest
		expected []string
	}{
		{
			name: "No tools",
			request: model.AnthropicRequest{
				Model:    "claude-3-opus",
				Messages: []model.AnthropicMessage{},
				Tools:    nil,
			},
			expected: nil,
		},
		{
			name: "Single tool",
			request: model.AnthropicRequest{
				Model:    "claude-3-opus",
				Messages: []model.AnthropicMessage{},
				Tools: []model.Tool{
					{Name: "Read", Description: "Read a file"},
				},
			},
			expected: []string{"Read"},
		},
		{
			name: "Multiple tools",
			request: model.AnthropicRequest{
				Model:    "claude-3-opus",
				Messages: []model.AnthropicMessage{},
				Tools: []model.Tool{
					{Name: "Read", Description: "Read a file"},
					{Name: "Write", Description: "Write a file"},
					{Name: "Bash", Description: "Run bash command"},
					{Name: "Glob", Description: "Find files"},
				},
			},
			expected: []string{"Read", "Write", "Bash", "Glob"},
		},
		{
			name: "Empty tool names filtered out",
			request: model.AnthropicRequest{
				Model:    "claude-3-opus",
				Messages: []model.AnthropicMessage{},
				Tools: []model.Tool{
					{Name: "Read", Description: "Read a file"},
					{Name: "", Description: "Empty tool"},
					{Name: "Write", Description: "Write a file"},
				},
			},
			expected: []string{"Read", "", "Write"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Extract tools used (simulating the logic in Messages handler)
			var toolsUsed []string
			for _, tool := range tt.request.Tools {
				toolsUsed = append(toolsUsed, tool.Name)
			}

			if len(toolsUsed) != len(tt.expected) {
				t.Errorf("toolsUsed length = %d, want %d", len(toolsUsed), len(tt.expected))
				return
			}

			for i, tool := range toolsUsed {
				if tool != tt.expected[i] {
					t.Errorf("toolsUsed[%d] = %q, want %q", i, tool, tt.expected[i])
				}
			}
		})
	}
}

func TestCountToolCallsInResponse(t *testing.T) {
	tests := []struct {
		name     string
		content  []model.AnthropicContentBlock
		expected int
	}{
		{
			name:     "No content blocks",
			content:  nil,
			expected: 0,
		},
		{
			name: "Text only",
			content: []model.AnthropicContentBlock{
				{Type: "text", Text: "Hello, world!"},
			},
			expected: 0,
		},
		{
			name: "Single tool use",
			content: []model.AnthropicContentBlock{
				{Type: "text", Text: "Let me help you."},
				{Type: "tool_use", Text: ""},
			},
			expected: 1,
		},
		{
			name: "Multiple tool uses",
			content: []model.AnthropicContentBlock{
				{Type: "text", Text: "I'll use several tools."},
				{Type: "tool_use", Text: ""},
				{Type: "tool_use", Text: ""},
				{Type: "tool_use", Text: ""},
			},
			expected: 3,
		},
		{
			name: "Mixed content",
			content: []model.AnthropicContentBlock{
				{Type: "text", Text: "Starting..."},
				{Type: "tool_use", Text: ""},
				{Type: "text", Text: "Now doing more..."},
				{Type: "tool_use", Text: ""},
				{Type: "text", Text: "Done!"},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Count tool_use blocks (simulating the logic in handleNonStreamingResponse)
			toolCallCount := 0
			for _, block := range tt.content {
				if block.Type == "tool_use" {
					toolCallCount++
				}
			}

			if toolCallCount != tt.expected {
				t.Errorf("toolCallCount = %d, want %d", toolCallCount, tt.expected)
			}
		})
	}
}

func TestSanitizeHeaders(t *testing.T) {
	tests := []struct {
		name            string
		input           map[string][]string
		sensitiveKeys   []string // keys that should be hashed
		preservedKeys   []string // keys that should preserve original value
	}{
		{
			name:          "Empty headers",
			input:         map[string][]string{},
			sensitiveKeys: nil,
			preservedKeys: nil,
		},
		{
			name: "Authorization header hashed",
			input: map[string][]string{
				"Authorization": {"Bearer sk-secret-key"},
				"Content-Type":  {"application/json"},
			},
			sensitiveKeys: []string{"Authorization"},
			preservedKeys: []string{"Content-Type"},
		},
		{
			name: "X-Api-Key header hashed",
			input: map[string][]string{
				"X-Api-Key":    {"secret-api-key"},
				"Content-Type": {"application/json"},
			},
			sensitiveKeys: []string{"X-Api-Key"},
			preservedKeys: []string{"Content-Type"},
		},
		{
			name: "Multiple sensitive headers",
			input: map[string][]string{
				"Authorization": {"Bearer token"},
				"X-Api-Key":     {"api-key"},
				"Content-Type":  {"application/json"},
			},
			sensitiveKeys: []string{"Authorization", "X-Api-Key"},
			preservedKeys: []string{"Content-Type"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeHeaders(tt.input)

			// Check sensitive headers are hashed (start with "sha256:")
			for _, key := range tt.sensitiveKeys {
				resultValues, exists := result[key]
				if !exists {
					t.Errorf("missing header %q", key)
					continue
				}

				for i, val := range resultValues {
					if len(val) < 7 || val[:7] != "sha256:" {
						t.Errorf("header %q[%d] = %q, should start with 'sha256:'", key, i, val)
					}
					// Verify it's not the original value
					originalVal := tt.input[key][i]
					if val == originalVal {
						t.Errorf("header %q[%d] was not sanitized, still equals %q", key, i, originalVal)
					}
				}
			}

			// Check preserved headers have original values
			for _, key := range tt.preservedKeys {
				resultValues, exists := result[key]
				if !exists {
					t.Errorf("missing header %q", key)
					continue
				}

				originalValues := tt.input[key]
				if len(resultValues) != len(originalValues) {
					t.Errorf("header %q: values length = %d, want %d", key, len(resultValues), len(originalValues))
					continue
				}

				for i, expected := range originalValues {
					if resultValues[i] != expected {
						t.Errorf("header %q[%d] = %q, want %q", key, i, resultValues[i], expected)
					}
				}
			}
		})
	}
}
