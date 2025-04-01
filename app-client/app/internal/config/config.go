package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config structure remains similar
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

type TCPConfig struct {
	ServerAddr            string        `mapstructure:"server_addr"`
	ConnectTimeout        time.Duration `mapstructure:"connect_timeout"`
	RequestTimeout        time.Duration `mapstructure:"request_timeout"`
	RetryAttempts         int           `mapstructure:"retry_attempts"`
	RetryDelay            time.Duration `mapstructure:"retry_delay"`
	EnableTLS             bool          `mapstructure:"enable_tls"`
	CaCertFile            string        `mapstructure:"ca_cert_file"`
	TlsInsecureSkipVerify bool          `mapstructure:"tls_insecure_skip_verify"`
}

// LoadConfig loads configuration using Viper.
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// --- Set Default Values ---
	v.SetDefault("env", "local")
	v.SetDefault("app_name", "faraway-client")
	v.SetDefault("shutdown_timeout", 5*time.Second)
	v.SetDefault("logger.level", "info")
	v.SetDefault("tcp.server_addr", "localhost:8081") // Default server address
	v.SetDefault("tcp.connect_timeout", 5*time.Second)
	v.SetDefault("tcp.request_timeout", 15*time.Second)
	v.SetDefault("tcp.retry_attempts", 3)
	v.SetDefault("tcp.retry_delay", 1*time.Second)
	v.SetDefault("tcp.enable_tls", false)
	v.SetDefault("tcp.ca_cert_file", "")
	v.SetDefault("tcp.tls_insecure_skip_verify", false)

	// --- Configure Viper ---
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if !errors.As(err, &configFileNotFoundError) {
				return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
			}
			fmt.Printf("Config file not found at %s, using defaults and environment variables\n", configPath)
		}
	} else {
		fmt.Println("No config file path provided, using defaults and environment variables")
	}

	// Enable reading from environment variables
	v.AutomaticEnv()
	// Optional: Set a prefix for environment variables (e.g., FARAWAY_CLIENT_TCP_SERVER_ADDR)
	// v.SetEnvPrefix("FARAWAY_CLIENT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// --- Unmarshal into Struct ---
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// --- Post-Load Validation ---
	if cfg.TCP.ServerAddr == "" {
		return nil, errors.New("TCP server address (tcp.server_addr) cannot be empty")
	}
	if cfg.TCP.EnableTLS && cfg.TCP.CaCertFile == "" && !cfg.TCP.TlsInsecureSkipVerify {
		// If TLS is on, we usually need a CA cert unless we explicitly skip verification (not recommended for production)
		fmt.Println("Warning: TLS is enabled but no CA certificate (ca_cert_file) provided, and insecure skip verify is false. System CAs will be used.")
		// Depending on requirements, you might want to make ca_cert_file mandatory here.
	}
	if cfg.TCP.EnableTLS && cfg.TCP.TlsInsecureSkipVerify {
		fmt.Println("Warning: TLS is enabled with insecure_skip_verify=true. Certificate validation is disabled!")
	}

	return &cfg, nil
}
