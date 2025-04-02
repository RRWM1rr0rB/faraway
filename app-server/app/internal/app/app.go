package app

import (
	"context"
	"fmt"

	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/core/closer"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/core/safe/errorgroup"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/errors"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/logging"

	"app-server/app/internal/config"
	"app-server/app/internal/controller/tcp/v1/mitigator"
	policy "app-server/app/internal/policy/mitigator"
)

// App represents the main server application.
type App struct {
	cfg          *config.AppConfig
	recover      errorgroup.RecoverFunc
	serverRunner Runner
	closer       *closer.LIFOCloser
}

// NewApp creates and initializes a new application instance.
func NewApp(ctx context.Context) (*App, error) {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load config")
	}

	// 2. Initialize Logger (using level from config)
	logger := logging.NewLogger(
		logging.WithLevel(cfg.LogLevel),
	)
	ctx = logging.ContextWithLogger(ctx, logger)

	logging.L(ctx).With(logging.StringAttr("app", cfg.AppName))
	logging.L(ctx).Info("Logger initialized", logging.StringAttr("level", cfg.LogLevel))

	// 3. Initialize Policy (PoW powQuote)
	powQuote := policy.New()
	logging.L(ctx).Info("PoW Quote initialized")

	// 4. Initialize Controller
	tcpController := mitigator.NewController(
		powQuote,
		cfg.TCP.HandlerTimeout)
	logging.L(ctx).Info("TCP Controller initialized")

	// 5. Initialize Runner
	serverRunner := NewServerRunner(tcpController, &cfg.TCP)
	logging.L(ctx).Info("Server Runner initialized")

	return &App{
		cfg:          cfg,
		serverRunner: serverRunner,
		closer:       closer.NewLIFOCloser(),
	}, nil
}

// Run starts the application's main logic.
// For this client, it just runs the client logic once.
func (a *App) Run(ctx context.Context) error {
	g, ctx := errorgroup.WithContext(ctx, errorgroup.WithRecover(a.recover))

	logging.L(ctx).Info("Starting main application logic")

	g.Go(func(ctx context.Context) error {
		return a.setupDebug(ctx, &a.cfg.Profiler)
	})

	// Run the client logic
	g.Go(func(ctx context.Context) error {
		// Pass the group's context to the runner
		return a.serverRunner.Run(ctx)
	})

	// Wait for the client runner to complete or context cancellation
	err := g.Wait()
	if err != nil {
		// Log the specific error from the runner
		logging.L(ctx).Error("Client runner failed", logging.ErrAttr(err))
		return fmt.Errorf("application run failed: %w", err)
	}

	logging.L(ctx).Info("Main application logic finished")
	return nil
}
