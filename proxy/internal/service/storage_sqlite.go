package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/seifghazi/claude-code-monitor/internal/config"
	"github.com/seifghazi/claude-code-monitor/internal/model"
)

type SQLiteStorageService struct {
	db     *sql.DB
	config *config.StorageConfig
}

func NewSQLiteStorageService(cfg *config.StorageConfig) (StorageService, error) {
	// Add SQLite-specific connection parameters for better concurrency
	dbPath := cfg.DBPath + "?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL"

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	service := &SQLiteStorageService{
		db:     db,
		config: cfg,
	}

	if err := service.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return service, nil
}

func (s *SQLiteStorageService) createTables() error {
	// Check if table exists
	var tableExists int
	err := s.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='requests'").Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("failed to check if table exists: %w", err)
	}

	if tableExists == 0 {
		// Fresh database - create table with all columns including new ones
		schema := `
		CREATE TABLE requests (
			id TEXT PRIMARY KEY,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			method TEXT NOT NULL,
			endpoint TEXT NOT NULL,
			headers TEXT NOT NULL,
			body TEXT NOT NULL,
			user_agent TEXT,
			content_type TEXT,
			prompt_grade TEXT,
			response TEXT,
			model TEXT,
			original_model TEXT,
			routed_model TEXT,
			provider TEXT,
			subagent_name TEXT,
			tools_used TEXT,
			tool_call_count INTEGER DEFAULT 0,
			input_tokens INTEGER DEFAULT 0,
			output_tokens INTEGER DEFAULT 0,
			cache_read_tokens INTEGER DEFAULT 0,
			cache_creation_tokens INTEGER DEFAULT 0,
			response_time_ms INTEGER DEFAULT 0,
			first_byte_time_ms INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX idx_timestamp ON requests(timestamp DESC);
		CREATE INDEX idx_endpoint ON requests(endpoint);
		CREATE INDEX idx_model ON requests(model);
		CREATE INDEX idx_provider ON requests(provider);
		CREATE INDEX idx_subagent ON requests(subagent_name);
		CREATE INDEX idx_timestamp_provider ON requests(timestamp DESC, provider);
		`
		_, err := s.db.Exec(schema)
		if err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	} else {
		// Existing database - run migrations to add new columns
		if err := s.runMigrations(); err != nil {
			return err
		}
	}
	// ALWAYS run conversation search migrations (for both fresh and existing databases)
	if err := s.runConversationSearchMigrations(); err != nil {
		return err
	}

	// ALWAYS run Claude session data migrations (todos, plans)
	if err := s.runClaudeSessionDataMigrations(); err != nil {
		return err
	}

	return nil
}

func (s *SQLiteStorageService) runMigrations() error {
	// Add new columns if they don't exist (for existing databases)
	migrations := []string{
		"ALTER TABLE requests ADD COLUMN provider TEXT",
		"ALTER TABLE requests ADD COLUMN subagent_name TEXT",
		"ALTER TABLE requests ADD COLUMN tools_used TEXT",
		"ALTER TABLE requests ADD COLUMN tool_call_count INTEGER DEFAULT 0",
		"ALTER TABLE requests ADD COLUMN input_tokens INTEGER DEFAULT 0",
		"ALTER TABLE requests ADD COLUMN output_tokens INTEGER DEFAULT 0",
		"ALTER TABLE requests ADD COLUMN cache_read_tokens INTEGER DEFAULT 0",
		"ALTER TABLE requests ADD COLUMN cache_creation_tokens INTEGER DEFAULT 0",
		"ALTER TABLE requests ADD COLUMN response_time_ms INTEGER DEFAULT 0",
		"ALTER TABLE requests ADD COLUMN first_byte_time_ms INTEGER DEFAULT 0",
	}

	for _, migration := range migrations {
		// Ignore errors - column may already exist
		s.db.Exec(migration)
	}

	// Create new indexes (ignore errors if they exist)
	s.db.Exec("CREATE INDEX IF NOT EXISTS idx_provider ON requests(provider)")
	s.db.Exec("CREATE INDEX IF NOT EXISTS idx_subagent ON requests(subagent_name)")
	s.db.Exec("CREATE INDEX IF NOT EXISTS idx_timestamp_provider ON requests(timestamp DESC, provider)")


	return nil
}

// runConversationSearchMigrations creates the conversation search tables and FTS5 index
func (s *SQLiteStorageService) runConversationSearchMigrations() error {
	// Check if conversations table already exists
	var conversationsExists int
	err := s.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='conversations'").Scan(&conversationsExists)
	if err != nil {
		return fmt.Errorf("failed to check if conversations table exists: %w", err)
	}

	if conversationsExists == 0 {
		// Create conversations metadata table
		conversationsSchema := `
		CREATE TABLE conversations (
			id TEXT PRIMARY KEY,
			project_path TEXT NOT NULL,
			project_name TEXT NOT NULL,
			start_time DATETIME,
			end_time DATETIME,
			message_count INTEGER DEFAULT 0,
			file_path TEXT NOT NULL UNIQUE,
			file_mtime DATETIME,
			indexed_at DATETIME
		);

		CREATE INDEX idx_conversations_project ON conversations(project_path);
		CREATE INDEX idx_conversations_mtime ON conversations(file_mtime DESC);
		CREATE INDEX idx_conversations_indexed ON conversations(indexed_at DESC);
		`

		if _, err := s.db.Exec(conversationsSchema); err != nil {
			return fmt.Errorf("failed to create conversations table: %w", err)
		}

		log.Println("✅ Created conversations table")
	}

	// Check if FTS5 table exists
	var ftsExists int
	err = s.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='conversations_fts'").Scan(&ftsExists)
	if err != nil {
		return fmt.Errorf("failed to check if FTS table exists: %w", err)
	}

	if ftsExists == 0 {
		// Create FTS5 virtual table
		ftsSchema := `
		CREATE VIRTUAL TABLE conversations_fts USING fts5(
			conversation_id UNINDEXED,
			message_uuid UNINDEXED,
			message_type,
			content_text,
			tool_names,
			timestamp UNINDEXED,
			tokenize='porter unicode61'
		);
		`

		if _, err := s.db.Exec(ftsSchema); err != nil {
			return fmt.Errorf("failed to create FTS table: %w", err)
		}

		log.Println("✅ Created conversations_fts FTS5 table")
	}

	// Check if conversation_messages table exists
	var messagesExists int
	err = s.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='conversation_messages'").Scan(&messagesExists)
	if err != nil {
		return fmt.Errorf("failed to check if conversation_messages table exists: %w", err)
	}

	if messagesExists == 0 {
		// Create conversation_messages table to store full message data
		messagesSchema := `
		CREATE TABLE conversation_messages (
			uuid TEXT PRIMARY KEY,
			conversation_id TEXT NOT NULL,
			parent_uuid TEXT,
			type TEXT NOT NULL,
			role TEXT,
			timestamp DATETIME NOT NULL,
			cwd TEXT,
			git_branch TEXT,
			session_id TEXT,
			agent_id TEXT,
			is_sidechain BOOLEAN DEFAULT FALSE,
			request_id TEXT,
			model TEXT,
			input_tokens INTEGER DEFAULT 0,
			output_tokens INTEGER DEFAULT 0,
			cache_read_tokens INTEGER DEFAULT 0,
			cache_creation_tokens INTEGER DEFAULT 0,
			content_json TEXT,
			tool_use_json TEXT,
			tool_result_json TEXT,
			FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
		);

		CREATE INDEX idx_messages_conversation ON conversation_messages(conversation_id);
		CREATE INDEX idx_messages_timestamp ON conversation_messages(timestamp);
		CREATE INDEX idx_messages_parent ON conversation_messages(parent_uuid);
		CREATE INDEX idx_messages_session ON conversation_messages(session_id);
		CREATE INDEX idx_messages_agent ON conversation_messages(agent_id);
		CREATE INDEX idx_messages_request ON conversation_messages(request_id);
		`

		if _, err := s.db.Exec(messagesSchema); err != nil {
			return fmt.Errorf("failed to create conversation_messages table: %w", err)
		}

		log.Println("✅ Created conversation_messages table")
	}

	return nil
}

