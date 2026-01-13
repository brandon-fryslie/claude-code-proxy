package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ConversationIndexer manages the indexing of Claude Code conversation logs
type ConversationIndexer struct {
	storage        *SQLiteStorageService
	watcher        *fsnotify.Watcher
	indexQueue     chan string
	debounceTimers map[string]*time.Timer
	mu             sync.Mutex
	done           chan struct{}
	claudeProjects string
}

// NewConversationIndexer creates a new conversation indexer
func NewConversationIndexer(storage *SQLiteStorageService) (*ConversationIndexer, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	return &ConversationIndexer{
		storage:        storage,
		watcher:        watcher,
		indexQueue:     make(chan string, 100),
		debounceTimers: make(map[string]*time.Timer),
		done:           make(chan struct{}),
		claudeProjects: filepath.Join(homeDir, ".claude", "projects"),
	}, nil
}

// Start begins the indexing service
func (ci *ConversationIndexer) Start() error {
	log.Println("🔍 Starting conversation indexer...")

	// Start the index queue processor
	go ci.processIndexQueue()

	// Start the file watcher
	go ci.watchFiles()

	// Perform initial indexing
	go func() {
		if err := ci.initialIndex(); err != nil {
			log.Printf("❌ Initial indexing failed: %v", err)
		}
	}()

	return nil
}

// Stop cleanly shuts down the indexer
func (ci *ConversationIndexer) Stop() {
	log.Println("🛑 Stopping conversation indexer...")
	close(ci.done)
	ci.watcher.Close()
	close(ci.indexQueue)
}

