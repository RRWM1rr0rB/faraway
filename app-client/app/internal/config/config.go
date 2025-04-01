package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/logging"
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

type TCPClientConfig struct {
	URL             string        `yaml:"URL" env:"TCP_CLIENT_URL"`
	ReadTimeout     time.Duration `env:"READ_TIMEOUT" envDefault:"2s"`
	SolutionTimeout time.Duration `env:"SOLUTION_TIMEOUT" envDefault:"10s"`
}

type ProfilerConfig struct {
	IsEnabled         bool          `yaml:"enabled" env:"PROFILER_ENABLED"`
	Host              string        `yaml:"host" env:"PROFILER_HOST"`
	Port              int           `yaml:"port" env:"PROFILER_PORT"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout" env:"PROFILER_READ_HEADER_TIMEOUT"`
}

type Config struct {
	App       AppConfig       `yaml:"app"`
	TCPClient TCPClientConfig `yaml:"tcp_client"`
	Profiler  ProfilerConfig  `yaml:"profiler"`
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

	if readInConfigErr := viper.ReadInConfig(); readInConfigErr != nil {
		log.Fatalf("failed to read config file: %v", readInConfigErr)
	}

	cfg = &Config{}
	if viperUnmErr := viper.Unmarshal(&cfg); viperUnmErr != nil {
		log.Fatalf("failed to unmarshal config: %v", viperUnmErr)
	}

	fmt.Printf("Config loaded from: %s\n", absPath)

	// Watch file changes
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("Config file changed:", e.Name)
		newCfg := &Config{}
		if unmarshalErr := viper.Unmarshal(&newCfg); unmarshalErr != nil {
			log.Printf("Failed to reload config: %v", unmarshalErr)
		} else {
			cfg = newCfg
			log.Println("Config reloaded successfully")
		}
	})

	return cfg
}