// runClaudeSessionDataMigrations creates tables for todos and plans
func (s *SQLiteStorageService) runClaudeSessionDataMigrations() error {
	// Check if claude_todos table exists
	var todosExists int
	err := s.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='claude_todos'").Scan(&todosExists)
	if err != nil {
		return fmt.Errorf("failed to check if claude_todos table exists: %w", err)
	}

	if todosExists == 0 {
		// Create claude_todos table
		todosSchema := `
		CREATE TABLE claude_todos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_uuid TEXT NOT NULL,
			agent_uuid TEXT,
			file_path TEXT NOT NULL,
			content TEXT NOT NULL,
			status TEXT NOT NULL,
			active_form TEXT,
			item_index INTEGER NOT NULL,
			modified_at DATETIME NOT NULL,
			indexed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(file_path, item_index)
		);

		CREATE INDEX idx_todos_session ON claude_todos(session_uuid);
		CREATE INDEX idx_todos_status ON claude_todos(status);
		CREATE INDEX idx_todos_modified ON claude_todos(modified_at);
		`

		if _, err := s.db.Exec(todosSchema); err != nil {
			return fmt.Errorf("failed to create claude_todos table: %w", err)
		}

		log.Println("✅ Created claude_todos table")
	}

	// Check if claude_todo_sessions table exists
	var todoSessionsExists int
	err = s.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='claude_todo_sessions'").Scan(&todoSessionsExists)
	if err != nil {
		return fmt.Errorf("failed to check if claude_todo_sessions table exists: %w", err)
	}

	if todoSessionsExists == 0 {
		// Create claude_todo_sessions table
		todoSessionsSchema := `
		CREATE TABLE claude_todo_sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_uuid TEXT NOT NULL,
			agent_uuid TEXT,
			file_path TEXT UNIQUE NOT NULL,
			file_size INTEGER NOT NULL,
			todo_count INTEGER NOT NULL,
			pending_count INTEGER NOT NULL,
			in_progress_count INTEGER NOT NULL,
			completed_count INTEGER NOT NULL,
			modified_at DATETIME NOT NULL,
			indexed_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		`

		if _, err := s.db.Exec(todoSessionsSchema); err != nil {
			return fmt.Errorf("failed to create claude_todo_sessions table: %w", err)
		}

		log.Println("✅ Created claude_todo_sessions table")
	}

	// Check if claude_plans table exists
	var plansExists int
	err = s.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='claude_plans'").Scan(&plansExists)
	if err != nil {
		return fmt.Errorf("failed to check if claude_plans table exists: %w", err)
	}

	if plansExists == 0 {
		// Create claude_plans table
		plansSchema := `
		CREATE TABLE claude_plans (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_name TEXT UNIQUE NOT NULL,
			display_name TEXT NOT NULL,
			content TEXT NOT NULL,
			preview TEXT NOT NULL,
			file_size INTEGER NOT NULL,
			modified_at DATETIME NOT NULL,
			indexed_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX idx_plans_modified ON claude_plans(modified_at);
		`

		if _, err := s.db.Exec(plansSchema); err != nil {
			return fmt.Errorf("failed to create claude_plans table: %w", err)
		}

		log.Println("✅ Created claude_plans table")
	}

	return nil
}

func (s *SQLiteStorageService) SaveRequest(request *model.RequestLog) (string, error) {
	headersJSON, err := json.Marshal(request.Headers)
	if err != nil {
		return "", fmt.Errorf("failed to marshal headers: %w", err)
	}

	bodyJSON, err := json.Marshal(request.Body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal body: %w", err)
	}

	toolsUsedJSON, err := json.Marshal(request.ToolsUsed)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tools_used: %w", err)
	}

	query := `
		INSERT INTO requests (id, timestamp, method, endpoint, headers, body, user_agent, content_type, model, original_model, routed_model, provider, subagent_name, tools_used, tool_call_count)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(query,
		request.RequestID,
		request.Timestamp,
		request.Method,
		request.Endpoint,
		string(headersJSON),
		string(bodyJSON),
		request.UserAgent,
		request.ContentType,
		request.Model,
		request.OriginalModel,
		request.RoutedModel,
		request.Provider,
		request.SubagentName,
		string(toolsUsedJSON),
		request.ToolCallCount,
	)

	if err != nil {
		return "", fmt.Errorf("failed to insert request: %w", err)
	}

	return request.RequestID, nil
}

func (s *SQLiteStorageService) GetRequests(page, limit int) ([]model.RequestLog, int, error) {
	// Get total count
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM requests").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get paginated results
	offset := (page - 1) * limit
	query := `
		SELECT id, timestamp, method, endpoint, headers, body, model, user_agent, content_type, prompt_grade, response, original_model, routed_model
		FROM requests
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query requests: %w", err)
	}
	defer rows.Close()

	var requests []model.RequestLog
	for rows.Next() {
		var req model.RequestLog
		var headersJSON, bodyJSON string
		var promptGradeJSON, responseJSON sql.NullString

		err := rows.Scan(
			&req.RequestID,
			&req.Timestamp,
			&req.Method,
			&req.Endpoint,
			&headersJSON,
			&bodyJSON,
			&req.Model,
			&req.UserAgent,
			&req.ContentType,
			&promptGradeJSON,
			&responseJSON,
			&req.OriginalModel,
			&req.RoutedModel,
		)
		if err != nil {
			// Error scanning row - skip
			continue
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal([]byte(headersJSON), &req.Headers); err != nil {
			// Error unmarshaling headers
			continue
		}

		var body interface{}
		if err := json.Unmarshal([]byte(bodyJSON), &body); err != nil {
			// Error unmarshaling body
			continue
		}
		req.Body = body

		if promptGradeJSON.Valid {
			var grade model.PromptGrade
			if err := json.Unmarshal([]byte(promptGradeJSON.String), &grade); err == nil {
				req.PromptGrade = &grade
			}
		}

		if responseJSON.Valid {
			var resp model.ResponseLog
			if err := json.Unmarshal([]byte(responseJSON.String), &resp); err == nil {
				req.Response = &resp
			}
		}

		requests = append(requests, req)
	}

	return requests, total, nil
}

