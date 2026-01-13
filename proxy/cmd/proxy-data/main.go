package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/seifghazi/claude-code-monitor/internal/config"
	"github.com/seifghazi/claude-code-monitor/internal/handler"
	"github.com/seifghazi/claude-code-monitor/internal/middleware"
	"github.com/seifghazi/claude-code-monitor/internal/service"
)

func main() {
	logger := log.New(os.Stdout, "proxy-data: ", log.LstdFlags|log.Lshortfile)

	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Use SQLite storage (full read/write access)
	storageService, err := service.NewSQLiteStorageService(&cfg.Storage)
	if err != nil {
		logger.Fatalf("Failed to initialize SQLite storage: %v", err)
	}
	logger.Println("SQLite database ready")

	// Start conversation indexer (fail-fast on error)
	sqliteStorage, ok := storageService.(*service.SQLiteStorageService)
	if !ok {
		logger.Fatalf("Storage service must be SQLite for indexer support")
	}

	indexer, err := service.NewConversationIndexer(sqliteStorage)
	if err != nil {
		logger.Fatalf("Failed to create conversation indexer: %v", err)
	}

	if err := indexer.Start(); err != nil {
		logger.Fatalf("Failed to start conversation indexer: %v", err)
	}
	defer indexer.Stop()
	logger.Println("Conversation indexer started")

	// Create data handler (full dependencies)
	h := handler.NewDataHandler(storageService, logger, cfg)
	h.SetIndexer(indexer)

	r := mux.NewRouter()

	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"*"}),
	)

	r.Use(middleware.Logging)

	// Health check
	r.HandleFunc("/health", h.Health).Methods("GET")

	// UI route
	r.HandleFunc("/", h.UI).Methods("GET")
	r.HandleFunc("/ui", h.UI).Methods("GET")

	// V1 API - Request endpoints
	r.HandleFunc("/api/requests", h.GetRequests).Methods("GET")
	r.HandleFunc("/api/requests/summary", h.GetRequestsSummary).Methods("GET")
	r.HandleFunc("/api/requests/latest-date", h.GetLatestRequestDate).Methods("GET")
	r.HandleFunc("/api/requests/{id}", h.GetRequestByID).Methods("GET")
	r.HandleFunc("/api/requests", h.DeleteRequests).Methods("DELETE")

	// V1 API - Stats endpoints
	r.HandleFunc("/api/stats", h.GetStats).Methods("GET")
	r.HandleFunc("/api/stats/hourly", h.GetHourlyStats).Methods("GET")
	r.HandleFunc("/api/stats/models", h.GetModelStats).Methods("GET")
	r.HandleFunc("/api/stats/providers", h.GetProviderStats).Methods("GET")
	r.HandleFunc("/api/stats/subagents", h.GetSubagentStats).Methods("GET")
	r.HandleFunc("/api/stats/tools", h.GetToolStats).Methods("GET")
	r.HandleFunc("/api/stats/performance", h.GetPerformanceStats).Methods("GET")

	// V1 API - Conversation endpoints (specific routes before parameterized)
	r.HandleFunc("/api/conversations", h.GetConversations).Methods("GET")
	r.HandleFunc("/api/conversations/search", h.SearchConversations).Methods("GET")
	r.HandleFunc("/api/conversations/project", h.GetConversationsByProject).Methods("GET")
	r.HandleFunc("/api/conversations/{id}", h.GetConversationByID).Methods("GET")

	// V2 API - cleaner response format for new dashboard
	r.HandleFunc("/api/v2/requests/summary", h.GetRequestsSummaryV2).Methods("GET")
	r.HandleFunc("/api/v2/requests/{id}", h.GetRequestByIDV2).Methods("GET")
	r.HandleFunc("/api/v2/conversations", h.GetConversationsV2).Methods("GET")
	r.HandleFunc("/api/v2/conversations/search", h.SearchConversations).Methods("GET")
	r.HandleFunc("/api/v2/conversations/reindex", h.ReindexConversationsV2).Methods("POST")
	// Specific routes must be registered BEFORE generic {id} routes
	r.HandleFunc("/api/v2/conversations/{id}/messages", h.GetConversationMessagesV2).Methods("GET")
	r.HandleFunc("/api/v2/conversations/{id}", h.GetConversationByIDV2).Methods("GET")
	r.HandleFunc("/api/v2/stats", h.GetWeeklyStatsV2).Methods("GET")
	r.HandleFunc("/api/v2/stats/hourly", h.GetHourlyStatsV2).Methods("GET")
	r.HandleFunc("/api/v2/stats/models", h.GetModelStatsV2).Methods("GET")
	r.HandleFunc("/api/v2/stats/providers", h.GetProvidersV2).Methods("GET")
	r.HandleFunc("/api/v2/stats/subagents", h.GetSubagentStatsV2).Methods("GET")
	r.HandleFunc("/api/v2/stats/performance", h.GetPerformanceStatsV2).Methods("GET")

	// V2 Configuration API
	r.HandleFunc("/api/v2/config", h.GetConfigV2).Methods("GET")
	r.HandleFunc("/api/v2/config/providers", h.GetProvidersV2).Methods("GET")
	r.HandleFunc("/api/v2/config/subagents", h.GetSubagentConfigV2).Methods("GET")

	// CC-VIZ Claude Directory API
	r.HandleFunc("/api/v2/claude/config", h.GetClaudeConfigV2).Methods("GET")
	r.HandleFunc("/api/v2/claude/projects", h.GetClaudeProjectsV2).Methods("GET")
	r.HandleFunc("/api/v2/claude/projects/{id}", h.GetClaudeProjectDetailV2).Methods("GET")

	r.NotFoundHandler = http.HandlerFunc(h.NotFound)

	// Get port from environment or default
	port := os.Getenv("PROXY_DATA_PORT")
	if port == "" {
		port = "8002"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      corsHandler(r),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		logger.Printf("proxy-data server running on http://localhost:%s", port)
		logger.Printf("Dashboard API endpoints available at:")
		logger.Printf("   - GET  /api/requests (Request data)")
		logger.Printf("   - GET  /api/stats (Statistics)")
		logger.Printf("   - GET  /api/conversations (Conversations)")
		logger.Printf("   - GET  /api/v2/* (V2 API)")
		logger.Printf("   - GET  /health")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Println("Shutting down proxy-data...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Println("proxy-data exited")
}
