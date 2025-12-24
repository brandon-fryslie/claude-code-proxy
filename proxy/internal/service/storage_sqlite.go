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

type sqliteStorageService struct {
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

	service := &sqliteStorageService{
		db:     db,
		config: cfg,
	}

	if err := service.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return service, nil
}

func (s *sqliteStorageService) createTables() error {
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

	return nil
}

func (s *sqliteStorageService) runMigrations() error {
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

func (s *sqliteStorageService) SaveRequest(request *model.RequestLog) (string, error) {
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

func (s *sqliteStorageService) GetRequests(page, limit int) ([]model.RequestLog, int, error) {
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

func (s *sqliteStorageService) ClearRequests() (int, error) {
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

func (s *sqliteStorageService) UpdateRequestWithGrading(requestID string, grade *model.PromptGrade) error {
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

func (s *sqliteStorageService) UpdateRequestWithResponse(request *model.RequestLog) error {
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

func (s *sqliteStorageService) EnsureDirectoryExists() error {
	// No directory needed for SQLite
	return nil
}

func (s *sqliteStorageService) GetRequestByShortID(shortID string) (*model.RequestLog, string, error) {
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

func (s *sqliteStorageService) GetConfig() *config.StorageConfig {
	return s.config
}

func (s *sqliteStorageService) GetAllRequests(modelFilter string) ([]*model.RequestLog, error) {
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

// GetRequestsSummary returns minimal data for list view - no body/headers, only usage from response
func (s *sqliteStorageService) GetRequestsSummary(modelFilter string) ([]*model.RequestSummary, error) {
	query := `
		SELECT id, timestamp, method, endpoint, model, original_model, routed_model, response
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

	var summaries []*model.RequestSummary
	for rows.Next() {
		var s model.RequestSummary
		var responseJSON sql.NullString

		err := rows.Scan(
			&s.RequestID,
			&s.Timestamp,
			&s.Method,
			&s.Endpoint,
			&s.Model,
			&s.OriginalModel,
			&s.RoutedModel,
			&responseJSON,
		)
		if err != nil {
			continue
		}

		// Only parse response to extract usage and status
		if responseJSON.Valid {
			var resp model.ResponseLog
			if err := json.Unmarshal([]byte(responseJSON.String), &resp); err == nil {
				s.StatusCode = resp.StatusCode
				s.ResponseTime = resp.ResponseTime

				// Extract usage from response body
				if resp.Body != nil {
					var respBody struct {
						Usage *model.AnthropicUsage `json:"usage"`
					}
					if err := json.Unmarshal(resp.Body, &respBody); err == nil && respBody.Usage != nil {
						s.Usage = respBody.Usage
					}
				}
			}
		}

		summaries = append(summaries, &s)
	}

	return summaries, nil
}

// GetRequestsSummaryPaginated returns minimal data for list view with pagination - super fast!
func (s *sqliteStorageService) GetRequestsSummaryPaginated(modelFilter, startTime, endTime string, offset, limit int) ([]*model.RequestSummary, int, error) {
	// First get total count
	countQuery := "SELECT COUNT(*) FROM requests"
	countArgs := []interface{}{}
	whereClauses := []string{}

	if modelFilter != "" && modelFilter != "all" {
		whereClauses = append(whereClauses, "LOWER(model) LIKE ?")
		countArgs = append(countArgs, "%"+strings.ToLower(modelFilter)+"%")
	}

	if startTime != "" && endTime != "" {
		whereClauses = append(whereClauses, "datetime(timestamp) >= datetime(?) AND datetime(timestamp) <= datetime(?)")
		countArgs = append(countArgs, startTime, endTime)
	}

	if len(whereClauses) > 0 {
		countQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var total int
	if err := s.db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	// Then get the requested page
	query := `
		SELECT id, timestamp, method, endpoint, model, original_model, routed_model, response
		FROM requests
	`
	args := []interface{}{}
	queryWhereClauses := []string{}

	if modelFilter != "" && modelFilter != "all" {
		queryWhereClauses = append(queryWhereClauses, "LOWER(model) LIKE ?")
		args = append(args, "%"+strings.ToLower(modelFilter)+"%")
	}

	if startTime != "" && endTime != "" {
		queryWhereClauses = append(queryWhereClauses, "datetime(timestamp) >= datetime(?) AND datetime(timestamp) <= datetime(?)")
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
		var s model.RequestSummary
		var responseJSON sql.NullString

		err := rows.Scan(
			&s.RequestID,
			&s.Timestamp,
			&s.Method,
			&s.Endpoint,
			&s.Model,
			&s.OriginalModel,
			&s.RoutedModel,
			&responseJSON,
		)
		if err != nil {
			continue
		}

		// Only parse response to extract usage and status
		if responseJSON.Valid {
			var resp model.ResponseLog
			if err := json.Unmarshal([]byte(responseJSON.String), &resp); err == nil {
				s.StatusCode = resp.StatusCode
				s.ResponseTime = resp.ResponseTime

				// Extract usage from response body
				if resp.Body != nil {
					var respBody struct {
						Usage *model.AnthropicUsage `json:"usage"`
					}
					if err := json.Unmarshal(resp.Body, &respBody); err == nil && respBody.Usage != nil {
						s.Usage = respBody.Usage
					}
				}
			}
		}

		summaries = append(summaries, &s)
	}

	log.Printf("📊 GetRequestsSummaryPaginated: returned %d requests (total: %d, limit: %d, offset: %d)", len(summaries), total, limit, offset)
	return summaries, total, nil
}

// GetStats returns aggregated statistics for the dashboard - lightning fast!
func (s *sqliteStorageService) GetStats(startDate, endDate string) (*model.DashboardStats, error) {
	stats := &model.DashboardStats{
		DailyStats: make([]model.DailyTokens, 0),
	}

	// Query each request individually to process all responses
	query := `
		SELECT timestamp, COALESCE(model, 'unknown') as model, response
		FROM requests
		WHERE datetime(timestamp) >= datetime(?) AND datetime(timestamp) <= datetime(?)
		ORDER BY timestamp
	`

	rows, err := s.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query stats: %w", err)
	}
	defer rows.Close()

	// Aggregate data in memory
	dailyMap := make(map[string]*model.DailyTokens)

	for rows.Next() {
		var timestamp, modelName, responseJSON string

		if err := rows.Scan(&timestamp, &modelName, &responseJSON); err != nil {
			continue
		}

		// Extract date from timestamp (format: 2025-11-28T13:03:29-08:00)
		date := strings.Split(timestamp, "T")[0]

		// Parse response to get usage
		var resp model.ResponseLog
		if err := json.Unmarshal([]byte(responseJSON), &resp); err != nil {
			continue
		}

		var usage *model.AnthropicUsage
		if resp.Body != nil {
			var respBody struct {
				Usage *model.AnthropicUsage `json:"usage"`
			}
			if err := json.Unmarshal(resp.Body, &respBody); err == nil {
				usage = respBody.Usage
			}
		}

		tokens := int64(0)
		if usage != nil {
			tokens = int64(
				usage.InputTokens +
					usage.OutputTokens +
					usage.CacheReadInputTokens +
					usage.CacheCreationInputTokens)
		}

		// Daily aggregation
		if daily, ok := dailyMap[date]; ok {
			daily.Tokens += tokens
			daily.Requests++
			// Update per-model stats
			if daily.Models == nil {
				daily.Models = make(map[string]model.ModelStats)
			}
			if modelStat, ok := daily.Models[modelName]; ok {
				modelStat.Tokens += tokens
				modelStat.Requests++
				daily.Models[modelName] = modelStat
			} else {
				daily.Models[modelName] = model.ModelStats{
					Tokens:   tokens,
					Requests: 1,
				}
			}
		} else {
			dailyMap[date] = &model.DailyTokens{
				Date:     date,
				Tokens:   tokens,
				Requests: 1,
				Models: map[string]model.ModelStats{
					modelName: {
						Tokens:   tokens,
						Requests: 1,
					},
				},
			}
		}

	}

	// Convert maps to slices
	for _, v := range dailyMap {
		stats.DailyStats = append(stats.DailyStats, *v)
	}

	return stats, nil
}

// GetHourlyStats returns hourly breakdown for a specific time range
func (s *sqliteStorageService) GetHourlyStats(startTime, endTime string) (*model.HourlyStatsResponse, error) {
	query := `
		SELECT timestamp, COALESCE(model, 'unknown') as model, response
		FROM requests
		WHERE datetime(timestamp) >= datetime(?) AND datetime(timestamp) <= datetime(?)
		ORDER BY timestamp
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
		var timestamp, modelName, responseJSON string

		if err := rows.Scan(&timestamp, &modelName, &responseJSON); err != nil {
			continue
		}

		// Extract hour from timestamp
		hour := 0
		if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
			hour = t.Hour()
		}

		// Parse response to get usage and response time
		var resp model.ResponseLog
		if err := json.Unmarshal([]byte(responseJSON), &resp); err != nil {
			continue
		}

		var usage *model.AnthropicUsage
		if resp.Body != nil {
			var respBody struct {
				Usage *model.AnthropicUsage `json:"usage"`
			}
			if err := json.Unmarshal(resp.Body, &respBody); err == nil {
				usage = respBody.Usage
			}
		}

		tokens := int64(0)
		if usage != nil {
			tokens = int64(
				usage.InputTokens +
					usage.OutputTokens +
					usage.CacheReadInputTokens +
					usage.CacheCreationInputTokens)
		}

		totalTokens += tokens
		totalRequests++

		// Track response time
		if resp.ResponseTime > 0 {
			totalResponseTime += resp.ResponseTime
			responseCount++
		}

		// Hourly aggregation
		if hourly, ok := hourlyMap[hour]; ok {
			hourly.Tokens += tokens
			hourly.Requests++
			// Update per-model stats
			if hourly.Models == nil {
				hourly.Models = make(map[string]model.ModelStats)
			}
			if modelStat, ok := hourly.Models[modelName]; ok {
				modelStat.Tokens += tokens
				modelStat.Requests++
				hourly.Models[modelName] = modelStat
			} else {
				hourly.Models[modelName] = model.ModelStats{
					Tokens:   tokens,
					Requests: 1,
				}
			}
		} else {
			hourlyMap[hour] = &model.HourlyTokens{
				Hour:     hour,
				Tokens:   tokens,
				Requests: 1,
				Models: map[string]model.ModelStats{
					modelName: {
						Tokens:   tokens,
						Requests: 1,
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

// GetModelStats returns model breakdown for a specific time range
func (s *sqliteStorageService) GetModelStats(startTime, endTime string) (*model.ModelStatsResponse, error) {
	query := `
		SELECT timestamp, COALESCE(model, 'unknown') as model, response
		FROM requests
		WHERE datetime(timestamp) >= datetime(?) AND datetime(timestamp) <= datetime(?)
		ORDER BY timestamp
	`

	rows, err := s.db.Query(query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query model stats: %w", err)
	}
	defer rows.Close()

	modelMap := make(map[string]*model.ModelTokens)

	for rows.Next() {
		var timestamp, modelName, responseJSON string

		if err := rows.Scan(&timestamp, &modelName, &responseJSON); err != nil {
			continue
		}

		// Parse response to get usage
		var resp model.ResponseLog
		if err := json.Unmarshal([]byte(responseJSON), &resp); err != nil {
			continue
		}

		var usage *model.AnthropicUsage
		if resp.Body != nil {
			var respBody struct {
				Usage *model.AnthropicUsage `json:"usage"`
			}
			if err := json.Unmarshal(resp.Body, &respBody); err == nil {
				usage = respBody.Usage
			}
		}

		tokens := int64(0)
		if usage != nil {
			tokens = int64(
				usage.InputTokens +
					usage.OutputTokens +
					usage.CacheReadInputTokens +
					usage.CacheCreationInputTokens)
		}

		// Model aggregation
		if modelStat, ok := modelMap[modelName]; ok {
			modelStat.Tokens += tokens
			modelStat.Requests++
		} else {
			modelMap[modelName] = &model.ModelTokens{
				Model:    modelName,
				Tokens:   tokens,
				Requests: 1,
			}
		}
	}

	// Convert map to slice
	modelStats := make([]model.ModelTokens, 0)
	for _, v := range modelMap {
		modelStats = append(modelStats, *v)
	}

	return &model.ModelStatsResponse{
		ModelStats: modelStats,
	}, nil
}

// GetLatestRequestDate returns the timestamp of the most recent request
func (s *sqliteStorageService) GetLatestRequestDate() (*time.Time, error) {
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

func (s *sqliteStorageService) Close() error {
	return s.db.Close()
}

// GetProviderStats returns analytics broken down by provider
func (s *sqliteStorageService) GetProviderStats(startTime, endTime string) (*model.ProviderStatsResponse, error) {
	query := `
		SELECT
			COALESCE(provider, 'unknown') as provider,
			COUNT(*) as requests,
			SUM(input_tokens) as input_tokens,
			SUM(output_tokens) as output_tokens,
			AVG(response_time_ms) as avg_response_ms,
			response
		FROM requests
		WHERE datetime(timestamp) >= datetime(?) AND datetime(timestamp) <= datetime(?)
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
		var responseJSON sql.NullString

		if err := rows.Scan(&stat.Provider, &stat.Requests, &inputTokens, &outputTokens, &avgResponseMs, &responseJSON); err != nil {
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
func (s *sqliteStorageService) GetSubagentStats(startTime, endTime string) (*model.SubagentStatsResponse, error) {
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
		WHERE datetime(timestamp) >= datetime(?) AND datetime(timestamp) <= datetime(?)
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
func (s *sqliteStorageService) GetToolStats(startTime, endTime string) (*model.ToolStatsResponse, error) {
	query := `
		SELECT tools_used, tool_call_count
		FROM requests
		WHERE datetime(timestamp) >= datetime(?) AND datetime(timestamp) <= datetime(?)
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
func (s *sqliteStorageService) GetPerformanceStats(startTime, endTime string) (*model.PerformanceStatsResponse, error) {
	query := `
		SELECT
			COALESCE(provider, 'unknown') as provider,
			COALESCE(model, 'unknown') as model,
			response_time_ms,
			first_byte_time_ms
		FROM requests
		WHERE datetime(timestamp) >= datetime(?) AND datetime(timestamp) <= datetime(?)
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