func (s *SQLiteStorageService) ClearRequests() (int, error) {
	result, err := s.db.Exec("DELETE FROM requests")
	if err != nil {
		return 0, fmt.Errorf("failed to clear requests: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}

func (s *SQLiteStorageService) UpdateRequestWithGrading(requestID string, grade *model.PromptGrade) error {
	gradeJSON, err := json.Marshal(grade)
	if err != nil {
		return fmt.Errorf("failed to marshal grade: %w", err)
	}

	query := "UPDATE requests SET prompt_grade = ? WHERE id = ?"
	_, err = s.db.Exec(query, string(gradeJSON), requestID)
	if err != nil {
		return fmt.Errorf("failed to update request with grading: %w", err)
	}

	return nil
}

func (s *SQLiteStorageService) UpdateRequestWithResponse(request *model.RequestLog) error {
	responseJSON, err := json.Marshal(request.Response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Extract token counts and timing from response for indexed columns
	var inputTokens, outputTokens, cacheReadTokens, cacheCreationTokens int
	var responseTimeMs, firstByteTimeMs int64
	var toolCallCount int

	if request.Response != nil {
		responseTimeMs = request.Response.ResponseTime
		firstByteTimeMs = request.Response.FirstByteTime
		toolCallCount = request.Response.ToolCallCount

		// Extract usage from response body
		if request.Response.Body != nil {
			var respBody struct {
				Usage *model.AnthropicUsage `json:"usage"`
			}
			if err := json.Unmarshal(request.Response.Body, &respBody); err == nil && respBody.Usage != nil {
				inputTokens = respBody.Usage.InputTokens
				outputTokens = respBody.Usage.OutputTokens
				cacheReadTokens = respBody.Usage.CacheReadInputTokens
				cacheCreationTokens = respBody.Usage.CacheCreationInputTokens
			}
		}
	}

	query := `UPDATE requests SET
		response = ?,
		input_tokens = ?,
		output_tokens = ?,
		cache_read_tokens = ?,
		cache_creation_tokens = ?,
		response_time_ms = ?,
		first_byte_time_ms = ?,
		tool_call_count = ?
		WHERE id = ?`

	_, err = s.db.Exec(query,
		string(responseJSON),
		inputTokens,
		outputTokens,
		cacheReadTokens,
		cacheCreationTokens,
		responseTimeMs,
		firstByteTimeMs,
		toolCallCount,
		request.RequestID,
	)
	if err != nil {
		return fmt.Errorf("failed to update request with response: %w", err)
	}

	return nil
}

func (s *SQLiteStorageService) EnsureDirectoryExists() error {
	// No directory needed for SQLite
	return nil
}

func (s *SQLiteStorageService) GetRequestByShortID(shortID string) (*model.RequestLog, string, error) {
	query := `
		SELECT id, timestamp, method, endpoint, headers, body, model, user_agent, content_type, prompt_grade, response, original_model, routed_model
		FROM requests
		WHERE id LIKE ?
		ORDER BY timestamp DESC
		LIMIT 1
	`

	var req model.RequestLog
	var headersJSON, bodyJSON string
	var promptGradeJSON, responseJSON sql.NullString

	err := s.db.QueryRow(query, "%"+shortID).Scan(
		&req.RequestID,
		&req.Timestamp,
		&req.Method,
		&req.Endpoint,
		&headersJSON,
		&bodyJSON,
		&req.Model,
		&req.UserAgent,
		&req.ContentType,
		&promptGradeJSON,
		&responseJSON,
		&req.OriginalModel,
		&req.RoutedModel,
	)

	if err == sql.ErrNoRows {
		return nil, "", fmt.Errorf("request with ID %s not found", shortID)
	}
	if err != nil {
		return nil, "", fmt.Errorf("failed to query request: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal([]byte(headersJSON), &req.Headers); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal headers: %w", err)
	}

	var body interface{}
	if err := json.Unmarshal([]byte(bodyJSON), &body); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal body: %w", err)
	}
	req.Body = body

	if promptGradeJSON.Valid {
		var grade model.PromptGrade
		if err := json.Unmarshal([]byte(promptGradeJSON.String), &grade); err == nil {
			req.PromptGrade = &grade
		}
	}

	if responseJSON.Valid {
		var resp model.ResponseLog
		if err := json.Unmarshal([]byte(responseJSON.String), &resp); err == nil {
			req.Response = &resp
		}
	}

	return &req, req.RequestID, nil
}

func (s *SQLiteStorageService) GetConfig() *config.StorageConfig {
	return s.config
}

func (s *SQLiteStorageService) GetAllRequests(modelFilter string) ([]*model.RequestLog, error) {
	query := `
		SELECT id, timestamp, method, endpoint, headers, body, model, user_agent, content_type, prompt_grade, response, original_model, routed_model
		FROM requests
	`
	args := []interface{}{}

	if modelFilter != "" && modelFilter != "all" {
		query += " WHERE LOWER(model) LIKE ?"
		args = append(args, "%"+strings.ToLower(modelFilter)+"%")

	}

	query += " ORDER BY timestamp DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query requests: %w", err)
	}
	defer rows.Close()

	var requests []*model.RequestLog
	for rows.Next() {
		var req model.RequestLog
		var headersJSON, bodyJSON string
		var promptGradeJSON, responseJSON sql.NullString

		err := rows.Scan(
			&req.RequestID,
			&req.Timestamp,
			&req.Method,
			&req.Endpoint,
			&headersJSON,
			&bodyJSON,
			&req.Model,
			&req.UserAgent,
			&req.ContentType,
			&promptGradeJSON,
			&responseJSON,
			&req.OriginalModel,
			&req.RoutedModel,
		)
		if err != nil {
			continue
		}

		if err := json.Unmarshal([]byte(headersJSON), &req.Headers); err != nil {
			continue
		}

		var body interface{}
		if err := json.Unmarshal([]byte(bodyJSON), &body); err != nil {
			continue
		}
		req.Body = body

		if promptGradeJSON.Valid {
			var grade model.PromptGrade
			if err := json.Unmarshal([]byte(promptGradeJSON.String), &grade); err == nil {
				req.PromptGrade = &grade
			}
		}

		if responseJSON.Valid {
			var resp model.ResponseLog
			if err := json.Unmarshal([]byte(responseJSON.String), &resp); err == nil {
				req.Response = &resp
			}
		}

		requests = append(requests, &req)
	}

	return requests, nil
}

// GetRequestsSummary returns minimal data for list view - uses indexed columns, no JSON parsing
func (s *SQLiteStorageService) GetRequestsSummary(modelFilter string) ([]*model.RequestSummary, error) {
	query := `
		SELECT id, timestamp, method, endpoint, model, original_model, routed_model,
			   provider, subagent_name, tool_call_count, response_time_ms, first_byte_time_ms,
			   input_tokens, output_tokens, cache_read_tokens, cache_creation_tokens
		FROM requests
	`
	args := []interface{}{}

	if modelFilter != "" && modelFilter != "all" {
		query += " WHERE model LIKE ?"
		args = append(args, "%"+modelFilter+"%")
	}

	query += " ORDER BY timestamp DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query requests: %w", err)
	}
	defer rows.Close()

	var summaries []*model.RequestSummary
	for rows.Next() {
		var sum model.RequestSummary
		var provider, subagentName sql.NullString
		var toolCallCount sql.NullInt64
		var responseTimeMs, firstByteTimeMs sql.NullInt64
		var inputTokens, outputTokens, cacheReadTokens, cacheCreationTokens sql.NullInt64

		err := rows.Scan(
			&sum.RequestID,
			&sum.Timestamp,
			&sum.Method,
			&sum.Endpoint,
			&sum.Model,
			&sum.OriginalModel,
			&sum.RoutedModel,
			&provider,
			&subagentName,
			&toolCallCount,
			&responseTimeMs,
			&firstByteTimeMs,
			&inputTokens,
			&outputTokens,
			&cacheReadTokens,
			&cacheCreationTokens,
		)
		if err != nil {
			continue
		}

		if provider.Valid {
			sum.Provider = provider.String
		}
		if subagentName.Valid {
			sum.SubagentName = subagentName.String
		}
		if toolCallCount.Valid {
			sum.ToolCallCount = int(toolCallCount.Int64)
		}
		if responseTimeMs.Valid {
			sum.ResponseTime = responseTimeMs.Int64
		}
		if firstByteTimeMs.Valid {
			sum.FirstByteTime = firstByteTimeMs.Int64
		}

		// Build usage from indexed columns
		if inputTokens.Valid || outputTokens.Valid {
			sum.Usage = &model.AnthropicUsage{
				InputTokens:             int(inputTokens.Int64),
				OutputTokens:            int(outputTokens.Int64),
				CacheReadInputTokens:    int(cacheReadTokens.Int64),
				CacheCreationInputTokens: int(cacheCreationTokens.Int64),
			}
		}

		summaries = append(summaries, &sum)
	}

	return summaries, nil
}

