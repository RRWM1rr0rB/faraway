package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/errors"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
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
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	vip := viper.New()
	var cfg AppConfig

	vip.SetDefault("app_name", "faraway-client")
	vip.SetDefault("log_level", "info")
	vip.SetDefault("tcp_client.url", "faraway-server:8080")
	vip.SetDefault("tcp_client.read_timeout", "15s")
	vip.SetDefault("tcp_client.solution_timeout", "60s")

	var configFlagPath = flag.String("config", "", "Path to the configuration file")
	flag.Parse()

	configPath := *configFlagPath
	if configPath != "" {
		vip.SetConfigFile(configPath)
		vip.SetConfigType("yaml")
		if err := vip.ReadInConfig(); err != nil {
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if !errors.As(err, &configFileNotFoundError) {
				return nil, errors.Wrap(err, "failed to read config file")
			}
			fmt.Printf("Config file %s not found, using defaults and environment variables.\n", configPath)
		} else {
			fmt.Printf("Loaded configuration from %s\n", configPath)
		}
	} else {
		viper.AutomaticEnv() // Чтение переменных окружения
		configPath := os.Getenv(envConfigPath)
		if configPath != "" {
			fmt.Printf("Using config path from environment variable %s: %s\n", envConfigPath, configPath)
			vip.SetConfigFile(configPath)
			vip.SetConfigType("yaml")
			if err := vip.ReadInConfig(); err != nil {
				fmt.Printf("Failed to read config file from env path: %v\n", err)
			} else {
				fmt.Printf("Loaded configuration from %s\n", configPath)
			}
		} else {
			fmt.Printf("Environment variable %s not set, using defaults and environment variables.\n", envConfigPath)
		}
	}

	if err := vip.Unmarshal(&cfg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal config")
	}

	// Optional: Add validation logic here
	if cfg.TCPClient.URL == "" {
		return nil, fmt.Errorf("tcp_client.url is required in config")
	}

	fmt.Printf("Configuration loaded: AppName=%s, LogLevel=%s, ServerURL=%s\n", cfg.AppName, cfg.LogLevel, cfg.TCPClient.URL)

	return &cfg, nil
}
