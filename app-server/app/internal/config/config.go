package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config remains the same structure
type Config struct {
	Env             string        `mapstructure:"env"`
	AppName         string        `mapstructure:"app_name"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	Logger          LoggerConfig  `mapstructure:"logger"`
	TCP             TCPConfig     `mapstructure:"tcp"`
}

type LoggerConfig struct {
	Level string `mapstructure:"level"`
}

type ProfilerConfig struct {
	IsEnabled         bool          `mapstructure:"enabled"`
	Host              string        `mapstructure:"host"`
	Port              int           `mapstructure:"port"`
	ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout"`
}

type TCPConfig struct {
	Addr           string        `mapstructure:"addr"`
	PowDifficulty  int           `mapstructure:"pow_difficulty"`
	EnableTLS      bool          `mapstructure:"enable_tls"`
	CertFile       string        `mapstructure:"cert_file"`
	KeyFile        string        `mapstructure:"key_file"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	HandlerTimeout time.Duration `mapstructure:"handler_timeout"`
}

// LoadConfig loads configuration using Viper.
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// --- Set Default Values ---
	v.SetDefault("env", "local")
	v.SetDefault("app_name", "faraway-server")
	v.SetDefault("shutdown_timeout", 5*time.Second)
	v.SetDefault("logger.level", "info")
	v.SetDefault("tcp.addr", ":8081") // Default listen address
	v.SetDefault("tcp.pow_difficulty", 20)
	v.SetDefault("tcp.enable_tls", false)
	v.SetDefault("tcp.cert_file", "")
	v.SetDefault("tcp.key_file", "")
	v.SetDefault("tcp.read_timeout", 10*time.Second)
	v.SetDefault("tcp.write_timeout", 10*time.Second)
	v.SetDefault("tcp.handler_timeout", 20*time.Second)

	// --- Configure Viper ---
	// Read config file if provided
	if configPath != "" {
		v.SetConfigFile(configPath)
		// Attempt to read the config file
		if err := v.ReadInConfig(); err != nil {
			// Handle specific error type: ConfigFileNotFoundError
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if !errors.As(err, &configFileNotFoundError) {
				// It's an error other than file not found
				return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
			}
			// Config file not found; proceed with defaults/env vars
			fmt.Printf("Config file not found at %s, using defaults and environment variables\n", configPath)
		}
	} else {
		// Optional: Search for config file in standard paths if no specific path is given
		// v.SetConfigName("config.server") // Name of config file (without extension)
		// v.AddConfigPath("./configs") // Path to look for the config file in
		// v.AddConfigPath(".")         // Current directory
		// if err := v.ReadInConfig(); err != nil { ... } // Handle error
		fmt.Println("No config file path provided, using defaults and environment variables")
	}

	// Enable reading from environment variables
	v.AutomaticEnv()
	// Optional: Set a prefix for environment variables (e.g., FARAWAY_SERVER_LOGGER_LEVEL)
	// v.SetEnvPrefix("FARAWAY_SERVER")
	// Replace dots in keys with underscores for env var names (e.g., logger.level -> LOGGER_LEVEL)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// --- Unmarshal into Struct ---
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// --- Post-Load Validation ---
	if cfg.TCP.EnableTLS && (cfg.TCP.CertFile == "" || cfg.TCP.KeyFile == "") {
		return nil, errors.New("TLS is enabled but cert_file or key_file is missing in config")
	}
	if cfg.TCP.Addr == "" {
		return nil, errors.New("TCP address (tcp.addr) cannot be empty")
	}

	return &cfg, nil
}
