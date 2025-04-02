package app

import (
	"context"

	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/errors"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/logging"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/pprof"

	"app-server/app/internal/config"
	"app-server/app/internal/controller/tcp/v1/mitigator"
)

// Runner defines the interface for runnable application components.
type Runner interface {
	Run(ctx context.Context) error
}

// ClientRunner implements the Runner interface to execute the main client logic.
type ClientRunner struct {
	controller *mitigator.Controller
	cfg        *config.TCPConfig
}

// NewServerRunner creates a new client runner.
func NewServerRunner(controller *mitigator.Controller, cfg *config.TCPConfig) *ClientRunner {
	if controller == nil {
		logging.Default().Error("controller cannot be nil")
	}
	if cfg == nil {
		logging.Default().Error("tcp client config cannot be nil")
	}
	return &ClientRunner{
		controller: controller,
		cfg:        cfg,
	}
}

// Run executes the GetQuote logic.
func (r *ClientRunner) Run(ctx context.Context) error {
	logging.L(ctx).With(logging.StringAttr("component", "server_runner"))
	logging.L(ctx).Info("Server runner started")

	err := r.controller.HandleConnection(ctx, r.cfg)
	if err != nil {
		logging.L(ctx).Error("Failed to get quote", logging.ErrAttr(err))
		return errors.Wrap(err, "client run failed")
	}

	logging.L(ctx).Info("Server runner finished successfully")
	return nil
}

func (a *App) setupDebug(ctx context.Context, cfg *config.ProfilerConfig) error {
	if !cfg.IsEnabled {
		logging.L(ctx).Info("debug service not started, because app is not in development mode")
		return nil
	}

	debugServer := pprof.NewServer(pprof.NewConfig(
		cfg.Host,
		cfg.Port,
		cfg.ReadHeaderTimeout,
	))

	go func() {
		logging.L(ctx).Info(
			"pprof debug server started",
			logging.StringAttr("host", cfg.Host),
			logging.IntAttr("port", cfg.Port),
		)

		err := debugServer.Run(ctx)
		if err != nil {
			logging.L(ctx).Error("error listen debug server", logging.ErrAttr(err))
		}
	}()

	return nil
}