// GetRequestsSummaryPaginated returns minimal data for list view with pagination - uses indexed columns
func (s *SQLiteStorageService) GetRequestsSummaryPaginated(modelFilter, startTime, endTime string, offset, limit int) ([]*model.RequestSummary, int, error) {
	// First get total count
	countQuery := "SELECT COUNT(*) FROM requests"
	countArgs := []interface{}{}
	whereClauses := []string{}

	if modelFilter != "" && modelFilter != "all" {
		whereClauses = append(whereClauses, "model LIKE ?")
		countArgs = append(countArgs, "%"+modelFilter+"%")
	}

	if startTime != "" && endTime != "" {
		whereClauses = append(whereClauses, "timestamp >= ? AND timestamp < ?")
		countArgs = append(countArgs, startTime, endTime)
	}

	if len(whereClauses) > 0 {
		countQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var total int
	if err := s.db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	// Then get the requested page - no response JSON needed
	query := `
		SELECT id, timestamp, method, endpoint, model, original_model, routed_model,
			   provider, subagent_name, tool_call_count, response_time_ms, first_byte_time_ms,
			   input_tokens, output_tokens, cache_read_tokens, cache_creation_tokens
		FROM requests
	`
	args := []interface{}{}
	queryWhereClauses := []string{}

	if modelFilter != "" && modelFilter != "all" {
		queryWhereClauses = append(queryWhereClauses, "model LIKE ?")
		args = append(args, "%"+modelFilter+"%")
	}

	if startTime != "" && endTime != "" {
		queryWhereClauses = append(queryWhereClauses, "timestamp >= ? AND timestamp < ?")
		args = append(args, startTime, endTime)
	}

	if len(queryWhereClauses) > 0 {
		query += " WHERE " + strings.Join(queryWhereClauses, " AND ")
	}

	query += " ORDER BY timestamp DESC"

	// Only add LIMIT if specified (0 means no limit)
	if limit > 0 {
		query += " LIMIT ? OFFSET ?"
		args = append(args, limit, offset)
	} else if offset > 0 {
		query += " OFFSET ?"
		args = append(args, offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query requests: %w", err)
	}
	defer rows.Close()

	var summaries []*model.RequestSummary
	for rows.Next() {
		var sum model.RequestSummary
		var provider, subagentName sql.NullString
		var toolCallCount sql.NullInt64
		var responseTimeMs, firstByteTimeMs sql.NullInt64
		var inputTokens, outputTokens, cacheReadTokens, cacheCreationTokens sql.NullInt64

		err := rows.Scan(
			&sum.RequestID,
			&sum.Timestamp,
			&sum.Method,
			&sum.Endpoint,
			&sum.Model,
			&sum.OriginalModel,
			&sum.RoutedModel,
			&provider,
			&subagentName,
			&toolCallCount,
			&responseTimeMs,
			&firstByteTimeMs,
			&inputTokens,
			&outputTokens,
			&cacheReadTokens,
			&cacheCreationTokens,
		)
		if err != nil {
			continue
		}

		if provider.Valid {
			sum.Provider = provider.String
		}
		if subagentName.Valid {
			sum.SubagentName = subagentName.String
		}
		if toolCallCount.Valid {
			sum.ToolCallCount = int(toolCallCount.Int64)
		}
		if responseTimeMs.Valid {
			sum.ResponseTime = responseTimeMs.Int64
		}
		if firstByteTimeMs.Valid {
			sum.FirstByteTime = firstByteTimeMs.Int64
		}

		// Build usage from indexed columns
		if inputTokens.Valid || outputTokens.Valid {
			sum.Usage = &model.AnthropicUsage{
				InputTokens:             int(inputTokens.Int64),
				OutputTokens:            int(outputTokens.Int64),
				CacheReadInputTokens:    int(cacheReadTokens.Int64),
				CacheCreationInputTokens: int(cacheCreationTokens.Int64),
			}
		}

		summaries = append(summaries, &sum)
	}

	return summaries, total, nil
}

// GetStats returns aggregated statistics for the dashboard - uses SQL aggregation
func (s *SQLiteStorageService) GetStats(startDate, endDate string) (*model.DashboardStats, error) {
	stats := &model.DashboardStats{
		DailyStats: make([]model.DailyTokens, 0),
	}

	// SQL aggregation - no JSON parsing needed
	query := `
		SELECT
			DATE(timestamp) as date,
			COALESCE(model, 'unknown') as model,
			COUNT(*) as requests,
			SUM(COALESCE(input_tokens, 0) + COALESCE(output_tokens, 0) + COALESCE(cache_read_tokens, 0) + COALESCE(cache_creation_tokens, 0)) as tokens
		FROM requests
		WHERE timestamp >= ? AND timestamp < ?
		GROUP BY date, model
		ORDER BY date
	`

	rows, err := s.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query stats: %w", err)
	}
	defer rows.Close()

	// Aggregate by date (models already grouped by SQL)
	dailyMap := make(map[string]*model.DailyTokens)

	for rows.Next() {
		var date, modelName string
		var requests int
		var tokens int64

		if err := rows.Scan(&date, &modelName, &requests, &tokens); err != nil {
			continue
		}

		if daily, ok := dailyMap[date]; ok {
			daily.Tokens += tokens
			daily.Requests += requests
			daily.Models[modelName] = model.ModelStats{
				Tokens:   tokens,
				Requests: requests,
			}
		} else {
			dailyMap[date] = &model.DailyTokens{
				Date:     date,
				Tokens:   tokens,
				Requests: requests,
				Models: map[string]model.ModelStats{
					modelName: {
						Tokens:   tokens,
						Requests: requests,
					},
				},
			}
		}
	}

	// Convert map to slice
	for _, v := range dailyMap {
		stats.DailyStats = append(stats.DailyStats, *v)
	}

	return stats, nil
}

// GetHourlyStats returns hourly breakdown for a specific time range - uses SQL aggregation
func (s *SQLiteStorageService) GetHourlyStats(startTime, endTime string) (*model.HourlyStatsResponse, error) {
	// SQL aggregation - no JSON parsing needed
	query := `
		SELECT
			CAST(strftime('%H', timestamp) AS INTEGER) as hour,
			COALESCE(model, 'unknown') as model,
			COUNT(*) as requests,
			SUM(COALESCE(input_tokens, 0) + COALESCE(output_tokens, 0) + COALESCE(cache_read_tokens, 0) + COALESCE(cache_creation_tokens, 0)) as tokens,
			SUM(COALESCE(response_time_ms, 0)) as total_response_time,
			SUM(CASE WHEN response_time_ms > 0 THEN 1 ELSE 0 END) as response_count
		FROM requests
		WHERE timestamp >= ? AND timestamp < ?
		GROUP BY hour, model
		ORDER BY hour
	`

	rows, err := s.db.Query(query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query hourly stats: %w", err)
	}
	defer rows.Close()

	hourlyMap := make(map[int]*model.HourlyTokens)
	var totalTokens int64
	var totalRequests int
	var totalResponseTime int64
	var responseCount int

	for rows.Next() {
		var hour, requests, rowResponseCount int
		var modelName string
		var tokens, rowResponseTime int64

		if err := rows.Scan(&hour, &modelName, &requests, &tokens, &rowResponseTime, &rowResponseCount); err != nil {
			continue
		}

		totalTokens += tokens
		totalRequests += requests
		totalResponseTime += rowResponseTime
		responseCount += rowResponseCount

		if hourly, ok := hourlyMap[hour]; ok {
			hourly.Tokens += tokens
			hourly.Requests += requests
			hourly.Models[modelName] = model.ModelStats{
				Tokens:   tokens,
				Requests: requests,
			}
		} else {
			hourlyMap[hour] = &model.HourlyTokens{
				Hour:     hour,
				Tokens:   tokens,
				Requests: requests,
				Models: map[string]model.ModelStats{
					modelName: {
						Tokens:   tokens,
						Requests: requests,
					},
				},
			}
		}
	}

	// Convert map to slice
	hourlyStats := make([]model.HourlyTokens, 0)
	for _, v := range hourlyMap {
		hourlyStats = append(hourlyStats, *v)
	}

	// Calculate average response time
	avgResponseTime := int64(0)
	if responseCount > 0 {
		avgResponseTime = totalResponseTime / int64(responseCount)
	}

	return &model.HourlyStatsResponse{
		HourlyStats:     hourlyStats,
		TodayTokens:     totalTokens,
		TodayRequests:   totalRequests,
		AvgResponseTime: avgResponseTime,
	}, nil
}

// GetModelStats returns model breakdown for a specific time range - uses SQL aggregation
func (s *SQLiteStorageService) GetModelStats(startTime, endTime string) (*model.ModelStatsResponse, error) {
	// SQL aggregation - no JSON parsing needed
	query := `
		SELECT
			COALESCE(model, 'unknown') as model,
			COUNT(*) as requests,
			SUM(COALESCE(input_tokens, 0) + COALESCE(output_tokens, 0) + COALESCE(cache_read_tokens, 0) + COALESCE(cache_creation_tokens, 0)) as tokens
		FROM requests
		WHERE timestamp >= ? AND timestamp < ?
		GROUP BY model
		ORDER BY tokens DESC
	`

	rows, err := s.db.Query(query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query model stats: %w", err)
	}
	defer rows.Close()

	modelStats := make([]model.ModelTokens, 0)

	for rows.Next() {
		var modelName string
		var requests int
		var tokens int64

		if err := rows.Scan(&modelName, &requests, &tokens); err != nil {
			continue
		}

		modelStats = append(modelStats, model.ModelTokens{
			Model:    modelName,
			Tokens:   tokens,
			Requests: requests,
		})
	}

	return &model.ModelStatsResponse{
		ModelStats: modelStats,
	}, nil
}

// GetLatestRequestDate returns the timestamp of the most recent request
func (s *SQLiteStorageService) GetLatestRequestDate() (*time.Time, error) {
	var timestamp string
	err := s.db.QueryRow("SELECT timestamp FROM requests ORDER BY timestamp DESC LIMIT 1").Scan(&timestamp)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query latest request: %w", err)
	}

	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	return &t, nil
}

