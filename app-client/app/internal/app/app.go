package app

import (
	"context"
	"net/http"

	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/core/clock"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/core/closer"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/core/safe/errorgroup"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/errors"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/logging"

	"app-client/app/internal/config"
	policyMitigator "app-client/app/internal/policy/mitigator"
)

const (
	cfgPath = "/Users/wm1rr0rb/workspace/faraway/configs/config.server.local.yaml"
)

type Runner interface {
	Run(context.Context) error
}

type App struct {
	cfg *config.Config

	policySpot *policyMitigator.Policy

	runners []Runner
	recover errorgroup.RecoverFunc
	closer  *closer.LIFOCloser
}

func (a *App) AddRunner(runner Runner) {
	a.runners = append(a.runners, runner)
}

//nolint:funlen
func NewApp(ctx context.Context) (*App, error) {
	app := App{
		closer: closer.NewLIFOCloser(),
	}

	cfg := config.LoadConfig(cfgPath)
	app.cfg = cfg

	logger := logging.NewLogger(
		logging.WithLevel(cfg.App.LogLevel),
		logging.WithIsJSON(cfg.App.IsLogJSON),
	)
	ctx = logging.ContextWithLogger(ctx, logger)

	logging.L(ctx).Info("config loaded", "config", cfg)

	// Init policy.

	app.policySpot = policyMitigator.NewPolicy(
		nil,
		cfg.TCPClient,
	)

	return &app, nil
}

func (a *App) Run(ctx context.Context) error {
	// Create error group with panic recovery
	g, ctx := errorgroup.WithContext(ctx, errorgroup.WithRecover(a.recover))

	// Setup graceful shutdown on signals
	g.Go(func(ctx context.Context) error {
		<-ctx.Done()
		return a.httpServer.Shutdown(context.Background())
	})

	// Start profiler if enabled
	if a.cfg.Profiler.IsEnabled {
		g.Go(func(ctx context.Context) error {
			return a.setupDebug(ctx)
		})
	}

	// Start additional runners
	for _, r := range a.runners {
		runner := r // capture loop variable
		g.Go(func(ctx context.Context) error {
			return runner.Run(ctx)
		})
	}

	logging.L(ctx).Info("application started")

	// Wait for all goroutines and return aggregated errors
	return g.Wait()
}
