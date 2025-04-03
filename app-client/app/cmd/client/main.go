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

	logging.L(ctx).Info("Starting application")

	newApp, err := app.NewApp(ctx)
	if err != nil {
		logging.L(ctx).Error("Failed to initialize application", logging.ErrAttr(err))
		os.Exit(1)
	}

	runErr := newApp.Run(ctx)
	if runErr != nil {
		logging.L(ctx).Error("app run failed", logging.ErrAttr(runErr))
		os.Exit(1)
	} else {
		logging.L(ctx).Info("Application finished task successfully")
	}

	go func() {
		<-sigs
		logging.L(ctx).Info("Received termination signal, shutting down...")
		cancel()
	}()

	<-ctx.Done()
	logging.L(ctx).Info("Application shutting down.")
}
