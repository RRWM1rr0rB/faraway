package app

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/core/clock"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/core/closer"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/core/safe/errorgroup"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/core/tcp"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/errors"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/logging"

	"app-server/app/internal/config"
	"app-server/app/internal/policy"
	policyMitigator "app-server/app/internal/policy/mitigator"
)

const (
	cfgPath = "/Users/wm1rr0rb/workspace/faraway/configs/config.server.local.yaml"
)

type Runner interface {
	Run(context.Context) error
}

type App struct {
	cfg *config.Config

	httpServer  *http.Server
	tcpServer   *tcp.TCPServer
	rateLimiter *tcp.RateLimiter
	quotes      []string
	tcpWg       sync.WaitGroup
	tcpCtx      context.Context
	tcpCancel   context.CancelFunc

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

	// Init Redis Client.
	redisClient, redisErr := app.initRedisClient(ctx, cfg.Redis)
	if redisErr != nil {
		return nil, errors.Wrap(redisErr, "can't create redis Client")
	}

	_ = redisClient

	defClock := clock.New()

	// Init services.

	// Init policy.
	basePolicy := policy.NewBasePolicy(
		defClock,
	)

	//----------------------------------------- Binance Policy ---------------------------------------------------------
	app.policySpot = policyMitigator.NewPolicy(
		basePolicy,
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

	// --- Запуск TCP сервера ---
	if a.tcpServer != nil {
		g.Go(func(ctx context.Context) error {
			logger := logging.L(ctx).With(logging.StringAttr("component", "tcp_server_runner"))
			logger.Info("Starting TCP server...")
			if err := a.tcpServer.Start(); err != nil {
				logger.Error("TCP server start failed", logging.ErrAttr(err))
				return fmt.Errorf("tcp server start error: %w", err)
			}
			logger.Info("TCP server stopped listening.")
			return nil
		})
	} else {
		logging.L(ctx).Warn("TCP server is nil, skipping run")
	}

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
