package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/core/tcp"

	"app-server/app/internal/config"
)

// App represents the main server application.
type App struct {
	log       *slog.Logger
	tcpServer *tcp.TCPServer
	cfg       *config.Config
}

// New creates a new server Application instance.
func New(ctx context.Context, log *slog.Logger, cfg *config.Config) (*App, error) {
	log.Debug("Setting up application dependencies...")
	app, err := setupDependencies(ctx, log, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to setup dependencies: %w", err)
	}
	log.Debug("Application dependencies setup complete.")
	return app, nil
}

// Run starts the server application.
// It blocks until the server is shut down or an error occurs.
func (a *App) Run() error {
	a.log.Info("Starting TCP server...")
	// Run the server. This will block until Shutdown is called or an error occurs.
	err := a.tcpServer.Run()
	if err != nil && !errors.Is(err, tcp.ErrServerClosed) {
		a.log.Error("TCP server failed", slog.String("error", err.Error()))
		return fmt.Errorf("tcp server run error: %w", err)
	}
	a.log.Info("TCP server finished running.")
	return nil // Normal exit or expected closure
}

// Stop gracefully shuts down the application.
func (a *App) Stop(ctx context.Context) error {
	a.log.Info("Attempting graceful shutdown...")

	shutdownCtx, cancel := context.WithTimeout(ctx, a.cfg.ShutdownTimeout)
	defer cancel()

	// Shutdown TCP server
	if a.tcpServer != nil {
		a.log.Debug("Shutting down TCP server...")
		if err := a.tcpServer.Shutdown(shutdownCtx); err != nil {
			a.log.Error("TCP server shutdown failed", slog.String("error", err.Error()))
			// Decide if this should return an error or just log it
			// For now, log and continue if other shutdowns are needed
			// return fmt.Errorf("tcp server shutdown failed: %w", err)
		} else {
			a.log.Info("TCP server shut down gracefully")
		}
	}

	// Add cleanup for other resources if needed (e.g., database connections)

	// Check if the shutdown context expired
	if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
		a.log.Warn("Shutdown timed out")
		return fmt.Errorf("shutdown timed out (%s)", a.cfg.ShutdownTimeout)
	}

	a.log.Info("Graceful shutdown completed.")
	return nil
}
