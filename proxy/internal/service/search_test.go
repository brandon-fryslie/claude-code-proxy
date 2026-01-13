package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/seifghazi/claude-code-monitor/internal/config"
	"github.com/seifghazi/claude-code-monitor/internal/model"
)

func TestSearchConversations(t *testing.T) {
	// Skip this test if FTS5 is not available (test build)
	if !fts5Enabled() {
		t.Skip("Skipping FTS5 search test - FTS5 not available in test build")
		return
	}

	// Create a temporary directory for test data
	tmpDir, err := os.MkdirTemp("", "search-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test database
	dbPath := filepath.Join(tmpDir, "test.db")
	cfg := &config.StorageConfig{
		DBPath: dbPath,
	}

	storage, err := NewSQLiteStorageService(cfg)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	sqliteStorage, ok := storage.(*SQLiteStorageService)
	if !ok {
		t.Fatal("Storage must be SQLite")
	}

	// Insert test conversations
	tx, err := sqliteStorage.db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Conversation 1: Contains "authentication"
	_, err = tx.Exec(`
		INSERT INTO conversations (id, project_path, project_name, start_time, end_time, message_count, file_path, file_mtime, indexed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		"conv-1",
		"/test/auth-project",
		"auth-project",
		time.Now().Format(time.RFC3339),
		time.Now().Format(time.RFC3339),
		2,
		"/test/conv-1.jsonl",
		time.Now().Format(time.RFC3339),
		time.Now().Format(time.RFC3339),
	)
	if err != nil {
		t.Fatalf("Failed to insert conversation 1: %v", err)
	}

	_, err = tx.Exec(`
		INSERT INTO conversations_fts (conversation_id, message_uuid, message_type, content_text, tool_names, timestamp)
		VALUES (?, ?, ?, ?, ?, ?)
	`, "conv-1", "msg-1", "user", "I need help with authentication", "", time.Now().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("Failed to insert FTS entry 1: %v", err)
	}

	// Conversation 2: Contains "bug"
	_, err = tx.Exec(`
		INSERT INTO conversations (id, project_path, project_name, start_time, end_time, message_count, file_path, file_mtime, indexed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		"conv-2",
		"/test/bug-project",
		"bug-project",
		time.Now().Format(time.RFC3339),
		time.Now().Format(time.RFC3339),
		1,
		"/test/conv-2.jsonl",
		time.Now().Format(time.RFC3339),
		time.Now().Format(time.RFC3339),
	)
	if err != nil {
		t.Fatalf("Failed to insert conversation 2: %v", err)
	}

	_, err = tx.Exec(`
		INSERT INTO conversations_fts (conversation_id, message_uuid, message_type, content_text, tool_names, timestamp)
		VALUES (?, ?, ?, ?, ?, ?)
	`, "conv-2", "msg-2", "user", "There is a bug in the code", "", time.Now().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("Failed to insert FTS entry 2: %v", err)
	}

	// Conversation 3: Contains both "authentication" and "bug"
	_, err = tx.Exec(`
		INSERT INTO conversations (id, project_path, project_name, start_time, end_time, message_count, file_path, file_mtime, indexed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		"conv-3",
		"/test/mixed-project",
		"mixed-project",
		time.Now().Format(time.RFC3339),
		time.Now().Format(time.RFC3339),
		1,
		"/test/conv-3.jsonl",
		time.Now().Format(time.RFC3339),
		time.Now().Format(time.RFC3339),
	)
	if err != nil {
		t.Fatalf("Failed to insert conversation 3: %v", err)
	}

	_, err = tx.Exec(`
		INSERT INTO conversations_fts (conversation_id, message_uuid, message_type, content_text, tool_names, timestamp)
		VALUES (?, ?, ?, ?, ?, ?)
	`, "conv-3", "msg-3", "user", "Fix authentication bug", "", time.Now().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("Failed to insert FTS entry 3: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Test 1: Search for "authentication"
	t.Run("search for authentication", func(t *testing.T) {
		results, err := storage.SearchConversations(model.SearchOptions{
			Query:  "authentication",
			Limit:  50,
			Offset: 0,
		})

		if err != nil {
			t.Fatalf("SearchConversations failed: %v", err)
		}

		if results.Total < 2 {
			t.Errorf("Expected at least 2 results for 'authentication', got %d", results.Total)
		}

		if len(results.Results) == 0 {
			t.Fatal("Expected results array to have items")
		}

		t.Logf("✅ Found %d conversations matching 'authentication'", results.Total)
	})

	// Test 2: Multi-term search (OR logic)
	t.Run("multi-term search", func(t *testing.T) {
		results, err := storage.SearchConversations(model.SearchOptions{
			Query:  "authentication bug",
			Limit:  50,
			Offset: 0,
		})

		if err != nil {
			t.Fatalf("SearchConversations failed: %v", err)
		}

		// Should match all 3 conversations (OR logic):
		// - conv-1: matches "authentication"
		// - conv-2: matches "bug"
		// - conv-3: matches both "authentication" and "bug"
		if results.Total < 3 {
			t.Errorf("Expected at least 3 results for 'authentication bug' (OR logic), got %d", results.Total)
		}

		t.Logf("✅ Found %d conversations matching 'authentication bug' (OR logic)", results.Total)
	})

	// Test 3: Empty query
	t.Run("empty query", func(t *testing.T) {
		results, err := storage.SearchConversations(model.SearchOptions{
			Query:  "",
			Limit:  50,
			Offset: 0,
		})

		if err != nil {
			t.Fatalf("SearchConversations with empty query failed: %v", err)
		}

		if results.Total != 0 {
			t.Errorf("Expected 0 results for empty query, got %d", results.Total)
		}

		if results.Results == nil {
			t.Error("Results array should not be nil (should be empty array)")
		}

		t.Log("✅ Empty query handled correctly")
	})

	// Test 4: No matches
	t.Run("no matches", func(t *testing.T) {
		results, err := storage.SearchConversations(model.SearchOptions{
			Query:  "nonexistent",
			Limit:  50,
			Offset: 0,
		})

		if err != nil {
			t.Fatalf("SearchConversations failed: %v", err)
		}

		if results.Total != 0 {
			t.Errorf("Expected 0 results for 'nonexistent', got %d", results.Total)
		}

		t.Log("✅ No matches handled correctly")
	})

	// Test 5: Pagination
	t.Run("pagination", func(t *testing.T) {
		// Get first page (limit 2)
		page1, err := storage.SearchConversations(model.SearchOptions{
			Query:  "authentication bug",
			Limit:  2,
			Offset: 0,
		})

		if err != nil {
			t.Fatalf("SearchConversations page 1 failed: %v", err)
		}

		// Get second page
		page2, err := storage.SearchConversations(model.SearchOptions{
			Query:  "authentication bug",
			Limit:  2,
			Offset: 2,
		})

		if err != nil {
			t.Fatalf("SearchConversations page 2 failed: %v", err)
		}

		// Both pages should have the same total count
		if page1.Total != page2.Total {
			t.Errorf("Page 1 total (%d) != Page 2 total (%d)", page1.Total, page2.Total)
		}

		// Page 1 should have results (if total >= 2)
		if page1.Total >= 2 && len(page1.Results) == 0 {
			t.Error("Expected page 1 to have results")
		}

		t.Logf("✅ Pagination works: page1 has %d results, page2 has %d results", len(page1.Results), len(page2.Results))
	})

	// Test 6: Project filter
	t.Run("project filter", func(t *testing.T) {
		results, err := storage.SearchConversations(model.SearchOptions{
			Query:       "authentication",
			ProjectPath: "/test/auth-project",
			Limit:       50,
			Offset:      0,
		})

		if err != nil {
			t.Fatalf("SearchConversations with project filter failed: %v", err)
		}

		// Should only match conversations from auth-project
		for _, match := range results.Results {
			if match.ProjectPath != "/test/auth-project" {
				t.Errorf("Expected only results from /test/auth-project, got %s", match.ProjectPath)
			}
		}

		t.Logf("✅ Project filter works: found %d results in auth-project", len(results.Results))
	})

	t.Log("✅ All search tests passed")
}

func TestSearchConversationsResponseFormat(t *testing.T) {
	// Create a temporary directory for test data
	tmpDir, err := os.MkdirTemp("", "search-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test database
	dbPath := filepath.Join(tmpDir, "test.db")
	cfg := &config.StorageConfig{
		DBPath: dbPath,
	}

	storage, err := NewSQLiteStorageService(cfg)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Test response format
	results, err := storage.SearchConversations(model.SearchOptions{
		Query:  "test",
		Limit:  50,
		Offset: 0,
	})

	if err != nil {
		t.Fatalf("SearchConversations failed: %v", err)
	}

	// Verify response structure
	if results.Query != "test" {
		t.Errorf("Expected query 'test', got '%s'", results.Query)
	}

	if results.Limit != 50 {
		t.Errorf("Expected limit 50, got %d", results.Limit)
	}

	if results.Offset != 0 {
		t.Errorf("Expected offset 0, got %d", results.Offset)
	}

	if results.Results == nil {
		t.Error("Results array should not be nil (should be empty array)")
	}

	t.Log("✅ Response format test passed")
}
