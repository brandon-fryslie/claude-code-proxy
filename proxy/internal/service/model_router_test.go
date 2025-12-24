package service

import (
	"context"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/seifghazi/claude-code-monitor/internal/config"
	"github.com/seifghazi/claude-code-monitor/internal/model"
	"github.com/seifghazi/claude-code-monitor/internal/provider"
)

func TestModelRouter_EdgeCases(t *testing.T) {
	// Setup with new provider format
	cfg := &config.Config{
		Providers: map[string]*config.ProviderConfig{
			"anthropic": {
				Format:  "anthropic",
				BaseURL: "https://api.anthropic.com",
			},
			"openai": {
				Format:  "openai",
				BaseURL: "https://api.openai.com",
			},
		},
		Subagents: config.SubagentsConfig{
			Mappings: map[string]string{
				// New format: provider:model
				"streaming-systems-engineer": "openai:gpt-4o",
			},
		},
	}

	providers := make(map[string]provider.Provider)
	providers["anthropic"] = nil
	providers["openai"] = nil

	logger := log.New(os.Stdout, "test: ", log.LstdFlags)
	router := NewModelRouter(cfg, providers, logger)

	tests := []struct {
		name          string
		request       *model.AnthropicRequest
		expectedRoute string
		expectedModel string
		description   string
	}{
		{
			name: "Regular Claude Code request (no Notes section)",
			request: &model.AnthropicRequest{
				Model: "claude-3-opus-20240229",
				System: []model.AnthropicSystemMessage{
					{Text: "You are Claude Code, Anthropic's official CLI for Claude."},
					{Text: "You are an interactive CLI tool that helps users with software engineering tasks. Use the instructions below and the tools available to you to assist the user."},
				},
			},
			expectedRoute: "anthropic",
			expectedModel: "claude-3-opus-20240229",
			description:   "Regular Claude Code requests should use original model",
		},
		{
			name: "Non-Claude Code request",
			request: &model.AnthropicRequest{
				Model: "claude-3-opus-20240229",
				System: []model.AnthropicSystemMessage{
					{Text: "You are a helpful assistant."},
				},
			},
			expectedRoute: "anthropic",
			expectedModel: "claude-3-opus-20240229",
			description:   "Non-Claude Code requests should use original model",
		},
		{
			name: "Single system message",
			request: &model.AnthropicRequest{
				Model:  "claude-3-opus-20240229",
				System: []model.AnthropicSystemMessage{},
			},
			expectedRoute: "anthropic",
			expectedModel: "claude-3-opus-20240229",
			description:   "Requests with no system messages should use original model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.request.System) == 2 {
				// Test extract static prompt for second message
				fullPrompt := tt.request.System[1].Text
				staticPrompt := router.extractStaticPrompt(fullPrompt)

				// Verify no "Notes:" in static prompt
				if contains(staticPrompt, "Notes:") {
					t.Errorf("Static prompt should not contain 'Notes:' section")
				}
			}

			// Log for manual verification
			t.Logf("Test case: %s", tt.description)
		})
	}
}

