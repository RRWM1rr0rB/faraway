package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/logging"

	"app-server/app/internal/app"
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

	runErrChan := make(chan error, 1)
	go func() {
		logging.L(ctx).Info("Application Run loop starting...")

		runErrChan <- newApp.Run(ctx)
		logging.L(ctx).Info("Application Run loop finished.")
	}()

	logging.L(ctx).Info("Application started successfully. Waiting for signal...")

	select {
	case sig := <-sigs:
		logging.L(ctx).Info("Received signal, initiating shutdown...", logging.StringAttr("signal", sig.String()))

		cancel()
	case runErrChanErr := <-runErrChan:
		if runErrChanErr != nil {
			logging.L(ctx).Error("Application run failed", logging.ErrAttr(runErrChanErr))
		} else {
			logging.L(ctx).Info("Application run completed.")
		}

		cancel()
	}

	logging.L(ctx).Info("Application shutdown complete.")
}
