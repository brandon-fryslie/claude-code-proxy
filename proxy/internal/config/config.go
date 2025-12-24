package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig               `yaml:"server"`
	Providers map[string]*ProviderConfig `yaml:"providers"`
	Storage   StorageConfig              `yaml:"storage"`
	Subagents SubagentsConfig            `yaml:"subagents"`
}

type ServerConfig struct {
	Port     string         `yaml:"port"`
	Timeouts TimeoutsConfig `yaml:"timeouts"`
	// Legacy fields
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type TimeoutsConfig struct {
	Read  string `yaml:"read"`
	Write string `yaml:"write"`
	Idle  string `yaml:"idle"`
}

// ProviderConfig is the unified configuration for all providers
type ProviderConfig struct {
	Format     string `yaml:"format"`     // Required: "anthropic" or "openai"
	BaseURL    string `yaml:"base_url"`   // Required: API base URL
	APIKey     string `yaml:"api_key"`    // Optional: API key (required for some providers)
	Version    string `yaml:"version"`    // Optional: API version (for Anthropic-format providers)
	MaxRetries int    `yaml:"max_retries"` // Optional: Max retry attempts
}

type StorageConfig struct {
	RequestsDir string `yaml:"requests_dir"`
	DBPath      string `yaml:"db_path"`
}

type SubagentsConfig struct {
	Enable   bool              `yaml:"enable"`
	Mappings map[string]string `yaml:"mappings"` // agentName -> "provider:model"
}

func Load() (*Config, error) {
	// Load .env file if it exists
	// Look for .env file in the project root (one level up from proxy/)
	envPath := filepath.Join("..", ".env")
	if err := godotenv.Load(envPath); err != nil {
		// If .env doesn't exist in parent directory, try current directory
		if err := godotenv.Load(".env"); err != nil {
			// .env file is optional, so we just log and continue
			// This allows the app to work with system environment variables only
		}
	}

	// Start with default configuration
	cfg := &Config{
		Server: ServerConfig{
			Port:         "3001",
			ReadTimeout:  600 * time.Second,
			WriteTimeout: 600 * time.Second,
			IdleTimeout:  600 * time.Second,
		},
		Providers: map[string]*ProviderConfig{
			"anthropic": {
				Format:     "anthropic",
				BaseURL:    "https://api.anthropic.com",
				Version:    "2023-06-01",
				MaxRetries: 3,
			},
		},
		Storage: StorageConfig{
			DBPath: "requests.db",
		},
		Subagents: SubagentsConfig{
			Enable:   false,
			Mappings: make(map[string]string),
		},
	}

	// Try to load config.yaml from the project root
	// The proxy binary is in proxy/ directory, config.yaml is in the parent
	configPath := filepath.Join(filepath.Dir(os.Args[0]), "..", "config.yaml")

	// If that doesn't work, try relative to current directory
	if _, err := os.Stat(configPath); err != nil {
		// Try common locations relative to where the binary might be run
		for _, tryPath := range []string{"config.yaml", "../config.yaml", "../../config.yaml"} {
			if _, err := os.Stat(tryPath); err == nil {
				configPath = tryPath
				break
			}
		}
	}

	cfg.loadFromFile(configPath)

	// Apply environment variable overrides AFTER loading from file
	if envPort := os.Getenv("PORT"); envPort != "" {
		cfg.Server.Port = envPort
	}
	if envTimeout := os.Getenv("READ_TIMEOUT"); envTimeout != "" {
		cfg.Server.ReadTimeout = getDuration("READ_TIMEOUT", cfg.Server.ReadTimeout)
	}
	if envTimeout := os.Getenv("WRITE_TIMEOUT"); envTimeout != "" {
		cfg.Server.WriteTimeout = getDuration("WRITE_TIMEOUT", cfg.Server.WriteTimeout)
	}
	if envTimeout := os.Getenv("IDLE_TIMEOUT"); envTimeout != "" {
		cfg.Server.IdleTimeout = getDuration("IDLE_TIMEOUT", cfg.Server.IdleTimeout)
	}

	// Override Anthropic provider settings if env vars are set and provider exists
	if anthropicCfg, exists := cfg.Providers["anthropic"]; exists {
		if envURL := os.Getenv("ANTHROPIC_FORWARD_URL"); envURL != "" {
			anthropicCfg.BaseURL = envURL
		}
		if envVersion := os.Getenv("ANTHROPIC_VERSION"); envVersion != "" {
			anthropicCfg.Version = envVersion
		}
		if envRetries := os.Getenv("ANTHROPIC_MAX_RETRIES"); envRetries != "" {
			anthropicCfg.MaxRetries = getInt("ANTHROPIC_MAX_RETRIES", anthropicCfg.MaxRetries)
		}
	}

	// Override OpenAI provider settings if env vars are set and provider exists
	if openaiCfg, exists := cfg.Providers["openai"]; exists {
		if envURL := os.Getenv("OPENAI_BASE_URL"); envURL != "" {
			openaiCfg.BaseURL = envURL
		}
		if envKey := os.Getenv("OPENAI_API_KEY"); envKey != "" {
			openaiCfg.APIKey = envKey
		}
	}

	// Override storage settings
	if envPath := os.Getenv("DB_PATH"); envPath != "" {
		cfg.Storage.DBPath = envPath
	}

	// After loading from file, apply any timeout conversions if needed
	if cfg.Server.Timeouts.Read != "" {
		if duration, err := time.ParseDuration(cfg.Server.Timeouts.Read); err == nil {
			cfg.Server.ReadTimeout = duration
		}
	}
	if cfg.Server.Timeouts.Write != "" {
		if duration, err := time.ParseDuration(cfg.Server.Timeouts.Write); err == nil {
			cfg.Server.WriteTimeout = duration
		}
	}
	if cfg.Server.Timeouts.Idle != "" {
		if duration, err := time.ParseDuration(cfg.Server.Timeouts.Idle); err == nil {
			cfg.Server.IdleTimeout = duration
		}
	}

	// Validate provider configurations
	if err := cfg.validateProviders(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validateProviders() error {
	for name, provider := range c.Providers {
		if provider.Format == "" {
			return fmt.Errorf("provider '%s' is missing required 'format' field (must be 'anthropic' or 'openai')", name)
		}
		if provider.Format != "anthropic" && provider.Format != "openai" {
			return fmt.Errorf("provider '%s' has invalid format '%s' (must be 'anthropic' or 'openai')", name, provider.Format)
		}
		if provider.BaseURL == "" {
			return fmt.Errorf("provider '%s' is missing required 'base_url' field", name)
		}
	}
	return nil
}

func (c *Config) loadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, c)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}

	return duration
}

func getInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}
