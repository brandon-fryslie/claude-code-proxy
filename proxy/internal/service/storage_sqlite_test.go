package service

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/seifghazi/claude-code-monitor/internal/config"
	"github.com/seifghazi/claude-code-monitor/internal/model"
)

func setupTestDB(t *testing.T) (StorageService, func()) {
	// Create a temporary database file
	tmpFile, err := os.CreateTemp("", "test_requests_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	cfg := &config.StorageConfig{
		DBPath: tmpFile.Name(),
	}

	storage, err := NewSQLiteStorageService(cfg)
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to create storage service: %v", err)
	}

	cleanup := func() {
		storage.Close()
		os.Remove(tmpFile.Name())
	}

	return storage, cleanup
}

func TestSaveRequest_NewFields(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a request with all new fields populated
	request := &model.RequestLog{
		RequestID:     "test-123",
		Timestamp:     "2024-01-15T10:30:00Z",
		Method:        "POST",
		Endpoint:      "/v1/messages",
		Headers:       map[string][]string{"Content-Type": {"application/json"}},
		Body:          map[string]interface{}{"model": "claude-3-opus"},
		Model:         "claude-3-opus",
		OriginalModel: "claude-3-opus",
		RoutedModel:   "gpt-4o",
		Provider:      "openai",
		SubagentName:  "code-reviewer",
		ToolsUsed:     []string{"Read", "Write", "Bash"},
		ToolCallCount: 0,
		UserAgent:     "test-agent",
		ContentType:   "application/json",
	}

	// Save the request
	id, err := storage.SaveRequest(request)
	if err != nil {
		t.Fatalf("SaveRequest() error = %v", err)
	}

	if id != "test-123" {
		t.Errorf("SaveRequest() returned id = %q, want %q", id, "test-123")
	}

	// Retrieve and verify
	summaries, err := storage.GetRequestsSummary("")
	if err != nil {
		t.Fatalf("GetRequestsSummary() error = %v", err)
	}

	if len(summaries) != 1 {
		t.Fatalf("Expected 1 summary, got %d", len(summaries))
	}

	// Note: GetRequestsSummary doesn't return all fields, so we just verify it doesn't error
}

func TestUpdateRequestWithResponse_TokensAndTiming(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()

	// First save a request
	request := &model.RequestLog{
		RequestID:     "test-456",
		Timestamp:     "2024-01-15T10:30:00Z",
		Method:        "POST",
		Endpoint:      "/v1/messages",
		Headers:       map[string][]string{},
		Body:          map[string]interface{}{},
		Model:         "claude-3-opus",
		OriginalModel: "claude-3-opus",
		RoutedModel:   "claude-3-opus",
		Provider:      "anthropic",
		UserAgent:     "test-agent",
		ContentType:   "application/json",
	}

	_, err := storage.SaveRequest(request)
	if err != nil {
		t.Fatalf("SaveRequest() error = %v", err)
	}

	// Create a response with usage data
	usage := &model.AnthropicUsage{
		InputTokens:              1000,
		OutputTokens:             500,
		CacheReadInputTokens:     200,
		CacheCreationInputTokens: 100,
	}

	responseBody := map[string]interface{}{
		"usage": usage,
		"content": []map[string]interface{}{
			{"type": "text", "text": "Hello"},
		},
	}
	responseBodyBytes, _ := json.Marshal(responseBody)

	request.Response = &model.ResponseLog{
		StatusCode:    200,
		Headers:       map[string][]string{},
		Body:          json.RawMessage(responseBodyBytes),
		ResponseTime:  1500,
		FirstByteTime: 200,
		IsStreaming:   false,
		CompletedAt:   "2024-01-15T10:30:01Z",
		ToolCallCount: 3,
	}

	// Update with response
	err = storage.UpdateRequestWithResponse(request)
	if err != nil {
		t.Fatalf("UpdateRequestWithResponse() error = %v", err)
	}

	// Verify the data was saved by getting summaries
	summaries, err := storage.GetRequestsSummary("")
	if err != nil {
		t.Fatalf("GetRequestsSummary() error = %v", err)
	}

	if len(summaries) != 1 {
		t.Fatalf("Expected 1 summary, got %d", len(summaries))
	}

	summary := summaries[0]
	if summary.ResponseTime != 1500 {
		t.Errorf("ResponseTime = %d, want 1500", summary.ResponseTime)
	}

	if summary.Usage == nil {
		t.Error("Usage should not be nil")
	} else {
		if summary.Usage.InputTokens != 1000 {
			t.Errorf("InputTokens = %d, want 1000", summary.Usage.InputTokens)
		}
		if summary.Usage.OutputTokens != 500 {
			t.Errorf("OutputTokens = %d, want 500", summary.Usage.OutputTokens)
		}
	}
}

