package app

import (
	"context"

	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/logging"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/pprof"

	"app-server/app/internal/config"
)

// Runner defines the interface for runnable application components.
type Runner interface {
	Run(ctx context.Context) error
}

// ClientRunner implements the Runner interface to execute the main client logic.
type ClientRunner struct {
	controller *mitigator.Controller
	cfg        *config.TCPClientConfig
}

func (a *App) setupDebug(ctx context.Context) error {
	if !a.cfg.App.IsDevelopment {
		logging.L(ctx).Info("debug service not started, because app is not in development mode")
		return nil
	}

	debugServer := pprof.NewServer(pprof.NewConfig(
		a.cfg.Profiler.Host,
		a.cfg.Profiler.Port,
		a.cfg.Profiler.ReadHeaderTimeout,
	))

	go func() {
		logging.L(ctx).Info(
			"pprof debug server started",
			logging.StringAttr("host", a.cfg.Profiler.Host),
			logging.IntAttr("port", a.cfg.Profiler.Port),
		)

		err := debugServer.Run(ctx)
		if err != nil {
			logging.L(ctx).Error("error listen debug server", logging.ErrAttr(err))
		}
	}()

	return nil
}