func (s *SQLiteStorageService) Close() error {
	return s.db.Close()
}

// GetProviderStats returns analytics broken down by provider
func (s *SQLiteStorageService) GetProviderStats(startTime, endTime string) (*model.ProviderStatsResponse, error) {
	query := `
		SELECT
			COALESCE(provider, 'unknown') as provider,
			COUNT(*) as requests,
			SUM(input_tokens) as input_tokens,
			SUM(output_tokens) as output_tokens,
			AVG(response_time_ms) as avg_response_ms
		FROM requests
		WHERE timestamp >= ? AND timestamp < ?
		GROUP BY provider
	`

	rows, err := s.db.Query(query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query provider stats: %w", err)
	}
	defer rows.Close()

	var providers []model.ProviderStats
	for rows.Next() {
		var stat model.ProviderStats
		var inputTokens, outputTokens sql.NullInt64
		var avgResponseMs sql.NullFloat64

		if err := rows.Scan(&stat.Provider, &stat.Requests, &inputTokens, &outputTokens, &avgResponseMs); err != nil {
			continue
		}

		if inputTokens.Valid {
			stat.InputTokens = inputTokens.Int64
		}
		if outputTokens.Valid {
			stat.OutputTokens = outputTokens.Int64
		}
		stat.TotalTokens = stat.InputTokens + stat.OutputTokens
		if avgResponseMs.Valid {
			stat.AvgResponseMs = int64(avgResponseMs.Float64)
		}

		providers = append(providers, stat)
	}

	return &model.ProviderStatsResponse{
		Providers: providers,
		StartTime: startTime,
		EndTime:   endTime,
	}, nil
}

