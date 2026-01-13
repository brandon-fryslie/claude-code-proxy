package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SessionDataIndexer manages indexing of Claude session data (todos, plans)
type SessionDataIndexer struct {
	storage   *SQLiteStorageService
	claudeDir string
}

// NewSessionDataIndexer creates a new session data indexer
func NewSessionDataIndexer(storage *SQLiteStorageService) (*SessionDataIndexer, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	return &SessionDataIndexer{
		storage:   storage,
		claudeDir: filepath.Join(homeDir, ".claude"),
	}, nil
}

// IndexTodos scans ~/.claude/todos/ and ingests into database
func (si *SessionDataIndexer) IndexTodos() (int, int, []string) {
	todosDir := filepath.Join(si.claudeDir, "todos")
	filesProcessed := 0
	todosIndexed := 0
	var errors []string

	err := filepath.Walk(todosDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		filesProcessed++

		// Parse filename: {session_uuid}-agent-{agent_uuid}.json
		baseName := strings.TrimSuffix(info.Name(), ".json")
		parts := strings.Split(baseName, "-agent-")
		sessionUUID := parts[0]
		agentUUID := ""
		if len(parts) > 1 {
			agentUUID = parts[1]
		}

		// Read and parse file
		content, err := os.ReadFile(path)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: read error: %v", info.Name(), err))
			return nil
		}

		if len(content) < 3 { // Empty or just "[]"
			// Still insert session with zero count
			if err := si.upsertTodoSession(path, sessionUUID, agentUUID, info.Size(), info.ModTime(), []TodoItem{}); err != nil {
				errors = append(errors, fmt.Sprintf("%s: session upsert error: %v", info.Name(), err))
			}
			return nil
		}

		var todos []TodoItem
		if err := json.Unmarshal(content, &todos); err != nil {
			errors = append(errors, fmt.Sprintf("%s: JSON parse error: %v", info.Name(), err))
			return nil
		}

		// Insert/update individual todos
		if err := si.upsertTodos(path, sessionUUID, agentUUID, info.ModTime(), todos); err != nil {
			errors = append(errors, fmt.Sprintf("%s: todos upsert error: %v", info.Name(), err))
			return nil
		}

		// Insert/update session aggregate
		if err := si.upsertTodoSession(path, sessionUUID, agentUUID, info.Size(), info.ModTime(), todos); err != nil {
			errors = append(errors, fmt.Sprintf("%s: session upsert error: %v", info.Name(), err))
			return nil
		}

		todosIndexed += len(todos)
		return nil
	})

	if err != nil {
		errors = append(errors, fmt.Sprintf("walk error: %v", err))
	}

	log.Printf("✅ Indexed %d todos from %d files", todosIndexed, filesProcessed)
	return filesProcessed, todosIndexed, errors
}

// IndexPlans scans ~/.claude/plans/ and ingests into database
func (si *SessionDataIndexer) IndexPlans() (int, []string) {
	plansDir := filepath.Join(si.claudeDir, "plans")
	plansIndexed := 0
	var errors []string

	err := filepath.Walk(plansDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: read error: %v", info.Name(), err))
			return nil
		}

		// Parse filename for display name: "peppy-yawning-teapot" -> "Peppy Yawning Teapot"
		baseName := strings.TrimSuffix(info.Name(), ".md")
		displayName := formatDisplayName(baseName)

		// Generate preview
		preview := string(content)
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}

		// Upsert into database
		if err := si.upsertPlan(info.Name(), displayName, string(content), preview, info.Size(), info.ModTime()); err != nil {
			errors = append(errors, fmt.Sprintf("%s: upsert error: %v", info.Name(), err))
			return nil
		}

		plansIndexed++
		return nil
	})

	if err != nil {
		errors = append(errors, fmt.Sprintf("walk error: %v", err))
	}

	log.Printf("✅ Indexed %d plans", plansIndexed)
	return plansIndexed, errors
}

// TodoItem represents a single todo entry
type TodoItem struct {
	Content    string `json:"content"`
	Status     string `json:"status"`
	ActiveForm string `json:"activeForm"`
}

// upsertTodos inserts or updates individual todo items in the database
func (si *SessionDataIndexer) upsertTodos(filePath, sessionUUID, agentUUID string, modTime time.Time, todos []TodoItem) error {
	tx, err := si.storage.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing todos for this file
	_, err = tx.Exec("DELETE FROM claude_todos WHERE file_path = ?", filePath)
	if err != nil {
		return fmt.Errorf("failed to delete existing todos: %w", err)
	}

	// Insert new todos
	stmt, err := tx.Prepare(`
		INSERT INTO claude_todos (session_uuid, agent_uuid, file_path, content, status, active_form, item_index, modified_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for i, todo := range todos {
		_, err = stmt.Exec(
			sessionUUID,
			agentUUID,
			filePath,
			todo.Content,
			todo.Status,
			todo.ActiveForm,
			i,
			modTime.Format(time.RFC3339),
		)
		if err != nil {
			return fmt.Errorf("failed to insert todo %d: %w", i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// upsertTodoSession inserts or updates the aggregated session stats
func (si *SessionDataIndexer) upsertTodoSession(filePath, sessionUUID, agentUUID string, fileSize int64, modTime time.Time, todos []TodoItem) error {
	// Calculate status counts
	var pendingCount, inProgressCount, completedCount int
	for _, todo := range todos {
		switch todo.Status {
		case "pending":
			pendingCount++
		case "in_progress":
			inProgressCount++
		case "completed":
			completedCount++
		}
	}

	_, err := si.storage.db.Exec(`
		INSERT OR REPLACE INTO claude_todo_sessions (
			session_uuid, agent_uuid, file_path, file_size, todo_count,
			pending_count, in_progress_count, completed_count, modified_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		sessionUUID,
		agentUUID,
		filePath,
		fileSize,
		len(todos),
		pendingCount,
		inProgressCount,
		completedCount,
		modTime.Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("failed to upsert todo session: %w", err)
	}

	return nil
}

// upsertPlan inserts or updates a plan document
func (si *SessionDataIndexer) upsertPlan(fileName, displayName, content, preview string, fileSize int64, modTime time.Time) error {
	_, err := si.storage.db.Exec(`
		INSERT OR REPLACE INTO claude_plans (
			file_name, display_name, content, preview, file_size, modified_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`,
		fileName,
		displayName,
		content,
		preview,
		fileSize,
		modTime.Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("failed to upsert plan: %w", err)
	}

	return nil
}

// formatDisplayName converts "peppy-yawning-teapot" to "Peppy Yawning Teapot"
func formatDisplayName(s string) string {
	words := strings.Split(s, "-")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