func TestMigration_ExistingDatabase(t *testing.T) {
	// Create a temporary database file
	tmpFile, err := os.CreateTemp("", "test_migration_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	cfg := &config.StorageConfig{
		DBPath: tmpFile.Name(),
	}

	// First, create storage - this will create the table
	storage1, err := NewSQLiteStorageService(cfg)
	if err != nil {
		t.Fatalf("Failed to create first storage service: %v", err)
	}

	// Save a request
	request := &model.RequestLog{
		RequestID:     "migration-test",
		Timestamp:     "2024-01-15T10:30:00Z",
		Method:        "POST",
		Endpoint:      "/v1/messages",
		Headers:       map[string][]string{},
		Body:          map[string]interface{}{},
		Model:         "claude-3-opus",
		OriginalModel: "claude-3-opus",
		RoutedModel:   "claude-3-opus",
		Provider:      "anthropic",
		SubagentName:  "test-agent",
		ToolsUsed:     []string{"Read"},
		UserAgent:     "test",
		ContentType:   "application/json",
	}

	_, err = storage1.SaveRequest(request)
	if err != nil {
		t.Fatalf("SaveRequest() error = %v", err)
	}

	storage1.Close()

	// Now reopen the database - migrations should run again without error
	storage2, err := NewSQLiteStorageService(cfg)
	if err != nil {
		t.Fatalf("Failed to create second storage service: %v", err)
	}
	defer storage2.Close()

	// Verify data still exists
	summaries, err := storage2.GetRequestsSummary("")
	if err != nil {
		t.Fatalf("GetRequestsSummary() error = %v", err)
	}

	if len(summaries) != 1 {
		t.Fatalf("Expected 1 summary after migration, got %d", len(summaries))
	}

	if summaries[0].RequestID != "migration-test" {
		t.Errorf("RequestID = %q, want %q", summaries[0].RequestID, "migration-test")
	}
}

func TestGetStats_WithProviderData(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()

	// Create requests with different providers
	requests := []*model.RequestLog{
		{
			RequestID:     "req-1",
			Timestamp:     "2024-01-15T10:00:00Z",
			Method:        "POST",
			Endpoint:      "/v1/messages",
			Headers:       map[string][]string{},
			Body:          map[string]interface{}{},
			Model:         "claude-3-opus",
			Provider:      "anthropic",
			SubagentName:  "",
			UserAgent:     "test",
			ContentType:   "application/json",
		},
		{
			RequestID:     "req-2",
			Timestamp:     "2024-01-15T11:00:00Z",
			Method:        "POST",
			Endpoint:      "/v1/messages",
			Headers:       map[string][]string{},
			Body:          map[string]interface{}{},
			Model:         "gpt-4o",
			Provider:      "openai",
			SubagentName:  "code-reviewer",
			UserAgent:     "test",
			ContentType:   "application/json",
		},
	}

	for _, req := range requests {
		_, err := storage.SaveRequest(req)
		if err != nil {
			t.Fatalf("SaveRequest() error = %v", err)
		}

		// Add a response with usage
		usage := map[string]interface{}{
			"usage": map[string]interface{}{
				"input_tokens":  100,
				"output_tokens": 50,
			},
		}
		usageBytes, _ := json.Marshal(usage)

		req.Response = &model.ResponseLog{
			StatusCode:   200,
			Headers:      map[string][]string{},
			Body:         json.RawMessage(usageBytes),
			ResponseTime: 1000,
			IsStreaming:  false,
			CompletedAt:  "2024-01-15T10:00:01Z",
		}
		storage.UpdateRequestWithResponse(req)
	}

	// Get stats
	stats, err := storage.GetStats("2024-01-15", "2024-01-16")
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}

	if len(stats.DailyStats) != 1 {
		t.Errorf("Expected 1 day of stats, got %d", len(stats.DailyStats))
	}

	if stats.DailyStats[0].Requests != 2 {
		t.Errorf("Expected 2 requests, got %d", stats.DailyStats[0].Requests)
	}
}

