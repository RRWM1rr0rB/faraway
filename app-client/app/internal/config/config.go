package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"

	ferrors "github.com/RRWM1rr0rB/faraway_lib/backend/golang/errors"
)

type (
	// AppConfig represents the application configuration structure.
	AppConfig struct {
		AppName  string `mapstructure:"app_name"`
		LogLevel string `mapstructure:"log_level"`

		TCPClient TCPClientConfig `mapstructure:"tcp_client"`
	}

	// TCPClientConfig holds the configuration specific to the TCP client.
	TCPClientConfig struct {
		URL             string        `mapstructure:"url"`              // Server address (e.g., "localhost:8081")
		ReadTimeout     time.Duration `mapstructure:"read_timeout"`     // Timeout for reading challenge/quote
		SolutionTimeout time.Duration `mapstructure:"solution_timeout"` // Max time allowed to solve PoW
		// Add TLS config if needed (e.g., InsecureSkipVerify)
		// TLSEnabled         bool `mapstructure:"tls_enabled"`
		// TLSInsecureSkipVerify bool `mapstructure:"tls_insecure_skip_verify"`
	}
)

// Load configuration from file and environment variables.
func Load() (*AppConfig, error) {
	vip := viper.New()
	var cfg AppConfig

	// Set default values
	vip.SetDefault("app_name", "faraway-client")
	vip.SetDefault("log_level", "info")
	vip.SetDefault("tcp_client.url", "localhost:8081")
	vip.SetDefault("tcp_client.read_timeout", "15s")
	vip.SetDefault("tcp_client.solution_timeout", "60s")

	// --- Configuration file setup ---
	configPath := os.Getenv(envConfigPath)
	if configPath == "" {
		configPath = defaultConfigPath
		fmt.Printf("Environment variable %s not set, using default config path: %s\n", envConfigPath, configPath)
	} else {
		fmt.Printf("Using config path from environment variable %s: %s\n", envConfigPath, configPath)
	}

	vip.SetConfigFile(configPath)
	vip.SetConfigType("yaml") // Or "json", "toml", etc.

	// Attempt to read the config file
	if err := vip.ReadInConfig(); err != nil {
		// It's okay if the config file doesn't exist if using defaults or env vars
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, ferrors.Wrap(err, "failed to read config file")
		}
		fmt.Printf("Config file %s not found, using defaults and environment variables.\n", configPath)
	} else {
		fmt.Printf("Loaded configuration from %s\n", configPath)
	}

	// --- Environment variables setup ---
	vip.AutomaticEnv() // Read environment variables that match keys

	// Unmarshal the config
	if err := vip.Unmarshal(&cfg); err != nil {
		return nil, ferrors.Wrap(err, "failed to unmarshal config")
	}

	// Optional: Add validation logic here
	if cfg.TCPClient.URL == "" {
		return nil, fmt.Errorf("tcp_client.url is required in config")
	}

	fmt.Printf("Configuration loaded: AppName=%s, LogLevel=%s, ServerURL=%s\n", cfg.AppName, cfg.LogLevel, cfg.TCPClient.URL)

	return &cfg, nil
}
