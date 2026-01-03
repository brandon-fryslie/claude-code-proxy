package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/seifghazi/claude-code-monitor/internal/config"
)

func TestConversationIndexer(t *testing.T) {
	// Create a temporary directory for test data
	tmpDir, err := os.MkdirTemp("", "indexer-test")
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

	// Create indexer
	indexer, err := NewConversationIndexer(sqliteStorage)
	if err != nil {
		t.Fatalf("Failed to create indexer: %v", err)
	}

	// Test that indexer can be created
	if indexer == nil {
		t.Fatal("Indexer should not be nil")
	}

	// Test that indexer has required fields
	if indexer.storage == nil {
		t.Fatal("Indexer storage should not be nil")
	}

	if indexer.watcher == nil {
		t.Fatal("Indexer watcher should not be nil")
	}

	if indexer.indexQueue == nil {
		t.Fatal("Indexer indexQueue should not be nil")
	}

	if indexer.debounceTimers == nil {
		t.Fatal("Indexer debounceTimers should not be nil")
	}

	if indexer.done == nil {
		t.Fatal("Indexer done channel should not be nil")
	}

	// Test that Start returns no error
	err = indexer.Start()
	if err != nil {
		t.Fatalf("Indexer.Start() failed: %v", err)
	}

	// Give indexer a moment to start
	time.Sleep(100 * time.Millisecond)

	// Stop the indexer
	indexer.Stop()

	t.Log("✅ Indexer lifecycle test passed")
}

func TestExtractMessageContent(t *testing.T) {
	tests := []struct {
		name          string
		message       string
		expectedText  string
		expectedTools []string
		shouldError   bool
	}{
		{
			name:          "simple text message",
			message:       `{"role":"user","content":"Hello world"}`,
			expectedText:  "Hello world",
			expectedTools: nil,
			shouldError:   false,
		},
		{
			name:          "message with tool use",
			message:       `{"role":"assistant","content":[{"type":"text","text":"Let me run that"},{"type":"tool_use","name":"bash","id":"123"}]}`,
			expectedText:  "Let me run that",
			expectedTools: []string{"bash"},
			shouldError:   false,
		},
		{
			name:          "empty message",
			message:       `{"role":"user","content":""}`,
			expectedText:  "",
			expectedTools: nil,
			shouldError:   false,
		},
		{
			name:          "nil message",
			message:       "",
			expectedText:  "",
			expectedTools: nil,
			shouldError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &ConversationMessage{
				Message: []byte(tt.message),
			}

			text, tools, err := ExtractMessageContent(msg)

			if tt.shouldError && err == nil {
				t.Fatal("Expected error but got none")
			}

			if !tt.shouldError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if text != tt.expectedText {
				t.Errorf("Expected text %q, got %q", tt.expectedText, text)
			}

			if len(tools) != len(tt.expectedTools) {
				t.Errorf("Expected %d tools, got %d", len(tt.expectedTools), len(tools))
			}
		})
	}

	t.Log("✅ Message content extraction tests passed")
}

func TestNeedsIndexing(t *testing.T) {
	// Create a temporary directory for test data
	tmpDir, err := os.MkdirTemp("", "indexer-test")
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

	// Create indexer
	indexer, err := NewConversationIndexer(sqliteStorage)
	if err != nil {
		t.Fatalf("Failed to create indexer: %v", err)
	}

	// Test 1: New file should need indexing
	filePath := "/tmp/test.jsonl"
	mtime := time.Now()
	needs, err := indexer.needsIndexing(filePath, mtime)
	if err != nil {
		t.Fatalf("needsIndexing failed: %v", err)
	}
	if !needs {
		t.Error("New file should need indexing")
	}

	// Test 2: Insert a conversation and test staleness
	_, err = sqliteStorage.db.Exec(`
		INSERT INTO conversations (id, project_path, project_name, start_time, end_time, message_count, file_path, file_mtime, indexed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		"test-session",
		"/test/project",
		"test-project",
		time.Now().Format(time.RFC3339),
		time.Now().Format(time.RFC3339),
		0,
		filePath,
		mtime.Add(-1*time.Hour).Format(time.RFC3339),
		mtime.Add(-1*time.Hour).Format(time.RFC3339),
	)
	if err != nil {
		t.Fatalf("Failed to insert test conversation: %v", err)
	}

	// File modified after indexing should need re-indexing
	needs, err = indexer.needsIndexing(filePath, mtime)
	if err != nil {
		t.Fatalf("needsIndexing failed: %v", err)
	}
	if !needs {
		t.Error("Modified file should need re-indexing")
	}

	t.Log("✅ Staleness detection tests passed")
}