func TestGetProviderStats(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()

	// Create requests with different providers
	requests := []*model.RequestLog{
		{
			RequestID:    "prov-1",
			Timestamp:    "2024-01-15T10:00:00Z",
			Method:       "POST",
			Endpoint:     "/v1/messages",
			Headers:      map[string][]string{},
			Body:         map[string]interface{}{},
			Model:        "claude-3-opus",
			Provider:     "anthropic",
			SubagentName: "",
			UserAgent:    "test",
			ContentType:  "application/json",
		},
		{
			RequestID:    "prov-2",
			Timestamp:    "2024-01-15T11:00:00Z",
			Method:       "POST",
			Endpoint:     "/v1/messages",
			Headers:      map[string][]string{},
			Body:         map[string]interface{}{},
			Model:        "gpt-4o",
			Provider:     "openai",
			SubagentName: "code-reviewer",
			UserAgent:    "test",
			ContentType:  "application/json",
		},
		{
			RequestID:    "prov-3",
			Timestamp:    "2024-01-15T12:00:00Z",
			Method:       "POST",
			Endpoint:     "/v1/messages",
			Headers:      map[string][]string{},
			Body:         map[string]interface{}{},
			Model:        "claude-3-sonnet",
			Provider:     "anthropic",
			SubagentName: "",
			UserAgent:    "test",
			ContentType:  "application/json",
		},
	}

	for _, req := range requests {
		_, err := storage.SaveRequest(req)
		if err != nil {
			t.Fatalf("SaveRequest() error = %v", err)
		}

		// Add response with tokens
		usage := map[string]interface{}{
			"usage": map[string]interface{}{
				"input_tokens":  100,
				"output_tokens": 50,
			},
		}
		usageBytes, _ := json.Marshal(usage)

		req.Response = &model.ResponseLog{
			StatusCode:   200,
			Headers:      map[string][]string{},
			Body:         json.RawMessage(usageBytes),
			ResponseTime: 1000,
			IsStreaming:  false,
			CompletedAt:  "2024-01-15T10:00:01Z",
		}
		storage.UpdateRequestWithResponse(req)
	}

	stats, err := storage.GetProviderStats("2024-01-15T00:00:00Z", "2024-01-16T00:00:00Z")
	if err != nil {
		t.Fatalf("GetProviderStats() error = %v", err)
	}

	if len(stats.Providers) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(stats.Providers))
	}

	// Find anthropic provider
	var anthropicStats *model.ProviderStats
	for i := range stats.Providers {
		if stats.Providers[i].Provider == "anthropic" {
			anthropicStats = &stats.Providers[i]
			break
		}
	}

	if anthropicStats == nil {
		t.Error("Expected to find anthropic provider stats")
	} else {
		if anthropicStats.Requests != 2 {
			t.Errorf("Anthropic requests = %d, want 2", anthropicStats.Requests)
		}
		if anthropicStats.InputTokens != 200 {
			t.Errorf("Anthropic input tokens = %d, want 200", anthropicStats.InputTokens)
		}
	}
}

