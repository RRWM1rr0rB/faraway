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
	// AppConfig remains the same structure
	AppConfig struct {
		Env             string         `mapstructure:"env"`
		AppName         string         `mapstructure:"app_name"`
		ShutdownTimeout time.Duration  `mapstructure:"shutdown_timeout"`
		LogLevel        string         `mapstructure:"log_level"`
		Profiler        ProfilerConfig `mapstructure:"profiler"`
		TCP             TCPConfig      `mapstructure:"tcp"`
	}

	ProfilerConfig struct {
		IsEnabled         bool          `mapstructure:"enabled"`
		Host              string        `mapstructure:"host"`
		Port              int           `mapstructure:"port"`
		ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout"`
	}

	TCPConfig struct {
		Addr           string        `mapstructure:"addr"`
		PowDifficulty  int32         `mapstructure:"pow_difficulty"`
		EnableTLS      bool          `mapstructure:"enable_tls"`
		CertFile       string        `mapstructure:"cert_file"`
		KeyFile        string        `mapstructure:"key_file"`
		ReadTimeout    time.Duration `mapstructure:"read_timeout"`
		WriteTimeout   time.Duration `mapstructure:"write_timeout"`
		HandlerTimeout time.Duration `mapstructure:"handler_timeout"`
	}
)

// Load loads configuration using Viper.
func Load() (*AppConfig, error) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	v := viper.New()
	var cfg AppConfig

	// --- Set Default Values ---
	v.SetDefault("env", "local")
	v.SetDefault("app_name", "faraway-server")
	v.SetDefault("shutdown_timeout", 5*time.Second)
	v.SetDefault("log_level", "info")
	v.SetDefault("profiler.enabled", true)
	v.SetDefault("profiler.host", "localhost")
	v.SetDefault("profiler.port", 6060)
	v.SetDefault("profiler.read_header_timeout", 5*time.Second)
	v.SetDefault("tcp.addr", ":8080") // Default listen address
	v.SetDefault("tcp.pow_difficulty", 15)
	v.SetDefault("tcp.enable_tls", false)
	v.SetDefault("tcp.cert_file", "")
	v.SetDefault("tcp.key_file", "")
	v.SetDefault("tcp.read_timeout", 10*time.Second)
	v.SetDefault("tcp.write_timeout", 10*time.Second)
	v.SetDefault("tcp.handler_timeout", 20*time.Second)

	// --- Configuration file setup ---
	var configFlagPath = flag.String("config", "", "Path to the configuration file")
	flag.Parse()

	configPath := *configFlagPath
	if configPath == "" {
		configPath = os.Getenv(envConfigPath)
		if configPath == "" {
			configPath = defaultConfigPath
			fmt.Printf("Environment variable %s not set, using default config path: %s\n", envConfigPath, configPath)
		} else {
			fmt.Printf("Using config path from environment variable %s: %s\n", envConfigPath, configPath)
		}
	} else {
		fmt.Printf("Using config path from flag: %s\n", configPath)
	}

	v.SetConfigFile(configPath)
	v.SetConfigType("yaml") // Or "json", "toml", etc.

	// Attempt to read the config file
	if err := v.ReadInConfig(); err != nil {
		// It's okay if the config file doesn't exist if using defaults or env vars
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, errors.Wrap(err, "failed to read config file")
		}
		fmt.Printf("Config file %s not found, using defaults and environment variables.\n", configPath)
	} else {
		fmt.Printf("Loaded configuration from %s\n", configPath)
	}

	// --- Environment variables setup ---
	v.AutomaticEnv() // Read environment variables that match keys

	// Unmarshal the config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal config")
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