func TestModelRouter_ExtractStaticPrompt(t *testing.T) {
	router := &ModelRouter{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Prompt with Notes section",
			input:    "You are an expert engineer.\n\nNotes:\n- Some dynamic content\n- More notes",
			expected: "You are an expert engineer.",
		},
		{
			name:     "Prompt without Notes section",
			input:    "You are an expert engineer.\nNo notes here.",
			expected: "You are an expert engineer.\nNo notes here.",
		},
		{
			name:     "Prompt with double newline before Notes",
			input:    "You are an expert.\n\nNotes:\nDynamic content",
			expected: "You are an expert.",
		},
		{
			name:     "Empty prompt",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := router.extractStaticPrompt(tt.input)
			if result != tt.expected {
				t.Errorf("extractStaticPrompt() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestModelRouter_ParseMappings(t *testing.T) {
	// Test that mappings are correctly parsed from provider:model format
	cfg := &config.Config{
		Providers: map[string]*config.ProviderConfig{
			"anthropic": {
				Format:  "anthropic",
				BaseURL: "https://api.anthropic.com",
			},
			"openai": {
				Format:  "openai",
				BaseURL: "https://api.openai.com",
			},
			"localllm": {
				Format:  "openai",
				BaseURL: "http://localhost:1234",
			},
		},
		Subagents: config.SubagentsConfig{
			Enable: true,
			Mappings: map[string]string{
				"code-reviewer": "openai:gpt-4o",
				"planner":       "localllm:my-local-model",
				"invalid":       "just-a-model", // Invalid format - missing provider
			},
		},
	}

	providers := make(map[string]provider.Provider)
	providers["anthropic"] = nil
	providers["openai"] = nil
	providers["localllm"] = nil

	logger := log.New(os.Stdout, "test: ", log.LstdFlags)
	router := NewModelRouter(cfg, providers, logger)

	// Verify parsed mappings
	if len(router.subagentMappings) != 2 {
		t.Errorf("Expected 2 valid mappings, got %d", len(router.subagentMappings))
	}

	// Check code-reviewer mapping
	if mapping, exists := router.subagentMappings["code-reviewer"]; exists {
		if mapping.ProviderName != "openai" {
			t.Errorf("Expected provider 'openai', got '%s'", mapping.ProviderName)
		}
		if mapping.ModelName != "gpt-4o" {
			t.Errorf("Expected model 'gpt-4o', got '%s'", mapping.ModelName)
		}
	} else {
		t.Error("code-reviewer mapping not found")
	}

	// Check planner mapping
	if mapping, exists := router.subagentMappings["planner"]; exists {
		if mapping.ProviderName != "localllm" {
			t.Errorf("Expected provider 'localllm', got '%s'", mapping.ProviderName)
		}
		if mapping.ModelName != "my-local-model" {
			t.Errorf("Expected model 'my-local-model', got '%s'", mapping.ModelName)
		}
	} else {
		t.Error("planner mapping not found")
	}

	// Verify invalid mapping was skipped
	if _, exists := router.subagentMappings["invalid"]; exists {
		t.Error("Invalid mapping should have been skipped")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && s[0:len(substr)] == substr) ||
		(len(s) > len(substr) && contains(s[1:], substr)))
}

// TestRoutingDecision_ProviderNameAndSubagentName tests that ProviderName and SubagentName
// are correctly populated in the RoutingDecision
func TestRoutingDecision_ProviderNameAndSubagentName(t *testing.T) {
	// Create a mock provider that implements the Provider interface
	mockAnthropicProvider := &mockProvider{name: "anthropic"}
	mockOpenAIProvider := &mockProvider{name: "openai"}

	cfg := &config.Config{
		Providers: map[string]*config.ProviderConfig{
			"anthropic": {
				Format:  "anthropic",
				BaseURL: "https://api.anthropic.com",
			},
			"openai": {
				Format:  "openai",
				BaseURL: "https://api.openai.com",
			},
		},
		Subagents: config.SubagentsConfig{
			Enable: true,
			Mappings: map[string]string{
				"test-agent": "openai:gpt-4o",
			},
		},
	}

	providers := make(map[string]provider.Provider)
	providers["anthropic"] = mockAnthropicProvider
	providers["openai"] = mockOpenAIProvider

	logger := log.New(os.Stdout, "test: ", log.LstdFlags)
	router := NewModelRouter(cfg, providers, logger)

	// Manually add a test agent prompt hash for testing
	testPrompt := "You are a test agent for unit testing."
	hash := router.hashString(testPrompt)
	router.customAgentPrompts[hash] = SubagentDefinition{
		Name:           "test-agent",
		TargetModel:    "gpt-4o",
		TargetProvider: "openai",
		FullPrompt:     testPrompt,
	}

	tests := []struct {
		name                 string
		request              *model.AnthropicRequest
		expectedProviderName string
		expectedSubagentName string
		expectedTargetModel  string
	}{
		{
			name: "Default route populates ProviderName",
			request: &model.AnthropicRequest{
				Model: "claude-3-opus-20240229",
				System: []model.AnthropicSystemMessage{
					{Text: "You are a helpful assistant."},
				},
			},
			expectedProviderName: "anthropic",
			expectedSubagentName: "",
			expectedTargetModel:  "claude-3-opus-20240229",
		},
		{
			name: "Subagent route populates both ProviderName and SubagentName",
			request: &model.AnthropicRequest{
				Model: "claude-3-opus-20240229",
				System: []model.AnthropicSystemMessage{
					{Text: "You are Claude Code, Anthropic's official CLI for Claude."},
					{Text: testPrompt},
				},
			},
			expectedProviderName: "openai",
			expectedSubagentName: "test-agent",
			expectedTargetModel:  "gpt-4o",
		},
		{
			name: "OpenAI model routes to OpenAI provider",
			request: &model.AnthropicRequest{
				Model:  "gpt-4o",
				System: []model.AnthropicSystemMessage{},
			},
			expectedProviderName: "openai",
			expectedSubagentName: "",
			expectedTargetModel:  "gpt-4o",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, err := router.DetermineRoute(tt.request)
			if err != nil {
				t.Fatalf("DetermineRoute() error = %v", err)
			}

			if decision.ProviderName != tt.expectedProviderName {
				t.Errorf("ProviderName = %q, want %q", decision.ProviderName, tt.expectedProviderName)
			}

			if decision.SubagentName != tt.expectedSubagentName {
				t.Errorf("SubagentName = %q, want %q", decision.SubagentName, tt.expectedSubagentName)
			}

			if decision.TargetModel != tt.expectedTargetModel {
				t.Errorf("TargetModel = %q, want %q", decision.TargetModel, tt.expectedTargetModel)
			}
		})
	}
}

// mockProvider implements provider.Provider for testing
type mockProvider struct {
	name string
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) ForwardRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	return nil, nil
}
