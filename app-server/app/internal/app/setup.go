package app

import (
	"context"

	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/errors"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/logging"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/pprof"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/redis"

	"app-server/app/internal/config"
)

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

func (a *App) initRedisClient(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	logging.WithAttrs(
		ctx,
		logging.StringAttr("address", cfg.Address),
		logging.IntAttr("db", cfg.DB),
		logging.BoolAttr("tls", cfg.TLS),
		logging.IntAttr("max-attempts", cfg.MaxAttempts),
		logging.DurationAttr("max_delay", cfg.MaxDelay),
	).Info("Redis initializing")

	redisConfig := redis.NewRedisConfig(
		cfg.Address,
		cfg.Password,
		cfg.DB,
		cfg.TLS,
	)

	redisClient, err := redis.NewClient(ctx, redisConfig)
	if err != nil {
		return nil, errors.Wrap(err, "redis.NewClient")
	}

	a.closer.Add(redisClient)

	return redisClient, nil
}