// initialIndex walks the Claude projects directory and indexes all conversations
func (ci *ConversationIndexer) initialIndex() error {
	startTime := time.Now()
	log.Printf("📂 Starting initial indexing of %s", ci.claudeProjects)

	var fileCount int
	var indexedCount int

	err := filepath.Walk(ci.claudeProjects, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("⚠️  Error accessing %s: %v", path, err)
			return nil // Continue walking
		}

		if !strings.HasSuffix(path, ".jsonl") {
			return nil
		}

		fileCount++

		// Check if file needs indexing
		needsIndex, err := ci.needsIndexing(path, info.ModTime())
		if err != nil {
			log.Printf("⚠️  Error checking if %s needs indexing: %v", path, err)
			return nil
		}

		if needsIndex {
			if err := ci.indexFile(path); err != nil {
				log.Printf("⚠️  Error indexing %s: %v", path, err)
			} else {
				indexedCount++
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk Claude projects: %w", err)
	}

	duration := time.Since(startTime)
	log.Printf("✅ Initial indexing complete: %d/%d files indexed in %v", indexedCount, fileCount, duration)

	return nil
}

// needsIndexing checks if a file needs to be indexed based on modification time
func (ci *ConversationIndexer) needsIndexing(filePath string, mtime time.Time) (bool, error) {
	query := "SELECT indexed_at FROM conversations WHERE file_path = ?"
	var indexedAt sql.NullString

	err := ci.storage.db.QueryRow(query, filePath).Scan(&indexedAt)
	if err == sql.ErrNoRows {
		return true, nil // File not indexed yet
	}
	if err != nil {
		return false, err
	}

	if !indexedAt.Valid {
		return true, nil
	}

	// Parse indexed_at timestamp
	indexedTime, err := time.Parse(time.RFC3339, indexedAt.String)
	if err != nil {
		return true, nil // If we can't parse, re-index
	}

	// Re-index if file modified after last indexing
	return mtime.After(indexedTime), nil
}

// indexFile indexes a single JSONL conversation file
func (ci *ConversationIndexer) indexFile(filePath string) error {
	// Parse the conversation file
	projectDir := filepath.Dir(filePath)
	projectRelPath, err := filepath.Rel(ci.claudeProjects, projectDir)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}

	convService := NewConversationService()
	conv, err := convService.(*conversationService).parseConversationFile(filePath, projectRelPath)
	if err != nil {
		return fmt.Errorf("failed to parse conversation: %w", err)
	}

	if conv == nil {
		return nil // Empty conversation
	}

	// Get file modification time
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Start transaction
	tx, err := ci.storage.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Upsert conversation metadata
	_, err = tx.Exec(`
		INSERT OR REPLACE INTO conversations (id, project_path, project_name, start_time, end_time, message_count, file_path, file_mtime, indexed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		conv.SessionID,
		conv.ProjectPath,
		conv.ProjectName,
		conv.StartTime.Format(time.RFC3339),
		conv.EndTime.Format(time.RFC3339),
		conv.MessageCount,
		filePath,
		fileInfo.ModTime().Format(time.RFC3339),
		time.Now().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to insert conversation: %w", err)
	}

	// Delete existing entries for this conversation
	// Only delete FTS entries if FTS5 is enabled
	if fts5Enabled() {
		_, err = tx.Exec("DELETE FROM conversations_fts WHERE conversation_id = ?", conv.SessionID)
		if err != nil {
			return fmt.Errorf("failed to delete old FTS entries: %w", err)
		}
	}

	_, err = tx.Exec("DELETE FROM conversation_messages WHERE conversation_id = ?", conv.SessionID)
	if err != nil {
		return fmt.Errorf("failed to delete old message entries: %w", err)
	}

	// Prepare FTS insert statement only if FTS5 is enabled
	var ftsStmt *sql.Stmt
	if fts5Enabled() {
		ftsStmt, err = tx.Prepare(`
			INSERT INTO conversations_fts (conversation_id, message_uuid, message_type, content_text, tool_names, timestamp)
			VALUES (?, ?, ?, ?, ?, ?)
		`)
		if err != nil {
			return fmt.Errorf("failed to prepare FTS insert statement: %w", err)
		}
		defer ftsStmt.Close()
	}

	// Prepare full message insert statement (use INSERT OR REPLACE to handle duplicate UUIDs)
	msgStmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO conversation_messages (
			uuid, conversation_id, parent_uuid, type, role, timestamp,
			cwd, git_branch, session_id, agent_id, is_sidechain,
			request_id, model, input_tokens, output_tokens,
			cache_read_tokens, cache_creation_tokens, content_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare message insert statement: %w", err)
	}
	defer msgStmt.Close()

	for _, msg := range conv.Messages {
		// Skip messages without UUIDs (metadata messages like file-history-snapshot)
		if msg.UUID == "" {
			continue
		}

		// Extract content for FTS
		text, toolNames, _ := ExtractMessageContent(msg)
		toolNamesStr := strings.Join(toolNames, " ")

		// Insert into FTS if there's content and FTS5 is enabled
		if fts5Enabled() && (text != "" || len(toolNames) > 0) {
			_, err = ftsStmt.Exec(
				conv.SessionID,
				msg.UUID,
				msg.Type,
				text,
				toolNamesStr,
				msg.Timestamp,
			)
			if err != nil {
				log.Printf("⚠️  Error inserting FTS entry for message %s: %v", msg.UUID, err)
			}
		}

		// Parse message content for additional fields
		var msgContent MessageContent
		var role, model string
		var inputTokens, outputTokens, cacheRead, cacheCreation int

		if len(msg.Message) > 0 {
			if err := json.Unmarshal(msg.Message, &msgContent); err == nil {
				role = msgContent.Role
				model = msgContent.Model
				if msgContent.Usage != nil {
					inputTokens = msgContent.Usage.InputTokens
					outputTokens = msgContent.Usage.OutputTokens
					cacheRead = msgContent.Usage.CacheReadInputTokens
					cacheCreation = msgContent.Usage.CacheCreationInputTokens
				}
			}
		}

		// Insert full message data
		var parentUUID interface{}
		if msg.ParentUUID != nil {
			parentUUID = *msg.ParentUUID
		}

		_, err = msgStmt.Exec(
			msg.UUID,
			conv.SessionID,
			parentUUID,
			msg.Type,
			role,
			msg.Timestamp,
			msg.CWD,
			msg.GitBranch,
			msg.SessionID,
			msg.AgentID,
			msg.IsSidechain,
			msg.RequestID,
			model,
			inputTokens,
			outputTokens,
			cacheRead,
			cacheCreation,
			string(msg.Message),
		)
		if err != nil {
			log.Printf("⚠️  Error inserting message %s: %v", msg.UUID, err)
			continue
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// watchFiles sets up file watching for incremental updates
func (ci *ConversationIndexer) watchFiles() {
	// Add the Claude projects directory to the watcher
	if err := ci.watcher.Add(ci.claudeProjects); err != nil {
		log.Printf("❌ Failed to add watcher for %s: %v", ci.claudeProjects, err)
		return
	}

	// Also watch subdirectories
	filepath.Walk(ci.claudeProjects, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			ci.watcher.Add(path)
		}
		return nil
	})

	log.Println("👁️  File watcher started")

	for {
		select {
		case event, ok := <-ci.watcher.Events:
			if !ok {
				return
			}

			// Only process .jsonl files
			if !strings.HasSuffix(event.Name, ".jsonl") {
				continue
			}

			switch event.Op {
			case fsnotify.Write, fsnotify.Create:
				ci.debounceIndexing(event.Name)
			case fsnotify.Remove:
				ci.removeConversation(event.Name)
			}

		case err, ok := <-ci.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("⚠️  File watcher error: %v", err)

		case <-ci.done:
			return
		}
	}
}

// debounceIndexing debounces file indexing to avoid re-indexing during active writes
func (ci *ConversationIndexer) debounceIndexing(filePath string) {
	ci.mu.Lock()
	defer ci.mu.Unlock()

	// Cancel existing timer if any
	if timer, exists := ci.debounceTimers[filePath]; exists {
		timer.Stop()
	}

	// Start new 5-second debounce timer
	ci.debounceTimers[filePath] = time.AfterFunc(5*time.Second, func() {
		ci.indexQueue <- filePath

		ci.mu.Lock()
		delete(ci.debounceTimers, filePath)
		ci.mu.Unlock()
	})
}

// removeConversation removes a conversation from the index when the file is deleted
func (ci *ConversationIndexer) removeConversation(filePath string) {
	_, err := ci.storage.db.Exec("DELETE FROM conversations WHERE file_path = ?", filePath)
	if err != nil {
		log.Printf("⚠️  Error removing conversation %s: %v", filePath, err)
	}
	// FTS entries are deleted via CASCADE or we can do it explicitly
	// For now, assume we need to do it explicitly since FTS tables don't support CASCADE
	// Only delete FTS entries if FTS5 is enabled
	if fts5Enabled() {
		sessionID, err := ci.getSessionIDFromPath(filePath)
		if err == nil {
			ci.storage.db.Exec("DELETE FROM conversations_fts WHERE conversation_id = ?", sessionID)
		}
	}
}

// getSessionIDFromPath extracts the session ID from a file path
func (ci *ConversationIndexer) getSessionIDFromPath(filePath string) (string, error) {
	var sessionID string
	err := ci.storage.db.QueryRow("SELECT id FROM conversations WHERE file_path = ?", filePath).Scan(&sessionID)
	return sessionID, err
}

// processIndexQueue processes files from the index queue
func (ci *ConversationIndexer) processIndexQueue() {
	for filePath := range ci.indexQueue {
		if err := ci.indexFile(filePath); err != nil {
			log.Printf("⚠️  Error indexing %s: %v", filePath, err)
		} else {
			log.Printf("📝 Indexed conversation: %s", filepath.Base(filePath))
		}
	}
}

// IndexBenchmarkStats contains statistics from a full indexing run
type IndexBenchmarkStats struct {
	Duration          time.Duration
	FilesFound        int
	FilesIndexed      int
	ErrorCount        int
	ConversationCount int
	MessageCount      int
	FTSEntriesCount   int
}

// RunFullIndexBenchmark runs a complete indexing pass and returns statistics
// This is a synchronous version of initialIndex() designed for benchmarking
func (ci *ConversationIndexer) RunFullIndexBenchmark() (*IndexBenchmarkStats, error) {
	stats := &IndexBenchmarkStats{}
	startTime := time.Now()

	err := filepath.Walk(ci.claudeProjects, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}

		if !strings.HasSuffix(path, ".jsonl") {
			return nil
		}

		stats.FilesFound++

		// Always index for benchmark (ignore needsIndexing check)
		if err := ci.indexFile(path); err != nil {
			stats.ErrorCount++
		} else {
			stats.FilesIndexed++
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk Claude projects: %w", err)
	}

	stats.Duration = time.Since(startTime)

	// Query final counts
	ci.storage.db.QueryRow("SELECT COUNT(*) FROM conversations").Scan(&stats.ConversationCount)
	ci.storage.db.QueryRow("SELECT COUNT(*) FROM conversation_messages").Scan(&stats.MessageCount)

	if fts5Enabled() {
		ci.storage.db.QueryRow("SELECT COUNT(*) FROM conversations_fts").Scan(&stats.FTSEntriesCount)
	} else {
		stats.FTSEntriesCount = -1
	}

	return stats, nil
}

// DB returns the underlying database connection for benchmarking access
func (ci *ConversationIndexer) DB() *sql.DB {
	return ci.storage.db
}
