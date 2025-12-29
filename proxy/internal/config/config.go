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
	Routing   RoutingConfig              `yaml:"routing"`
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
	Format           string `yaml:"format"`            // Required: "anthropic" or "openai"
	BaseURL          string `yaml:"base_url"`          // Required: API base URL
	APIKey           string `yaml:"api_key"`           // Optional: API key (required for some providers)
	Version          string `yaml:"version"`           // Optional: API version (for Anthropic-format providers)
	MaxRetries       int    `yaml:"max_retries"`       // Optional: Max retry attempts (default: 3)
	FallbackProvider string `yaml:"fallback_provider"` // Optional: Provider to use when this one fails
	CircuitBreaker   CircuitBreakerConfig `yaml:"circuit_breaker"` // Optional: Circuit breaker settings
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	Enabled     bool   `yaml:"enabled"`      // Optional: Enable circuit breaker (default: true for providers with fallback)
	MaxFailures int    `yaml:"max_failures"` // Optional: Failures before opening circuit (default: 5)
	Timeout     string `yaml:"timeout"`      // Optional: Time before retry in half-open state (default: 30s)

	// Parsed timeout duration (not in YAML)
	TimeoutDuration time.Duration `yaml:"-"`
}

type StorageConfig struct {
	RequestsDir string `yaml:"requests_dir"`
	DBPath      string `yaml:"db_path"`
}

type SubagentsConfig struct {
	Enable   bool              `yaml:"enable"`
	Mappings map[string]string `yaml:"mappings"` // agentName -> "provider:model"
}

// RoutingConfig holds preference-based routing configuration
type RoutingConfig struct {
	Preferences      PreferencesConfig                 `yaml:"preferences"`
	Tasks            map[string]TaskRoutingConfig      `yaml:"tasks"`
	ProviderProfiles map[string]ProviderProfileConfig  `yaml:"provider_profiles"`
}

// PreferencesConfig holds default routing preferences
type PreferencesConfig struct {
	Default string `yaml:"default"` // cost, speed, quality, balanced
}

// TaskRoutingConfig defines routing for a specific task type
type TaskRoutingConfig struct {
	Preference string   `yaml:"preference"` // cost, speed, quality, balanced
	Providers  []string `yaml:"providers"`  // Preferred providers for this task
}

// ProviderProfileConfig describes provider characteristics
type ProviderProfileConfig struct {
	Speed   int `yaml:"speed"`   // 1-10 scale
	Cost    int `yaml:"cost"`    // 1-10 scale
	Quality int `yaml:"quality"` // 1-10 scale
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
		Routing: RoutingConfig{
			Preferences: PreferencesConfig{
				Default: "balanced",
			},
			Tasks:            make(map[string]TaskRoutingConfig),
			ProviderProfiles: make(map[string]ProviderProfileConfig),
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

	// Parse circuit breaker timeout durations and apply defaults
	for name, provider := range cfg.Providers {
		// Apply circuit breaker defaults
		if provider.MaxRetries == 0 {
			provider.MaxRetries = 3
		}

		// Parse circuit breaker timeout
		if provider.CircuitBreaker.Timeout != "" {
			if duration, err := time.ParseDuration(provider.CircuitBreaker.Timeout); err == nil {
				provider.CircuitBreaker.TimeoutDuration = duration
			} else {
				return nil, fmt.Errorf("provider '%s': invalid circuit_breaker.timeout '%s': %w", name, provider.CircuitBreaker.Timeout, err)
			}
		} else {
			// Default timeout: 30s
			provider.CircuitBreaker.TimeoutDuration = 30 * time.Second
		}

		// Default max failures: 5
		if provider.CircuitBreaker.MaxFailures == 0 {
			provider.CircuitBreaker.MaxFailures = 5
		}

		// Enable circuit breaker by default if fallback is configured
		if provider.FallbackProvider != "" && !provider.CircuitBreaker.Enabled {
			provider.CircuitBreaker.Enabled = true
		}
	}

	// Apply routing defaults
	if cfg.Routing.Preferences.Default == "" {
		cfg.Routing.Preferences.Default = "balanced"
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

		// Validate fallback provider exists
		if provider.FallbackProvider != "" {
			if _, exists := c.Providers[provider.FallbackProvider]; !exists {
				return fmt.Errorf("provider '%s' has invalid fallback_provider '%s' (provider does not exist)", name, provider.FallbackProvider)
			}

			// Prevent circular fallback
			if provider.FallbackProvider == name {
				return fmt.Errorf("provider '%s' cannot have itself as fallback_provider", name)
			}

			// Check for fallback chains (A -> B -> A)
			if err := c.checkFallbackChain(name, provider.FallbackProvider, make(map[string]bool)); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkFallbackChain detects circular fallback chains
func (c *Config) checkFallbackChain(original string, current string, visited map[string]bool) error {
	if visited[current] {
		return fmt.Errorf("circular fallback chain detected involving provider '%s'", original)
	}

	visited[current] = true

	if provider, exists := c.Providers[current]; exists && provider.FallbackProvider != "" {
		return c.checkFallbackChain(original, provider.FallbackProvider, visited)
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