func TestGetSubagentStats(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()

	// Create requests with different subagents
	requests := []*model.RequestLog{
		{
			RequestID:     "sub-1",
			Timestamp:     "2024-01-15T10:00:00Z",
			Method:        "POST",
			Endpoint:      "/v1/messages",
			Headers:       map[string][]string{},
			Body:          map[string]interface{}{},
			Model:         "claude-3-opus",
			OriginalModel: "claude-3-opus",
			RoutedModel:   "gpt-4o",
			Provider:      "openai",
			SubagentName:  "code-reviewer",
			UserAgent:     "test",
			ContentType:   "application/json",
		},
		{
			RequestID:     "sub-2",
			Timestamp:     "2024-01-15T11:00:00Z",
			Method:        "POST",
			Endpoint:      "/v1/messages",
			Headers:       map[string][]string{},
			Body:          map[string]interface{}{},
			Model:         "claude-3-opus",
			OriginalModel: "claude-3-opus",
			RoutedModel:   "gpt-4o-mini",
			Provider:      "openai",
			SubagentName:  "test-runner",
			UserAgent:     "test",
			ContentType:   "application/json",
		},
		{
			RequestID:     "sub-3",
			Timestamp:     "2024-01-15T12:00:00Z",
			Method:        "POST",
			Endpoint:      "/v1/messages",
			Headers:       map[string][]string{},
			Body:          map[string]interface{}{},
			Model:         "claude-3-opus",
			OriginalModel: "claude-3-opus",
			RoutedModel:   "gpt-4o",
			Provider:      "openai",
			SubagentName:  "code-reviewer",
			UserAgent:     "test",
			ContentType:   "application/json",
		},
	}

	for _, req := range requests {
		_, err := storage.SaveRequest(req)
		if err != nil {
			t.Fatalf("SaveRequest() error = %v", err)
		}

		usage := map[string]interface{}{
			"usage": map[string]interface{}{
				"input_tokens":  100,
				"output_tokens": 50,
			},
		}
		usageBytes, _ := json.Marshal(usage)

		req.Response = &model.ResponseLog{
			StatusCode:   200,
			Headers:      map[string][]string{},
			Body:         json.RawMessage(usageBytes),
			ResponseTime: 500,
			IsStreaming:  false,
			CompletedAt:  "2024-01-15T10:00:01Z",
		}
		storage.UpdateRequestWithResponse(req)
	}

	stats, err := storage.GetSubagentStats("2024-01-15T00:00:00Z", "2024-01-16T00:00:00Z")
	if err != nil {
		t.Fatalf("GetSubagentStats() error = %v", err)
	}

	if len(stats.Subagents) != 2 {
		t.Errorf("Expected 2 subagents, got %d", len(stats.Subagents))
	}

	// Find code-reviewer subagent
	var codeReviewerStats *model.SubagentStats
	for i := range stats.Subagents {
		if stats.Subagents[i].SubagentName == "code-reviewer" {
			codeReviewerStats = &stats.Subagents[i]
			break
		}
	}

	if codeReviewerStats == nil {
		t.Error("Expected to find code-reviewer subagent stats")
	} else {
		if codeReviewerStats.Requests != 2 {
			t.Errorf("code-reviewer requests = %d, want 2", codeReviewerStats.Requests)
		}
		if codeReviewerStats.Provider != "openai" {
			t.Errorf("code-reviewer provider = %q, want openai", codeReviewerStats.Provider)
		}
		if codeReviewerStats.TargetModel != "gpt-4o" {
			t.Errorf("code-reviewer target model = %q, want gpt-4o", codeReviewerStats.TargetModel)
		}
	}
}

func TestGetToolStats(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()

	// Create requests with different tools
	requests := []*model.RequestLog{
		{
			RequestID:    "tool-1",
			Timestamp:    "2024-01-15T10:00:00Z",
			Method:       "POST",
			Endpoint:     "/v1/messages",
			Headers:      map[string][]string{},
			Body:         map[string]interface{}{},
			Model:        "claude-3-opus",
			Provider:     "anthropic",
			ToolsUsed:    []string{"Read", "Write", "Bash"},
			UserAgent:    "test",
			ContentType:  "application/json",
		},
		{
			RequestID:    "tool-2",
			Timestamp:    "2024-01-15T11:00:00Z",
			Method:       "POST",
			Endpoint:     "/v1/messages",
			Headers:      map[string][]string{},
			Body:         map[string]interface{}{},
			Model:        "claude-3-opus",
			Provider:     "anthropic",
			ToolsUsed:    []string{"Read", "Glob"},
			UserAgent:    "test",
			ContentType:  "application/json",
		},
	}

	for _, req := range requests {
		_, err := storage.SaveRequest(req)
		if err != nil {
			t.Fatalf("SaveRequest() error = %v", err)
		}

		// Add response with tool call count
		req.Response = &model.ResponseLog{
			StatusCode:    200,
			Headers:       map[string][]string{},
			Body:          json.RawMessage(`{}`),
			ResponseTime:  500,
			IsStreaming:   false,
			CompletedAt:   "2024-01-15T10:00:01Z",
			ToolCallCount: 2,
		}
		storage.UpdateRequestWithResponse(req)
	}

	stats, err := storage.GetToolStats("2024-01-15T00:00:00Z", "2024-01-16T00:00:00Z")
	if err != nil {
		t.Fatalf("GetToolStats() error = %v", err)
	}

	// Should have tools: Read (2), Write (1), Bash (1), Glob (1)
	if len(stats.Tools) < 4 {
		t.Errorf("Expected at least 4 tools, got %d", len(stats.Tools))
	}

	// Find Read tool stats
	var readStats *model.ToolStats
	for i := range stats.Tools {
		if stats.Tools[i].ToolName == "Read" {
			readStats = &stats.Tools[i]
			break
		}
	}

	if readStats == nil {
		t.Error("Expected to find Read tool stats")
	} else {
		if readStats.UsageCount != 2 {
			t.Errorf("Read usage count = %d, want 2", readStats.UsageCount)
		}
	}
}

