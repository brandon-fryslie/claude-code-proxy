package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/seifghazi/claude-code-monitor/internal/config"
	"github.com/seifghazi/claude-code-monitor/internal/model"
	"github.com/seifghazi/claude-code-monitor/internal/service"
)

// setupTestDataHandler creates a DataHandler with in-memory SQLite for testing
func setupTestDataHandler(t *testing.T) (*DataHandler, *service.SQLiteStorageService, func()) {
	t.Helper()

	// Create temp directory for test database
	tmpDir, err := os.MkdirTemp("", "handler_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create test storage with in-memory database
	cfg := &config.StorageConfig{
		DBPath: filepath.Join(tmpDir, "test.db"),
	}

	storage, err := service.NewSQLiteStorageService(cfg)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create storage: %v", err)
	}

	sqliteStorage := storage.(*service.SQLiteStorageService)

	// Create handler
	appCfg := &config.Config{
		Storage: *cfg,
	}
	handler := NewDataHandler(storage, nil, appCfg)

	// Cleanup function
	cleanup := func() {
		sqliteStorage.Close()
		os.RemoveAll(tmpDir)
	}

	return handler, sqliteStorage, cleanup
}

func TestSearchConversations_EmptyQuery(t *testing.T) {
	handler, _, cleanup := setupTestDataHandler(t)
	defer cleanup()

	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Missing query parameter",
			url:            "/api/conversations/search",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Query parameter 'q' is required",
		},
		{
			name:           "Empty query parameter",
			url:            "/api/conversations/search?q=",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Query parameter 'q' is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			handler.SearchConversations(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}

			var result map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if result["error"] != tt.expectedError {
				t.Errorf("error = %q, want %q", result["error"], tt.expectedError)
			}
		})
	}
}

func TestSearchConversations_NoResults(t *testing.T) {
	handler, _, cleanup := setupTestDataHandler(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/conversations/search?q=xyzabc123nonexistent", nil)
	w := httptest.NewRecorder()

	handler.SearchConversations(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var result model.SearchResults
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result.Total != 0 {
		t.Errorf("total = %d, want 0", result.Total)
	}

	if len(result.Results) != 0 {
		t.Errorf("results count = %d, want 0", len(result.Results))
	}

	if result.Query != "xyzabc123nonexistent" {
		t.Errorf("query = %q, want 'xyzabc123nonexistent'", result.Query)
	}
}

func TestSearchConversations_Pagination(t *testing.T) {
	handler, _, cleanup := setupTestDataHandler(t)
	defer cleanup()

	tests := []struct {
		name           string
		query          string
		limit          string
		offset         string
		expectedLimit  int
		expectedOffset int
	}{
		{
			name:           "Default pagination",
			query:          "test",
			limit:          "",
			offset:         "",
			expectedLimit:  50,
			expectedOffset: 0,
		},
		{
			name:           "Custom limit",
			query:          "test",
			limit:          "10",
			offset:         "",
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name:           "Custom offset",
			query:          "test",
			limit:          "",
			offset:         "5",
			expectedLimit:  50,
			expectedOffset: 5,
		},
		{
			name:           "Custom limit and offset",
			query:          "test",
			limit:          "10",
			offset:         "5",
			expectedLimit:  10,
			expectedOffset: 5,
		},
		{
			name:           "Max limit enforced (200)",
			query:          "test",
			limit:          "999",
			offset:         "",
			expectedLimit:  50,
			expectedOffset: 0,
		},
		{
			name:           "Invalid limit defaults",
			query:          "test",
			limit:          "invalid",
			offset:         "",
			expectedLimit:  50,
			expectedOffset: 0,
		},
		{
			name:           "Negative limit defaults",
			query:          "test",
			limit:          "-5",
			offset:         "",
			expectedLimit:  50,
			expectedOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/conversations/search?q=" + tt.query
			if tt.limit != "" {
				url += "&limit=" + tt.limit
			}
			if tt.offset != "" {
				url += "&offset=" + tt.offset
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			handler.SearchConversations(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
			}

			var result model.SearchResults
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if result.Limit != tt.expectedLimit {
				t.Errorf("limit = %d, want %d", result.Limit, tt.expectedLimit)
			}

			if result.Offset != tt.expectedOffset {
				t.Errorf("offset = %d, want %d", result.Offset, tt.expectedOffset)
			}

			// Verify we don't return more results than the limit
			if len(result.Results) > tt.expectedLimit {
				t.Errorf("results count = %d, exceeds limit %d", len(result.Results), tt.expectedLimit)
			}
		})
	}
}

func TestSearchConversations_ResponseFormat(t *testing.T) {
	handler, _, cleanup := setupTestDataHandler(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/conversations/search?q=test", nil)
	w := httptest.NewRecorder()

	handler.SearchConversations(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var result model.SearchResults
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify response structure
	if result.Query != "test" {
		t.Errorf("Query field = %q, want 'test'", result.Query)
	}

	if result.Limit == 0 {
		t.Error("Limit field is zero")
	}

	if result.Results == nil {
		t.Error("Results field is nil (should be empty array)")
	}

	// Total should be a valid number (0 or more)
	if result.Total < 0 {
		t.Errorf("Total = %d, should be >= 0", result.Total)
	}

	// Offset should be a valid number (0 or more)
	if result.Offset < 0 {
		t.Errorf("Offset = %d, should be >= 0", result.Offset)
	}
}

func TestSearchConversations_HTTPStatusCodes(t *testing.T) {
	handler, _, cleanup := setupTestDataHandler(t)
	defer cleanup()

	tests := []struct {
		name           string
		url            string
		expectedStatus int
	}{
		{
			name:           "Valid query returns 200",
			url:            "/api/conversations/search?q=test",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing query returns 400",
			url:            "/api/conversations/search",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty query returns 400",
			url:            "/api/conversations/search?q=",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Query with special chars returns 200",
			url:            "/api/conversations/search?q=test%20auth",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			handler.SearchConversations(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d. Body: %s", w.Code, tt.expectedStatus, w.Body.String())
			}
		})
	}
}


func TestSearchConversations_ProjectFilter(t *testing.T) {
	handler, _, cleanup := setupTestDataHandler(t)
	defer cleanup()

	// Test that project filter parameter is accepted and doesn't cause errors
	req := httptest.NewRequest("GET", "/api/conversations/search?q=test&project=/test/path", nil)
	w := httptest.NewRecorder()

	handler.SearchConversations(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var result model.SearchResults
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Should return valid result structure even with no matches
	if result.Results == nil {
		t.Error("Results should be empty array, not nil")
	}
}
