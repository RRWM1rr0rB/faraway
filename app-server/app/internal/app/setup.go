package app

import (
	"context"
	"fmt"
	"slog"

	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/core/tcp"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/logging"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/pprof"

	"app-server/app/internal/config"
	"app-server/app/internal/controller/tcp/v1/mitigator"
	policy "app-server/app/internal/policy/mitigator"
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
	if !a.cfg.IsDevelopment {
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

// setupDependencies initializes all dependencies for the application.
func setupDependencies(ctx context.Context, cfg *config.Config) (*App, error) {
	// 1. Setup Policies (Wisdom Provider)
	wisdomProvider := policy.New(log)
	log.Info("Wisdom provider initialized")

	// 2. Setup TCP Handler
	tcpHandler := mitigator.NewController(wisdomProvider, cfg.TCP.HandlerTimeout)
	log.Info("TCP handler initialized")

	// 3. Setup TCP Server Options
	serverOpts := []tcp.ServerOption{
		tcp.WithAddress(cfg.TCP.Addr),
		tcp.WithLogger(log),
		tcp.WithPowDifficulty(cfg.TCP.PowDifficulty),
		tcp.WithReadTimeout(cfg.TCP.ReadTimeout),
		tcp.WithWriteTimeout(cfg.TCP.WriteTimeout),
	}

	if cfg.TCP.EnableTLS {
		tlsConfig, err := tcp.SetupServerTLS(cfg.TCP.CertFile, cfg.TCP.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to setup TLS config: %w", err)
		}
		serverOpts = append(serverOpts, tcp.WithTLS(tlsConfig))
		log.Info("TLS enabled for TCP server")
	} else {
		log.Info("TLS is disabled for TCP server")
	}

	// 4. Setup TCP Server
	tcpServer := tcp.NewServer(tcpHandler.HandleConnection, serverOpts...)
	log.Info("TCP server initialized", slog.String("address", cfg.TCP.Addr))

	// 5. Create App
	app := &App{
		log:       log,
		tcpServer: tcpServer,
		cfg:       cfg,
	}

	return app, nil
}