// GetSubagentStats returns analytics broken down by subagent
func (s *SQLiteStorageService) GetSubagentStats(startTime, endTime string) (*model.SubagentStatsResponse, error) {
	query := `
		SELECT
			COALESCE(subagent_name, '') as subagent_name,
			COALESCE(provider, 'unknown') as provider,
			COALESCE(routed_model, model) as target_model,
			COUNT(*) as requests,
			SUM(input_tokens) as input_tokens,
			SUM(output_tokens) as output_tokens,
			AVG(response_time_ms) as avg_response_ms
		FROM requests
		WHERE timestamp >= ? AND timestamp < ?
		  AND subagent_name IS NOT NULL AND subagent_name != ''
		GROUP BY subagent_name, provider, target_model
	`

	rows, err := s.db.Query(query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query subagent stats: %w", err)
	}
	defer rows.Close()

	var subagents []model.SubagentStats
	for rows.Next() {
		var stat model.SubagentStats
		var inputTokens, outputTokens sql.NullInt64
		var avgResponseMs sql.NullFloat64

		if err := rows.Scan(&stat.SubagentName, &stat.Provider, &stat.TargetModel, &stat.Requests, &inputTokens, &outputTokens, &avgResponseMs); err != nil {
			continue
		}

		if inputTokens.Valid {
			stat.InputTokens = inputTokens.Int64
		}
		if outputTokens.Valid {
			stat.OutputTokens = outputTokens.Int64
		}
		stat.TotalTokens = stat.InputTokens + stat.OutputTokens
		if avgResponseMs.Valid {
			stat.AvgResponseMs = int64(avgResponseMs.Float64)
		}

		subagents = append(subagents, stat)
	}

	return &model.SubagentStatsResponse{
		Subagents: subagents,
		StartTime: startTime,
		EndTime:   endTime,
	}, nil
}

// GetToolStats returns analytics broken down by tool usage
func (s *SQLiteStorageService) GetToolStats(startTime, endTime string) (*model.ToolStatsResponse, error) {
	query := `
		SELECT tools_used, tool_call_count
		FROM requests
		WHERE timestamp >= ? AND timestamp < ?
		  AND tools_used IS NOT NULL AND tools_used != '[]' AND tools_used != 'null'
	`

	rows, err := s.db.Query(query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query tool stats: %w", err)
	}
	defer rows.Close()

	toolUsageCount := make(map[string]int)  // How many requests included this tool
	toolCallCount := make(map[string]int)   // Total calls across all requests

	for rows.Next() {
		var toolsUsedJSON string
		var callCount int

		if err := rows.Scan(&toolsUsedJSON, &callCount); err != nil {
			continue
		}

		var tools []string
		if err := json.Unmarshal([]byte(toolsUsedJSON), &tools); err != nil {
			continue
		}

		// Count each tool's presence in this request
		for _, tool := range tools {
			if tool != "" {
				toolUsageCount[tool]++
			}
		}
	}

	var toolStats []model.ToolStats
	for toolName, usageCount := range toolUsageCount {
		stat := model.ToolStats{
			ToolName:   toolName,
			UsageCount: usageCount,
			CallCount:  toolCallCount[toolName],
		}
		if usageCount > 0 {
			stat.AvgCallsPerRequest = float64(toolCallCount[toolName]) / float64(usageCount)
		}
		toolStats = append(toolStats, stat)
	}

	return &model.ToolStatsResponse{
		Tools:     toolStats,
		StartTime: startTime,
		EndTime:   endTime,
	}, nil
}