func TestGetPerformanceStats(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()

	// Create requests with different response times
	requests := []*model.RequestLog{
		{
			RequestID:   "perf-1",
			Timestamp:   "2024-01-15T10:00:00Z",
			Method:      "POST",
			Endpoint:    "/v1/messages",
			Headers:     map[string][]string{},
			Body:        map[string]interface{}{},
			Model:       "claude-3-opus",
			Provider:    "anthropic",
			UserAgent:   "test",
			ContentType: "application/json",
		},
		{
			RequestID:   "perf-2",
			Timestamp:   "2024-01-15T11:00:00Z",
			Method:      "POST",
			Endpoint:    "/v1/messages",
			Headers:     map[string][]string{},
			Body:        map[string]interface{}{},
			Model:       "claude-3-opus",
			Provider:    "anthropic",
			UserAgent:   "test",
			ContentType: "application/json",
		},
		{
			RequestID:   "perf-3",
			Timestamp:   "2024-01-15T12:00:00Z",
			Method:      "POST",
			Endpoint:    "/v1/messages",
			Headers:     map[string][]string{},
			Body:        map[string]interface{}{},
			Model:       "gpt-4o",
			Provider:    "openai",
			UserAgent:   "test",
			ContentType: "application/json",
		},
	}

	responseTimes := []int64{500, 1000, 750}
	firstByteTimes := []int64{100, 200, 150}

	for i, req := range requests {
		_, err := storage.SaveRequest(req)
		if err != nil {
			t.Fatalf("SaveRequest() error = %v", err)
		}

		req.Response = &model.ResponseLog{
			StatusCode:    200,
			Headers:       map[string][]string{},
			Body:          json.RawMessage(`{}`),
			ResponseTime:  responseTimes[i],
			FirstByteTime: firstByteTimes[i],
			IsStreaming:   false,
			CompletedAt:   "2024-01-15T10:00:01Z",
		}
		storage.UpdateRequestWithResponse(req)
	}

	stats, err := storage.GetPerformanceStats("2024-01-15T00:00:00Z", "2024-01-16T00:00:00Z")
	if err != nil {
		t.Fatalf("GetPerformanceStats() error = %v", err)
	}

	// Should have stats for 2 provider/model combinations
	if len(stats.Stats) != 2 {
		t.Errorf("Expected 2 performance stat entries, got %d", len(stats.Stats))
	}

	// Find anthropic/claude-3-opus stats
	var opusStats *model.PerformanceStats
	for i := range stats.Stats {
		if stats.Stats[i].Provider == "anthropic" && stats.Stats[i].Model == "claude-3-opus" {
			opusStats = &stats.Stats[i]
			break
		}
	}

	if opusStats == nil {
		t.Error("Expected to find anthropic/claude-3-opus performance stats")
	} else {
		if opusStats.RequestCount != 2 {
			t.Errorf("Request count = %d, want 2", opusStats.RequestCount)
		}
		// Average of 500 and 1000 is 750
		if opusStats.AvgResponseMs != 750 {
			t.Errorf("Avg response time = %d, want 750", opusStats.AvgResponseMs)
		}
	}
}
