package service

import (
	"time"

	"github.com/seifghazi/claude-code-monitor/internal/config"
	"github.com/seifghazi/claude-code-monitor/internal/model"
)

type StorageService interface {
	SaveRequest(request *model.RequestLog) (string, error)
	GetRequests(page, limit int) ([]model.RequestLog, int, error)
	ClearRequests() (int, error)
	UpdateRequestWithGrading(requestID string, grade *model.PromptGrade) error
	UpdateRequestWithResponse(request *model.RequestLog) error
	EnsureDirectoryExists() error
	GetRequestByShortID(shortID string) (*model.RequestLog, string, error)
	GetConfig() *config.StorageConfig
	GetAllRequests(modelFilter string) ([]*model.RequestLog, error)
	GetRequestsSummary(modelFilter string) ([]*model.RequestSummary, error)
	GetRequestsSummaryPaginated(modelFilter, startTime, endTime string, offset, limit int) ([]*model.RequestSummary, int, error)
	GetStats(startDate, endDate string) (*model.DashboardStats, error)
	GetHourlyStats(startTime, endTime string) (*model.HourlyStatsResponse, error)
	GetModelStats(startTime, endTime string) (*model.ModelStatsResponse, error)
	GetLatestRequestDate() (*time.Time, error)
	Close() error

	// New analytics endpoints
	GetProviderStats(startTime, endTime string) (*model.ProviderStatsResponse, error)
	GetSubagentStats(startTime, endTime string) (*model.SubagentStatsResponse, error)
	GetToolStats(startTime, endTime string) (*model.ToolStatsResponse, error)
	GetPerformanceStats(startTime, endTime string) (*model.PerformanceStatsResponse, error)

	// Conversation search
	SearchConversations(opts model.SearchOptions) (*model.SearchResults, error)

	// Indexed conversations - fast database lookup
	GetIndexedConversations(limit int) ([]*model.IndexedConversation, error)
}