// GetPerformanceStats returns response time analytics by provider/model
func (s *SQLiteStorageService) GetPerformanceStats(startTime, endTime string) (*model.PerformanceStatsResponse, error) {
	query := `
		SELECT
			COALESCE(provider, 'unknown') as provider,
			COALESCE(model, 'unknown') as model,
			response_time_ms,
			first_byte_time_ms
		FROM requests
		WHERE timestamp >= ? AND timestamp < ?
		  AND response_time_ms > 0
		ORDER BY provider, model
	`

	rows, err := s.db.Query(query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query performance stats: %w", err)
	}
	defer rows.Close()

	// Collect response times by provider+model
	type key struct {
		provider string
		model    string
	}
	responseTimes := make(map[key][]int64)
	firstByteTimes := make(map[key][]int64)

	for rows.Next() {
		var provider, modelName string
		var responseTimeMs, firstByteTimeMs int64

		if err := rows.Scan(&provider, &modelName, &responseTimeMs, &firstByteTimeMs); err != nil {
			continue
		}

		k := key{provider: provider, model: modelName}
		responseTimes[k] = append(responseTimes[k], responseTimeMs)
		if firstByteTimeMs > 0 {
			firstByteTimes[k] = append(firstByteTimes[k], firstByteTimeMs)
		}
	}

	var stats []model.PerformanceStats
	for k, times := range responseTimes {
		if len(times) == 0 {
			continue
		}

		// Sort for percentile calculation
		sortedTimes := make([]int64, len(times))
		copy(sortedTimes, times)
		sortInt64Slice(sortedTimes)

		stat := model.PerformanceStats{
			Provider:      k.provider,
			Model:         k.model,
			RequestCount:  len(times),
			AvgResponseMs: avgInt64(times),
			P50ResponseMs: percentileInt64(sortedTimes, 50),
			P95ResponseMs: percentileInt64(sortedTimes, 95),
			P99ResponseMs: percentileInt64(sortedTimes, 99),
		}

		if fbt, exists := firstByteTimes[k]; exists && len(fbt) > 0 {
			stat.AvgFirstByteMs = avgInt64(fbt)
		}

		stats = append(stats, stat)
	}

	return &model.PerformanceStatsResponse{
		Stats:     stats,
		StartTime: startTime,
		EndTime:   endTime,
	}, nil
}

