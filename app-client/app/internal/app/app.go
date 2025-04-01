package app

import (
	"context"
	"fmt"

	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/core/closer"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/core/safe/errorgroup"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/errors"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/logging"

	"app-client/app/internal/config"
	"app-client/app/internal/controller/tcp/v1/mitigator"
	policy "app-client/app/internal/policy/mitigator"
)

// App is the main application structure.
type App struct {
	cfg          *config.AppConfig
	recover      errorgroup.RecoverFunc
	clientRunner Runner
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

	// 3. Initialize Policy (PoW Solver)
	powSolver := policy.NewPoWSolver(cfg.TCPClient.SolutionTimeout)
	logging.L(ctx).Info("PoW Solver initialized", logging.DurationAttr("max_solution_time", cfg.TCPClient.SolutionTimeout))

	// 4. Initialize Controller
	tcpController := mitigator.NewController(powSolver)
	logging.L(ctx).Info("TCP Controller initialized")

	// 5. Initialize Runner
	clientRunner := NewClientRunner(tcpController, &cfg.TCPClient)
	logging.L(ctx).Info("Client Runner initialized")

	return &App{
		cfg:          cfg,
		clientRunner: clientRunner,
		closer:       closer.NewLIFOCloser(),
	}, nil
}

// Run starts the application's main logic.
// For this client, it just runs the client logic once.
func (a *App) Run(ctx context.Context) error {
	g, ctx := errorgroup.WithContext(ctx, errorgroup.WithRecover(a.recover))

	logging.L(ctx).Info("Starting main application logic")

	// Run the client logic
	g.Go(func(ctx context.Context) error {
		// Pass the group's context to the runner
		return a.clientRunner.Run(ctx)
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
