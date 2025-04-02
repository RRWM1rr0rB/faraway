package mitigator

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/core/tcp"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/errors"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/logging"

	"app-server/app/internal/config"
	"app-server/app/internal/policy/mitigator"
)

func (c *Controller) HandleConnection(ctx context.Context, cfg *config.TCPConfig) error {
	// --- Find handler ---
	handler := func(conn net.Conn) {
		defer conn.Close()
		localCtx := logging.ContextWithLogger(
			context.Background(),
			logging.L(ctx).With("remote_addr", conn.RemoteAddr().String()),
		)

		logging.L(localCtx).Info("Handling new connection")

		// 1. Handling PoW Challenge
		// Sent 'conn' (type net.Conn), then handler receive how argument.
		challenge, challengeErr := c.sendChallenge(localCtx, conn)
		if challengeErr != nil {
			logging.L(localCtx).Error("failed to send challenge", "error", challengeErr)
			return // finalize the processing of this connection.
		}
		logging.L(localCtx).Info("Challenge sent")

		// 2. Handling PoW Solution
		// Transmitting 'conn'
		solution, solutionErr := c.waitSolution(conn) // Add localCtx for consistency.
		if solutionErr != nil {
			// Use errors.Is to check standard network/io errors.
			if errors.Is(solutionErr, io.EOF) || errors.Is(solutionErr, net.ErrClosed) {
				logging.L(localCtx).Warn("client disconnected before sending solution", "error", solutionErr)
			} else {
				logging.L(localCtx).Error("failed to receive solution", "error", solutionErr)
			}
			return // Завершаем обработку
		}
		logging.L(localCtx).Info("Solution received", "nonce", solution)

		// 3. Validate PoW Solution
		powSolution := &mitigator.PoWSolution{
			Nonce: solution,
		}
		if !c.policy.ValidatePoWSolution(challenge, powSolution) {
			logging.L(localCtx).Info("incorrect PoW solution", "solution", solution)
			// Send an error message to the client before closing if you need to.
			// conn.Write([]byte("ERROR: Incorrect PoW\n"))
			return //  Incorrect solution, terminate processing.
		}
		logging.L(localCtx).Info("PoW solution validated successfully")

		// 4. Send Quote to Client
		// Transmitting 'conn'
		err := c.sendQuote(localCtx, conn)
		if err != nil {
			logging.L(localCtx).Error("failed to send quote", "error", err)
			return // Finalize processing.
		}
		logging.L(localCtx).Info("Quote sent successfully")

		logging.L(localCtx).Info("Client connection handled successfully.")
	}

	// Create a new server.
	server, err := tcp.NewServer(cfg.Addr, handler, nil,
		tcp.WithServerLogger(log.New(os.Stdout, "[SERVER] ", log.LstdFlags)),
		tcp.WithServerTimeout(5*time.Minute),
	)
	if err != nil {
		logging.L(ctx).Error("Error creating server", "error", err)
		return err
	}

	// Start the server to accept connections.
	logging.L(ctx).Info("Starting server...", "address", cfg.Addr)
	err = server.Start()
	if err != nil {
		logging.L(ctx).Error("Error starting server", "error", err)
		return err
	}

	logging.L(ctx).Info("Server started successfully. Waiting for connections.")

	shutdownComplete := make(chan struct{}) // Channel for signal about finished stop.
	go func() {
		<-ctx.Done() // Wait cancel signal.
		logging.L(ctx).Info("Shutdown signal received. Stopping server...")
		shutdownTimeout := 10 * time.Second
		if stopErr := server.StopWithTimeout(shutdownTimeout); stopErr != nil {
			logging.L(ctx).Error("Server shutdown error", "error", stopErr)
		} else {
			logging.L(ctx).Info("Server stopped gracefully.")
		}
		close(shutdownComplete)
	}()

	logging.L(ctx).Info("HandleConnection is now blocking until context is cancelled.") // Добавим лог

	select {
	case <-ctx.Done():
		logging.L(ctx).Info("Context cancelled in HandleConnection. Waiting for shutdown...")
		<-shutdownComplete
		logging.L(ctx).Info("Server shutdown confirmed in HandleConnection.")
		return nil
	}
}

func (c *Controller) sendChallenge(ctx context.Context, server net.Conn) (*mitigator.PoWChallenge, error) {
	challenge, err := c.policy.GeneratePoWChallenge(5)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate pow challenge")
	}
	logging.L(ctx).Info("sending challenge")

	pow := tcp.PoWChallenge{
		Timestamp:   challenge.Timestamp,
		RandomBytes: challenge.RandomBytes,
		Difficulty:  challenge.Difficulty,
	}

	err = tcp.WritePoWChallenge(server, &pow)
	if err != nil {
		return nil, fmt.Errorf("failed to send challenge to client: %w", err)
	}

	return challenge, nil
}

func (c *Controller) waitSolution(server net.Conn) (uint64, error) {
	data, err := tcp.ReadPoWSolution(server)
	if err != nil {
		return 0, err
	}

	return data.Nonce, nil
}

func (c *Controller) sendQuote(ctx context.Context, server net.Conn) error {
	quote, err := c.policy.GetWisdom(ctx)
	if err != nil {
		return fmt.Errorf("failed to get quote: %w", err)
	}

	_ = quote

	err = tcp.WritePoWSolution(server, nil)
	if err != nil {
		return fmt.Errorf("failed to send quote to client: %w", err)
	}

	return nil
}
