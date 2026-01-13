package service

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
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

// TestIndexerWithRealData is a P1 integration test that indexes real JSONL files
// and verifies they are correctly stored in the database
func TestIndexerWithRealData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary directory for test database
	tmpDir, err := os.MkdirTemp("", "indexer-realdata-test")
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

	// Find real JSONL files
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	claudeProjectsPath := filepath.Join(homeDir, ".claude", "projects")
	if _, err := os.Stat(claudeProjectsPath); os.IsNotExist(err) {
		t.Skipf("Claude projects directory not found: %s", claudeProjectsPath)
	}

	// Find up to 15 real JSONL files to index
	var filesToIndex []string
	err = filepath.Walk(claudeProjectsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		if !strings.HasSuffix(path, ".jsonl") {
			return nil
		}

		// Skip very large files (>10MB) to keep test fast
		if info.Size() > 10*1024*1024 {
			return nil
		}

		filesToIndex = append(filesToIndex, path)

		// Collect at least 15 files
		if len(filesToIndex) >= 15 {
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk Claude projects: %v", err)
	}

	if len(filesToIndex) == 0 {
		t.Skip("No JSONL files found to test with")
	}

	t.Logf("Found %d JSONL files to index", len(filesToIndex))

	// Index each file
	var successCount, errorCount int
	for _, filePath := range filesToIndex {
		if err := indexer.indexFile(filePath); err != nil {
			t.Logf("⚠️  Error indexing %s: %v", filepath.Base(filePath), err)
			errorCount++
		} else {
			successCount++
		}
	}

	t.Logf("✅ Indexed %d/%d files successfully", successCount, len(filesToIndex))

	// Verify conversations are in the database
	var conversationCount int
	err = sqliteStorage.db.QueryRow("SELECT COUNT(*) FROM conversations").Scan(&conversationCount)
	if err != nil {
		t.Fatalf("Failed to count conversations: %v", err)
	}

	if conversationCount == 0 {
		t.Fatal("No conversations found in database after indexing")
	}

	t.Logf("📊 Database contains %d conversations", conversationCount)

	// Verify messages are in the database
	var messageCount int
	err = sqliteStorage.db.QueryRow("SELECT COUNT(*) FROM conversation_messages").Scan(&messageCount)
	if err != nil {
		t.Fatalf("Failed to count messages: %v", err)
	}

	if messageCount == 0 {
		t.Fatal("No messages found in database after indexing")
	}

	t.Logf("📊 Database contains %d messages", messageCount)

	// Verify messages have content
	var messageWithContent int
	err = sqliteStorage.db.QueryRow("SELECT COUNT(*) FROM conversation_messages WHERE content_json IS NOT NULL AND content_json != ''").Scan(&messageWithContent)
	if err != nil {
		t.Fatalf("Failed to count messages with content: %v", err)
	}

	if messageWithContent == 0 {
		t.Fatal("No messages with content found in database")
	}

	t.Logf("📊 %d messages have content", messageWithContent)

	// Test re-indexing (should be no-op for unchanged files)
	for _, filePath := range filesToIndex[:3] {
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		needs, err := indexer.needsIndexing(filePath, fileInfo.ModTime())
		if err != nil {
			t.Errorf("needsIndexing failed for %s: %v", filePath, err)
			continue
		}

		if needs {
			t.Errorf("File %s should not need re-indexing (no changes)", filepath.Base(filePath))
		}
	}

	t.Log("✅ Integration test with real data passed")
}

