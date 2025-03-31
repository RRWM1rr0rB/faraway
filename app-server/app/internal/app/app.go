package app

import (
	"context"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/core/uuid/google_uuid"
	"net/http"

	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/core/clock"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/core/closer"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/core/safe/errorgroup"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/errors"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/logging"

	"app-server/app/internal/config"
	serviceSpot "server/app/internal/domain/spot/service"
	"server/app/internal/policy"
	policySpot "server/app/internal/policy/spot"
)

const (
	cfgPath = "/Users/wm1rr0rb/Downloads/NewWork/trade/configs/config.yaml"
)

type Runner interface {
	Run(context.Context) error
}

type App struct {
	cfg *config.Config

	httpRouter         *chi.Mux
	httpServer         *http.Server
	metricsHTTTPServer *metrics.Server

	policySpot       *policySpot.Policy
	adapterSpotKline *wsKline.SpotWebSocket

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

	// Init storage and service.
	//----------------------------------------- Storage and Service ----------------------------------------------------
	storageSpot := storageSpot.NewStorage(postgresClient)
	serviceSpot := serviceSpot.NewService(storageSpot)

	//----------------------------------------- End Storage and Service ------------------------------------------------

	// Init policy.
	basePolicy := policy.NewBasePolicy(
		defClock,
	)

	//----------------------------------------- Binance Policy ---------------------------------------------------------
	app.policySpot = policySpot.NewPolicy(
		basePolicy,
		serviceSpot,
		cfg,
	)

	app.adapterSpotKline = wsKline.NewSpotWebSocket(
		app.policySpot,
		cfg.Binance.Symbol,
		cfg.Binance.Interval,
		cfg.WebSocket.ReconnectTimes,
		cfg.WebSocket.ReconnectDelay,
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

	// Start HTTP server
	g.Go(func(ctx context.Context) error {
		return a.httpServer.ListenAndServe()
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
