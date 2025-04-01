package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"app-server/app/internal/app"
	"app-server/app/internal/config"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "configs/config.server.local.yaml", "path to config file")
}

func main() {
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Basic logger setup until config is loaded
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	log.Info("Logger initialized")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Error("Failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Setup logger based on config
	log = setupLogger(cfg.Logger.Level)
	log.Info("Configuration loaded successfully", slog.String("env", cfg.Env))
	log.Debug("Full configuration", slog.Any("config", cfg))

	application, err := app.New(ctx, log, cfg)
	if err != nil {
		log.Error("Failed to setup application", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log.Info("Starting application", slog.String("name", cfg.AppName), slog.String("addr", cfg.TCP.Addr))

	go func() {
		if err := application.Run(); err != nil {
			log.Error("Application run failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()

	log.Info("Received shutdown signal. Shutting down gracefully...")

	// Perform graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := application.Stop(shutdownCtx); err != nil {
		log.Error("Application shutdown failed", slog.String("error", err.Error()))
		os.Exit(1) // Force exit if shutdown fails
	}

	log.Info("Application stopped gracefully")
}

func setupLogger(level string) *slog.Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
}
