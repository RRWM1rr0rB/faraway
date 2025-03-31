package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type AppConfig struct {
	ID            string `yaml:"id" env:"APP_ID"`
	Name          string `yaml:"name" env:"APP_NAME"`
	Version       string `yaml:"version" env:"APP_VERSION"`
	IsDevelopment bool   `yaml:"is_dev" env:"APP_IS_DEVELOPMENT"`
	LogLevel      string `yaml:"log_level" env:"APP_LOG_LEVEL"`
	IsLogJSON     bool   `yaml:"is_log_json" env:"APP_IS_LOG_JSON"`
	Domain        string `yaml:"domain" env:"APP_DOMAIN"`
}

type HTTPConfig struct {
	Host              string        `yaml:"host" env:"HTTP_HOST"`
	Port              int           `yaml:"port" env:"HTTP_PORT"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout" env:"HTTP_READ_HEADER_TIMEOUT"`
}

type RedisConfig struct {
	Address     string        `yaml:"address" env:"REDIS_ADDRESS"`
	Password    string        `yaml:"password" env:"REDIS_PASSWORD"`
	DB          int           `yaml:"db" env:"REDIS_DB"`
	TLS         bool          `yaml:"is_tls" env:"REDIS_ISTLS"`
	MaxAttempts int           `yaml:"max_attempts" env:"REDIS_MAX_ATTEMTPS"`
	MaxDelay    time.Duration `yaml:"max_delay" env:"REDIS_MAX_DELAY"`
}

type ProfilerConfig struct {
	IsEnabled         bool          `yaml:"enabled" env:"PROFILER_ENABLED"`
	Host              string        `yaml:"host" env:"PROFILER_HOST"`
	Port              int           `yaml:"port" env:"PROFILER_PORT"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout" env:"PROFILER_READ_HEADER_TIMEOUT"`
}

type Config struct {
	App      AppConfig      `yaml:"app"`
	HTTP     HTTPConfig     `yaml:"http"`
	Redis    RedisConfig    `yaml:"redis"`
	Profiler ProfilerConfig `yaml:"profiler"`
}

var (
	cfg *Config
)

func LoadConfig(configPath string) *Config {
	if configPath == "" {
		if envPath, ok := os.LookupEnv("CONFIG_PATH"); ok {
			configPath = envPath
		} else {
			log.Fatal("No config path provided and CONFIG_PATH not set")
		}
	}

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		log.Fatalf("failed to get absolute path of config: %v", err)
	}

	viper.SetConfigFile(absPath)
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}

	cfg = &Config{}
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}

	fmt.Printf("Config loaded from: %s\n", absPath)

	// Watch file changes
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("Config file changed:", e.Name)
		newCfg := &Config{}
		if err := viper.Unmarshal(&newCfg); err != nil {
			log.Printf("Failed to reload config: %v", err)
		} else {
			cfg = newCfg
			log.Println("Config reloaded successfully")
		}
	})

	return cfg
}
