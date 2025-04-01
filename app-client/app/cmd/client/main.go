package client

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/logging"

	"app-client/app/internal/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	newApp, err := app.NewApp(ctx)
	if err != nil {
		logging.L(ctx).Error("can't init a new app", logging.ErrAttr(err))
		os.Exit(1)
	}

	go func() {
		if runErr := newApp.Run(ctx); runErr != nil {
			logging.L(ctx).Error("app run failed", logging.ErrAttr(runErr))
		}
	}()

	<-sigChan
	logging.L(ctx).Info("shutting down application")
	cancel()
}