// Helper functions for statistics
func sortInt64Slice(s []int64) {
	for i := 0; i < len(s)-1; i++ {
		for j := i + 1; j < len(s); j++ {
			if s[i] > s[j] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

func avgInt64(s []int64) int64 {
	if len(s) == 0 {
		return 0
	}
	var sum int64
	for _, v := range s {
		sum += v
	}
	return sum / int64(len(s))
}

func percentileInt64(sorted []int64, p int) int64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := (len(sorted) * p) / 100
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

// SearchConversations performs FTS5 search on conversation content
func (s *SQLiteStorageService) SearchConversations(opts model.SearchOptions) (*model.SearchResults, error) {
	// Build FTS5 query - convert user input to OR logic for multi-term searches
	terms := strings.Fields(opts.Query)
	if len(terms) == 0 {
		return &model.SearchResults{
			Query:   opts.Query,
			Results: []*model.ConversationMatch{},
			Total:   0,
			Limit:   opts.Limit,
			Offset:  opts.Offset,
		}, nil
	}

	// Escape FTS5 special characters and build OR query
	var escapedTerms []string
	for _, term := range terms {
		// Escape double quotes by doubling them
		escaped := strings.ReplaceAll(term, `"`, `""`)
		escapedTerms = append(escapedTerms, fmt.Sprintf(`"%s"`, escaped))
	}
	ftsQuery := strings.Join(escapedTerms, " OR ")

	// Build the main query
	query := `
		SELECT
			c.id AS conversation_id,
			c.project_name,
			c.project_path,
			c.end_time AS last_activity,
			COUNT(f.rowid) AS match_count
		FROM conversations_fts f
		JOIN conversations c ON f.conversation_id = c.id
		WHERE conversations_fts MATCH ?
	`
	args := []interface{}{ftsQuery}

	// Add project filter if specified
	if opts.ProjectPath != "" {
		query += " AND c.project_path = ?"
		args = append(args, opts.ProjectPath)
	}

	query += `
		GROUP BY c.id
		ORDER BY match_count DESC, c.end_time DESC
	`

	// Get total count first
	countQuery := strings.Replace(query,
		"SELECT\n\t\t\tc.id AS conversation_id,\n\t\t\tc.project_name,\n\t\t\tc.project_path,\n\t\t\tc.end_time AS last_activity,\n\t\t\tCOUNT(f.rowid) AS match_count",
		"SELECT COUNT(DISTINCT c.id)",
		1)

	// Remove ORDER BY and GROUP BY for count query
	countQuery = strings.Split(countQuery, "GROUP BY")[0]

	var total int
	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Add pagination
	query += " LIMIT ? OFFSET ?"
	args = append(args, opts.Limit, opts.Offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query conversations: %w", err)
	}
	defer rows.Close()

	results := make([]*model.ConversationMatch, 0)
	for rows.Next() {
		var match model.ConversationMatch
		var lastActivity sql.NullString

		if err := rows.Scan(
			&match.ConversationID,
			&match.ProjectName,
			&match.ProjectPath,
			&lastActivity,
			&match.MatchCount,
		); err != nil {
			continue
		}

		if lastActivity.Valid {
			if t, err := time.Parse(time.RFC3339, lastActivity.String); err == nil {
				match.LastActivity = t
			}
		}

		results = append(results, &match)
	}

	return &model.SearchResults{
		Query:   opts.Query,
		Results: results,
		Total:   total,
		Limit:   opts.Limit,
		Offset:  opts.Offset,
	}, nil
}

// GetIndexedConversations returns conversations from the database index - very fast
func (s *SQLiteStorageService) GetIndexedConversations(limit int) ([]*model.IndexedConversation, error) {
	query := `
		SELECT id, project_path, project_name, start_time, end_time, message_count
		FROM conversations
		WHERE message_count > 0
		ORDER BY end_time DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	log.Printf("🔍 GetIndexedConversations: limit=%d, final query: %s", limit, query)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query indexed conversations: %w", err)
	}
	defer rows.Close()

	var conversations []*model.IndexedConversation
	for rows.Next() {
		var conv model.IndexedConversation
		var startTime, endTime sql.NullString

		if err := rows.Scan(
			&conv.ID,
			&conv.ProjectPath,
			&conv.ProjectName,
			&startTime,
			&endTime,
			&conv.MessageCount,
		); err != nil {
			continue
		}

		if startTime.Valid {
			if t, err := time.Parse(time.RFC3339, startTime.String); err == nil {
				conv.StartTime = t
			}
		}
		if endTime.Valid {
			if t, err := time.Parse(time.RFC3339, endTime.String); err == nil {
				conv.EndTime = t
			}
		}

		conversations = append(conversations, &conv)
	}

	return conversations, nil
}

// GetConversationFilePath returns the file path and project path for a conversation by ID
func (s *SQLiteStorageService) GetConversationFilePath(conversationID string) (string, string, error) {
	var filePath, projectPath string
	err := s.db.QueryRow(
		"SELECT file_path, project_path FROM conversations WHERE id = ?",
		conversationID,
	).Scan(&filePath, &projectPath)
	if err != nil {
		return "", "", fmt.Errorf("conversation not found: %w", err)
	}
	return filePath, projectPath, nil
}

// GetConversationMessages returns messages for a conversation from the database
func (s *SQLiteStorageService) GetConversationMessages(conversationID string, limit, offset int) ([]*model.DBConversationMessage, int, error) {
	// Get total count
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM conversation_messages WHERE conversation_id = ?", conversationID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count messages: %w", err)
	}

	// Set default limit
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT uuid, conversation_id, parent_uuid, type, role, timestamp,
		       cwd, git_branch, session_id, agent_id, is_sidechain,
		       request_id, model, input_tokens, output_tokens,
		       cache_read_tokens, cache_creation_tokens, content_json
		FROM conversation_messages
		WHERE conversation_id = ?
		ORDER BY timestamp ASC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, conversationID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []*model.DBConversationMessage
	for rows.Next() {
		var msg model.DBConversationMessage
		var parentUUID, role, cwd, gitBranch, sessionID, agentID, requestID, modelName sql.NullString
		var contentJSON sql.NullString
		var timestampStr string

		err := rows.Scan(
			&msg.UUID,
			&msg.ConversationID,
			&parentUUID,
			&msg.Type,
			&role,
			&timestampStr,
			&cwd,
			&gitBranch,
			&sessionID,
			&agentID,
			&msg.IsSidechain,
			&requestID,
			&modelName,
			&msg.InputTokens,
			&msg.OutputTokens,
			&msg.CacheReadTokens,
			&msg.CacheCreationTokens,
			&contentJSON,
		)
		if err != nil {
			log.Printf("⚠️ Error scanning message row: %v", err)
			continue
		}

		// Parse timestamp
		if t, err := time.Parse(time.RFC3339, timestampStr); err == nil {
			msg.Timestamp = t
		}

		// Handle nullable fields
		if parentUUID.Valid {
			msg.ParentUUID = &parentUUID.String
		}
		if role.Valid {
			msg.Role = role.String
		}
		if cwd.Valid {
			msg.CWD = cwd.String
		}
		if gitBranch.Valid {
			msg.GitBranch = gitBranch.String
		}
		if sessionID.Valid {
			msg.SessionID = sessionID.String
		}
		if agentID.Valid {
			msg.AgentID = agentID.String
		}
		if requestID.Valid {
			msg.RequestID = requestID.String
		}
		if modelName.Valid {
			msg.Model = modelName.String
		}
		if contentJSON.Valid {
			msg.Content = json.RawMessage(contentJSON.String)
		}

		messages = append(messages, &msg)
	}

	return messages, total, nil
}

// GetConversationMessagesWithSubagents returns messages including subagent messages merged by timestamp
func (s *SQLiteStorageService) GetConversationMessagesWithSubagents(conversationID string, limit, offset int) ([]*model.DBConversationMessage, int, error) {
	// Get messages from parent conversation + all subagent conversations
	// Subagent messages have session_id matching the parent conversation_id

	// Set default limit
	if limit <= 0 {
		limit = 100
	}

	// First get the count of all messages (parent + subagents)
	var total int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM conversation_messages
		WHERE conversation_id = ? OR session_id = ?
	`, conversationID, conversationID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count messages: %w", err)
	}

	// Get all messages from parent and subagents, ordered by timestamp
	query := `
		SELECT uuid, conversation_id, parent_uuid, type, role, timestamp,
		       cwd, git_branch, session_id, agent_id, is_sidechain,
		       request_id, model, input_tokens, output_tokens,
		       cache_read_tokens, cache_creation_tokens, content_json
		FROM conversation_messages
		WHERE conversation_id = ? OR session_id = ?
		ORDER BY timestamp ASC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, conversationID, conversationID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []*model.DBConversationMessage
	for rows.Next() {
		var msg model.DBConversationMessage
		var parentUUID, role, cwd, gitBranch, sessionID, agentID, requestID, modelName sql.NullString
		var contentJSON sql.NullString
		var timestampStr string

		err := rows.Scan(
			&msg.UUID,
			&msg.ConversationID,
			&parentUUID,
			&msg.Type,
			&role,
			&timestampStr,
			&cwd,
			&gitBranch,
			&sessionID,
			&agentID,
			&msg.IsSidechain,
			&requestID,
			&modelName,
			&msg.InputTokens,
			&msg.OutputTokens,
			&msg.CacheReadTokens,
			&msg.CacheCreationTokens,
			&contentJSON,
		)
		if err != nil {
			log.Printf("⚠️ Error scanning message row: %v", err)
			continue
		}

		// Parse timestamp
		if t, err := time.Parse(time.RFC3339, timestampStr); err == nil {
			msg.Timestamp = t
		}

		// Handle nullable fields
		if parentUUID.Valid {
			msg.ParentUUID = &parentUUID.String
		}
		if role.Valid {
			msg.Role = role.String
		}
		if cwd.Valid {
			msg.CWD = cwd.String
		}
		if gitBranch.Valid {
			msg.GitBranch = gitBranch.String
		}
		if sessionID.Valid {
			msg.SessionID = sessionID.String
		}
		if agentID.Valid {
			msg.AgentID = agentID.String
		}
		if requestID.Valid {
			msg.RequestID = requestID.String
		}
		if modelName.Valid {
			msg.Model = modelName.String
		}
		if contentJSON.Valid {
			msg.Content = json.RawMessage(contentJSON.String)
		}

		messages = append(messages, &msg)
	}

	return messages, total, nil
}

// ReindexConversations triggers a full re-index by clearing indexed_at timestamps
func (s *SQLiteStorageService) ReindexConversations() error {
	_, err := s.db.Exec("UPDATE conversations SET indexed_at = NULL")
	if err != nil {
		return fmt.Errorf("failed to clear indexed_at: %w", err)
	}
	log.Println("🔄 Cleared indexed_at timestamps - conversations will be re-indexed")
	return nil
}
