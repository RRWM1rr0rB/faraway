package main

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

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigs
		logging.L(ctx).Info("Received termination signal, shutting down...")
		cancel()
	}()

	logging.L(ctx).Info("Starting application")

	newApp, err := app.NewApp(ctx)
	if err != nil {
		logging.L(ctx).Error("Failed to initialize application", logging.ErrAttr(err))
		os.Exit(1)
	}

	go func() {
		if runErr := newApp.Run(ctx); runErr != nil {
			logging.L(ctx).Error("app run failed", logging.ErrAttr(runErr))
		}
	}()

	logging.L(ctx).Info("Application finished successfully")
}