// TestSearchIndexedConversations tests that indexed conversations are searchable
func TestSearchIndexedConversations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary directory for test database
	tmpDir, err := os.MkdirTemp("", "indexer-search-test")
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

	// Find and index a few real JSONL files
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	claudeProjectsPath := filepath.Join(homeDir, ".claude", "projects")
	if _, err := os.Stat(claudeProjectsPath); os.IsNotExist(err) {
		t.Skipf("Claude projects directory not found: %s", claudeProjectsPath)
	}

	// Index up to 5 files
	var filesIndexed int
	err = filepath.Walk(claudeProjectsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || filesIndexed >= 5 {
			return nil
		}

		if !strings.HasSuffix(path, ".jsonl") || info.Size() > 5*1024*1024 {
			return nil
		}

		if err := indexer.indexFile(path); err == nil {
			filesIndexed++
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk Claude projects: %v", err)
	}

	if filesIndexed == 0 {
		t.Skip("No files indexed for search test")
	}

	t.Logf("✅ Indexed %d files for search testing", filesIndexed)

	// Test basic retrieval (not full-text search, since FTS5 is disabled in test mode)
	var sessionID string
	err = sqliteStorage.db.QueryRow("SELECT id FROM conversations LIMIT 1").Scan(&sessionID)
	if err != nil {
		t.Fatalf("Failed to get a conversation ID: %v", err)
	}

	// Verify we can retrieve messages for this conversation
	rows, err := sqliteStorage.db.Query("SELECT uuid, type, timestamp FROM conversation_messages WHERE conversation_id = ?", sessionID)
	if err != nil {
		t.Fatalf("Failed to query messages: %v", err)
	}
	defer rows.Close()

	var messageCount int
	for rows.Next() {
		var uuid, msgType, timestamp string
		if err := rows.Scan(&uuid, &msgType, &timestamp); err != nil {
			t.Errorf("Failed to scan message row: %v", err)
			continue
		}
		messageCount++

		// Verify UUID is not empty
		if uuid == "" {
			t.Error("Message UUID should not be empty")
		}

		// Verify timestamp can be parsed
		if _, err := time.Parse(time.RFC3339, timestamp); err != nil {
			t.Errorf("Failed to parse timestamp %q: %v", timestamp, err)
		}
	}

	if messageCount == 0 {
		t.Fatal("No messages found for conversation")
	}

	t.Logf("✅ Found %d messages for conversation %s", messageCount, sessionID)
	t.Log("✅ Search integration test passed")
}

// TestFileWatcherDetectsChanges tests that the file watcher detects modifications
func TestFileWatcherDetectsChanges(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary directory for test data and database
	tmpDir, err := os.MkdirTemp("", "indexer-watcher-test")
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

	// Create a custom indexer with a test directory
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	testProjectsDir := filepath.Join(tmpDir, "projects")
	if err := os.MkdirAll(testProjectsDir, 0755); err != nil {
		t.Fatalf("Failed to create test projects dir: %v", err)
	}

	indexer := &ConversationIndexer{
		storage:        sqliteStorage,
		watcher:        watcher,
		indexQueue:     make(chan string, 100),
		debounceTimers: make(map[string]*time.Timer),
		mu:             sync.Mutex{},
		done:           make(chan struct{}),
		claudeProjects: testProjectsDir,
	}

	// Start the indexer
	if err := indexer.Start(); err != nil {
		t.Fatalf("Failed to start indexer: %v", err)
	}
	defer indexer.Stop()

	// Give the indexer time to start
	time.Sleep(200 * time.Millisecond)

	// Create a test JSONL file
	testFile := filepath.Join(testProjectsDir, "test-session.jsonl")
	content := `{"uuid":"msg-001","timestamp":"2024-01-01T10:00:00Z","sessionId":"test-session","type":"message","userType":"user","message":{"role":"user","content":"Hello"},"cwd":"/tmp"}
{"uuid":"msg-002","timestamp":"2024-01-01T10:00:01Z","sessionId":"test-session","type":"message","userType":"assistant","message":{"role":"assistant","content":"Hi there!"},"cwd":"/tmp"}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Wait for debounce timer (5 seconds) + processing time
	t.Log("⏳ Waiting for file to be indexed (6 seconds)...")
	time.Sleep(6 * time.Second)

	// Verify the conversation was indexed
	var conversationCount int
	err = sqliteStorage.db.QueryRow("SELECT COUNT(*) FROM conversations WHERE id = ?", "test-session").Scan(&conversationCount)
	if err != nil {
		t.Fatalf("Failed to count conversations: %v", err)
	}

	if conversationCount == 0 {
		t.Error("File watcher did not index the new file")
	} else {
		t.Log("✅ File watcher detected and indexed new file")
	}

	// Verify messages were indexed
	var messageCount int
	err = sqliteStorage.db.QueryRow("SELECT COUNT(*) FROM conversation_messages WHERE conversation_id = ?", "test-session").Scan(&messageCount)
	if err != nil {
		t.Fatalf("Failed to count messages: %v", err)
	}

	if messageCount != 2 {
		t.Errorf("Expected 2 messages, got %d", messageCount)
	} else {
		t.Logf("✅ File watcher indexed %d messages", messageCount)
	}

	t.Log("✅ File watcher integration test passed")
}
