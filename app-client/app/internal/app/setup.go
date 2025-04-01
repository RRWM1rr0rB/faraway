package app

import (
	"context"
	"fmt"

	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/errors"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/logging"

	"app-client/app/internal/config"
	"app-client/app/internal/controller/tcp/v1/mitigator"
)

// Runner defines the interface for runnable application components.
type Runner interface {
	Run(ctx context.Context) error
}

// ClientRunner implements the Runner interface to execute the main client logic.
type ClientRunner struct {
	controller *mitigator.Controller
	cfg        *config.TCPClientConfig
}

// NewClientRunner creates a new client runner.
func NewClientRunner(controller *mitigator.Controller, cfg *config.TCPClientConfig) *ClientRunner {
	if controller == nil {
		logging.Default().Error("controller cannot be nil")
	}
	if cfg == nil {
		logging.Default().Error("tcp client config cannot be nil")
	}
	return &ClientRunner{
		controller: controller,
		cfg:        cfg,
	}
}

// Run executes the GetQuote logic.
func (r *ClientRunner) Run(ctx context.Context) error {
	logging.L(ctx).With(logging.StringAttr("component", "client_runner"))
	logging.L(ctx).Info("Client runner started")

	quote, err := r.controller.GetQuote(ctx, r.cfg)
	if err != nil {
		logging.L(ctx).Error("Failed to get quote", logging.ErrAttr(err))
		return errors.Wrap(err, "client run failed")
	}

	// Print the received quote to standard output
	fmt.Println("\n--- Word of Wisdom ---")
	fmt.Println(quote)
	fmt.Println("----------------------")

	logging.L(ctx).Info("Client runner finished successfully")
	return nil
}
