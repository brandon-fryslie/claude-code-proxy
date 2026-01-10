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
	logger := log.New(os.Stdout, "proxy-core: ", log.LstdFlags|log.Lshortfile)

	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize providers dynamically based on format
	providers := make(map[string]provider.Provider)
	for name, providerCfg := range cfg.Providers {
		switch providerCfg.Format {
		case "anthropic":
			providers[name] = provider.NewAnthropicProvider(name, providerCfg)
			logger.Printf("Initialized Anthropic-format provider: %s (%s)", name, providerCfg.BaseURL)
		case "openai":
			providers[name] = provider.NewOpenAIProvider(name, providerCfg)
			logger.Printf("Initialized OpenAI-format provider: %s (%s)", name, providerCfg.BaseURL)
		default:
			logger.Printf("Unknown provider format '%s' for provider '%s', skipping", providerCfg.Format, name)
		}
	}

	if len(providers) == 0 {
		logger.Fatalf("No providers configured. Please configure at least one provider in config.yaml")
	}

	// Initialize model router
	modelRouter := service.NewModelRouter(cfg, providers, logger)

	// Use SQLite storage (write-only for proxy-core)
	storageService, err := service.NewSQLiteStorageService(&cfg.Storage)
	if err != nil {
		logger.Fatalf("Failed to initialize SQLite storage: %v", err)
	}
	logger.Println("SQLite database ready")

	// Create core handler (minimal dependencies)
	h := handler.NewCoreHandler(storageService, logger, modelRouter, cfg)

	r := mux.NewRouter()

	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"*"}),
	)

	r.Use(middleware.Logging)

	// Core proxy routes only
	r.HandleFunc("/v1/chat/completions", h.ChatCompletions).Methods("POST")
	r.HandleFunc("/v1/messages", h.Messages).Methods("POST")
	r.HandleFunc("/v1/models", h.Models).Methods("GET")
	r.HandleFunc("/health", h.Health).Methods("GET")

	r.NotFoundHandler = http.HandlerFunc(h.NotFound)

	// Get port from environment or config
	port := os.Getenv("PROXY_CORE_PORT")
	if port == "" {
		port = "8001"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      corsHandler(r),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		logger.Printf("proxy-core server running on http://localhost:%s", port)
		logger.Printf("Endpoints:")
		logger.Printf("   - POST /v1/messages (Anthropic format)")
		logger.Printf("   - GET  /v1/models")
		logger.Printf("   - GET  /health")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Println("Shutting down proxy-core...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Println("proxy-core exited")
}
