package app

import (
	"context"
	"fmt"

	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/core/closer"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/core/safe/errorgroup"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/errors"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/logging"

	"app-server/app/internal/config"
	policy "app-server/app/internal/policy/mitigator"
)

type App struct {
	cfg          *config.Config
	serverRunner Runner
	recover      errorgroup.RecoverFunc
	closer       *closer.LIFOCloser
}

//nolint:funlen
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

	// 3. Initialize Policy.
	powSolver := policy.NewPoWSolver(cfg.TCPClient.SolutionTimeout)
	logging.L(ctx).Info("PoW Solver initialized", logging.DurationAttr("max_solution_time", cfg.TCPClient.SolutionTimeout))

	return &App{
		cfg:          cfg,
		serverRunner: serverRunner,
		closer:       closer.NewLIFOCloser(),
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	// Create error group with panic recovery
	g, ctx := errorgroup.WithContext(ctx, errorgroup.WithRecover(a.recover))

	logging.L(ctx).Info("Starting main application logic")

	// Run the server logic
	g.Go(func(ctx context.Context) error {
		// Pass the group's context to the runner
		return a.serverRunner.Run(ctx)
	})

	// Start profiler if enabled
	if a.cfg.Profiler.IsEnabled {
		g.Go(func(ctx context.Context) error {
			return a.setupDebug(ctx)
		})
	}

	// Wait for the server runner to complete or context cancellation.
	err := g.Wait()
	if err != nil {
		// Log the specific error from the runner.
		logging.L(ctx).Error("Client runner failed", logging.ErrAttr(err))
		return fmt.Errorf("application run failed: %w", err)
	}

	logging.L(ctx).Info("Main application logic finished")
	return nil
}
