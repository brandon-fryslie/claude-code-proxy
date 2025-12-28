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
	"github.com/seifghazi/claude-code-monitor/internal/provider"
	"github.com/seifghazi/claude-code-monitor/internal/service"
)

func main() {
	logger := log.New(os.Stdout, "proxy: ", log.LstdFlags|log.Lshortfile)

	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("❌ Failed to load configuration: %v", err)
	}

	// Initialize providers dynamically based on format
	providers := make(map[string]provider.Provider)
	for name, providerCfg := range cfg.Providers {
		switch providerCfg.Format {
		case "anthropic":
			providers[name] = provider.NewAnthropicProvider(name, providerCfg)
			logger.Printf("📡 Initialized Anthropic-format provider: %s (%s)", name, providerCfg.BaseURL)
		case "openai":
			// For Plano provider, use PlanoProvider for better logging and future extensibility
			// Otherwise use OpenAIProvider for standard OpenAI API
			if name == "plano" {
				providers[name] = provider.NewPlanoProvider(name, providerCfg)
				logger.Printf("📡 Initialized Plano (multi-LLM) provider: %s (%s)", name, providerCfg.BaseURL)
			} else {
				providers[name] = provider.NewOpenAIProvider(name, providerCfg)
				logger.Printf("📡 Initialized OpenAI-format provider: %s (%s)", name, providerCfg.BaseURL)
			}
		default:
			logger.Printf("⚠️  Unknown provider format '%s' for provider '%s', skipping", providerCfg.Format, name)
		}
	}

	if len(providers) == 0 {
		logger.Fatalf("❌ No providers configured. Please configure at least one provider in config.yaml")
	}

	// Initialize model router
	modelRouter := service.NewModelRouter(cfg, providers, logger)

	// Use SQLite storage
	storageService, err := service.NewSQLiteStorageService(&cfg.Storage)
	if err != nil {
		logger.Fatalf("❌ Failed to initialize SQLite storage: %v", err)
	}
	logger.Println("🗿 SQLite database ready")

	// Start conversation indexer
	sqliteStorage, ok := storageService.(*service.SQLiteStorageService)
	if ok {
		indexer, err := service.NewConversationIndexer(sqliteStorage)
		if err != nil {
			logger.Printf("⚠️  Failed to create conversation indexer: %v", err)
		} else {
			if err := indexer.Start(); err != nil {
				logger.Printf("⚠️  Failed to start conversation indexer: %v", err)
			}
			defer indexer.Stop()
		}
	}

	h := handler.New(storageService, logger, modelRouter, cfg)

	r := mux.NewRouter()

	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"*"}),
	)

	r.Use(middleware.Logging)

	r.HandleFunc("/v1/chat/completions", h.ChatCompletions).Methods("POST")
	r.HandleFunc("/v1/messages", h.Messages).Methods("POST")
	r.HandleFunc("/v1/models", h.Models).Methods("GET")
	r.HandleFunc("/health", h.Health).Methods("GET")

	r.HandleFunc("/", h.UI).Methods("GET")
	r.HandleFunc("/ui", h.UI).Methods("GET")
	r.HandleFunc("/api/requests", h.GetRequests).Methods("GET")
	r.HandleFunc("/api/requests/summary", h.GetRequestsSummary).Methods("GET")
	r.HandleFunc("/api/requests/latest-date", h.GetLatestRequestDate).Methods("GET")
	r.HandleFunc("/api/requests/{id}", h.GetRequestByID).Methods("GET")
	r.HandleFunc("/api/stats", h.GetStats).Methods("GET")
	r.HandleFunc("/api/stats/hourly", h.GetHourlyStats).Methods("GET")
	r.HandleFunc("/api/stats/models", h.GetModelStats).Methods("GET")
	r.HandleFunc("/api/stats/providers", h.GetProviderStats).Methods("GET")
	r.HandleFunc("/api/stats/subagents", h.GetSubagentStats).Methods("GET")
	r.HandleFunc("/api/stats/tools", h.GetToolStats).Methods("GET")
	r.HandleFunc("/api/stats/performance", h.GetPerformanceStats).Methods("GET")
	r.HandleFunc("/api/requests", h.DeleteRequests).Methods("DELETE")
	r.HandleFunc("/api/conversations", h.GetConversations).Methods("GET")
	r.HandleFunc("/api/conversations/search", h.SearchConversations).Methods("GET")
	r.HandleFunc("/api/conversations/project", h.GetConversationsByProject).Methods("GET")
	r.HandleFunc("/api/conversations/{id}", h.GetConversationByID).Methods("GET")

	// V2 API - cleaner response format for new dashboard
	r.HandleFunc("/api/v2/requests/summary", h.GetRequestsSummaryV2).Methods("GET")
	r.HandleFunc("/api/v2/requests/{id}", h.GetRequestByIDV2).Methods("GET")
	r.HandleFunc("/api/v2/conversations", h.GetConversationsV2).Methods("GET")
	r.HandleFunc("/api/v2/conversations/search", h.SearchConversations).Methods("GET")
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

	r.NotFoundHandler = http.HandlerFunc(h.NotFound)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      corsHandler(r),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		logger.Printf("🚀 Claude Code Monitor Server running on http://localhost:%s", cfg.Server.Port)
		logger.Printf("📡 API endpoints available at:")
		logger.Printf("   - POST http://localhost:%s/v1/messages (Anthropic format)", cfg.Server.Port)
		logger.Printf("   - GET  http://localhost:%s/v1/models", cfg.Server.Port)
		logger.Printf("   - GET  http://localhost:%s/health", cfg.Server.Port)
		logger.Printf("🎨 Web UI available at:")
		logger.Printf("   - GET  http://localhost:%s/ (Request Visualizer)", cfg.Server.Port)
		logger.Printf("   - GET  http://localhost:%s/api/requests (Request API)", cfg.Server.Port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("❌ Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	logger.Println("✅ Server exited")
}
