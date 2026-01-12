package config

import (
	"encoding/json"
	"testing"
)

// TestProviderConfigJSONContract ensures Go structs serialize to snake_case JSON
// that matches TypeScript frontend expectations.
//
// This test prevents the common bug where Go's default JSON serialization uses
// PascalCase field names, but the TypeScript frontend expects snake_case.
func TestProviderConfigJSONContract(t *testing.T) {
	cfg := &ProviderConfig{
		Format:           "anthropic",
		BaseURL:          "https://api.anthropic.com",
		Version:          "2023-06-01",
		MaxRetries:       3,
		FallbackProvider: "openai",
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// These snake_case fields must be present (TypeScript expects them)
	requiredFields := []string{"format", "base_url", "max_retries", "circuit_breaker"}
	for _, field := range requiredFields {
		if _, ok := result[field]; !ok {
			t.Errorf("Required snake_case field '%s' not found. Got: %v", field, keysOf(result))
		}
	}

	// Optional fields should use snake_case when present
	// (version, fallback_provider, api_key - handled by struct tags with omitempty)

	// PascalCase fields must NOT be present (would indicate missing JSON tags)
	forbiddenFields := []string{"Format", "BaseURL", "Version", "MaxRetries", "CircuitBreaker", "FallbackProvider", "APIKey"}
	for _, field := range forbiddenFields {
		if _, ok := result[field]; ok {
			t.Errorf("PascalCase field '%s' found - JSON tag missing! TypeScript will not be able to read this field", field)
		}
	}

	t.Logf("✓ ProviderConfig serializes correctly: %s", string(data))
}

// TestCircuitBreakerConfigJSONContract ensures CircuitBreakerConfig uses snake_case
func TestCircuitBreakerConfigJSONContract(t *testing.T) {
	cfg := &CircuitBreakerConfig{
		Enabled:     true,
		MaxFailures: 5,
		Timeout:     "30s",
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Required snake_case fields
	requiredFields := []string{"enabled", "max_failures"}
	for _, field := range requiredFields {
		if _, ok := result[field]; !ok {
			t.Errorf("Required snake_case field '%s' not found. Got: %v", field, keysOf(result))
		}
	}

	// TimeoutDuration should be excluded from JSON (internal field)
	if _, ok := result["TimeoutDuration"]; ok {
		t.Error("Internal field 'TimeoutDuration' should not appear in JSON")
	}
	if _, ok := result["timeout_duration"]; ok {
		t.Error("Internal field 'timeout_duration' should not appear in JSON")
	}

	// PascalCase must not appear
	forbiddenFields := []string{"Enabled", "MaxFailures", "Timeout"}
	for _, field := range forbiddenFields {
		if _, ok := result[field]; ok {
			t.Errorf("PascalCase field '%s' found - JSON tag missing!", field)
		}
	}

	t.Logf("✓ CircuitBreakerConfig serializes correctly: %s", string(data))
}

// TestSubagentsConfigJSONContract ensures SubagentsConfig uses snake_case
func TestSubagentsConfigJSONContract(t *testing.T) {
	cfg := &SubagentsConfig{
		Enable:   true,
		Mappings: map[string]string{"code-reviewer": "openai:gpt-4o"},
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Required snake_case fields
	if _, ok := result["enable"]; !ok {
		t.Errorf("Required field 'enable' not found")
	}
	if _, ok := result["mappings"]; !ok {
		t.Errorf("Required field 'mappings' not found")
	}

	// PascalCase must not appear
	if _, ok := result["Enable"]; ok {
		t.Error("PascalCase field 'Enable' found - JSON tag missing!")
	}
	if _, ok := result["Mappings"]; ok {
		t.Error("PascalCase field 'Mappings' found - JSON tag missing!")
	}

	t.Logf("✓ SubagentsConfig serializes correctly: %s", string(data))
}

// TestRoutingConfigJSONContract ensures RoutingConfig uses snake_case
func TestRoutingConfigJSONContract(t *testing.T) {
	cfg := &RoutingConfig{
		Preferences: PreferencesConfig{Default: "balanced"},
		Tasks:       map[string]TaskRoutingConfig{"code-review": {Preference: "quality", Providers: []string{"anthropic"}}},
		ProviderProfiles: map[string]ProviderProfileConfig{
			"anthropic": {Speed: 7, Cost: 8, Quality: 9},
		},
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Required snake_case fields
	requiredFields := []string{"preferences", "tasks", "provider_profiles"}
	for _, field := range requiredFields {
		if _, ok := result[field]; !ok {
			t.Errorf("Required snake_case field '%s' not found. Got: %v", field, keysOf(result))
		}
	}

	// PascalCase must not appear
	forbiddenFields := []string{"Preferences", "Tasks", "ProviderProfiles"}
	for _, field := range forbiddenFields {
		if _, ok := result[field]; ok {
			t.Errorf("PascalCase field '%s' found - JSON tag missing!", field)
		}
	}

	t.Logf("✓ RoutingConfig serializes correctly: %s", string(data))
}

func keysOf(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
