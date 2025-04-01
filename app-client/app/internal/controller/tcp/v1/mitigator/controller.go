package mitigator

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/core/tcp"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/errors"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/logging"

	"app-client/app/internal/config"
	"app-client/app/internal/policy/mitigator"
)

// GetQuote connects to the server, solves PoW, and retrieves a quote.
func (c *Controller) GetQuote(ctx context.Context, cfg *config.TCPClientConfig) (string, error) {
	logger := logging.L(ctx).With(logging.StringAttr("component", "tcp_controller"))
	logger.Info("Attempting to get quote from server", logging.StringAttr("url", cfg.URL))

	// 1. Connect to Server
	// Assuming NewClient handles basic connection setup.
	// Add TLS config here if needed based on cfg.TLSEnabled etc.
	client, err := tcp.NewClient(cfg.URL, nil /* tls config */)
	if err != nil {
		return "", errors.Wrap(err, "failed to create tcp client")
	}
	defer func() {
		logger.Debug("Closing connection")
		if closeErr := client.Close(); closeErr != nil {
			logger.Warn("Error closing TCP client connection", logging.ErrAttr(closeErr))
		}
	}()

	if connectErr := client.Connect(); connectErr != nil {
		return "", errors.Wrap(connectErr, "failed to connect to tcp server")
	}
	logger.Info("Connected successfully", logging.StringAttr("remote_addr", client.RemoteAddr().String()))

	// 2. Wait for PoW Challenge
	logging.L(ctx).Debug("Waiting for PoW challenge...")
	challenge, challengeErr := c.waitPoWChallenge(ctx, client, cfg)
	if challengeErr != nil {
		return "", errors.Wrap(challengeErr, "failed to wait for PoW challenge")
	}

	// 3. Solve PoW Challenge
	solveCtx, cancelSolve := context.WithTimeout(ctx, cfg.SolutionTimeout)
	defer cancelSolve()
	solution, solutionErr := c.policy.SolvePoWChallenge(solveCtx, *challenge)
	if solutionErr != nil {
		return "", errors.Wrap(err, "failed to solve PoW challenge")
	}
	logger.Debug("PoW solution found", logging.Uint64Attr("nonce", solution.Nonce))

	// 4. Send PoW Solution
	solutionBytes, jsonErr := json.Marshal(solution)
	if jsonErr != nil {
		return "", errors.Wrap(jsonErr, "failed to marshal PoW solution")
	}

	logger.Debug("Sending PoW solution...")
	// Add newline for server-side reading with buffered reader.
	if writeErr := client.Write(append(solutionBytes, '\n')); writeErr != nil {
		return "", errors.Wrap(writeErr, "failed to write PoW solution")
	}

	// 5. Wait for Quote
	logger.Debug("Waiting for quote...")
	// Set deadline for reading the quote
	readCtxQuote, cancelReadQuote := context.WithTimeout(ctx, cfg.ReadTimeout)
	defer cancelReadQuote()
	quoteRespBytes, quoteRespBytesErr := readFromClient(readCtxQuote, client)
	if quoteRespBytesErr != nil {
		return "", errors.Wrap(quoteRespBytesErr, "failed to read quote response")
	}

	var quoteResp mitigator.QuoteResponse
	if jsonUnmarshalErr := json.Unmarshal(quoteRespBytes, &quoteResp); jsonUnmarshalErr != nil {
		logger.Error(
			"Failed to unmarshal quote response JSON",
			logging.ErrAttr(jsonUnmarshalErr),
			logging.StringAttr("raw_data", string(quoteRespBytes)),
		)
		return "", errors.Wrap(jsonUnmarshalErr, "failed to unmarshal quote response")
	}

	if quoteResp.Error != "" {
		logger.Warn("Received error message from server", logging.StringAttr("error", quoteResp.Error))
		return "", fmt.Errorf("server error: %s", quoteResp.Error)
	}

	if quoteResp.Quote == "" {
		logger.Warn("Received empty quote from server")
		return "", fmt.Errorf("received empty quote from server")
	}

	logger.Info("Quote received successfully")
	return quoteResp.Quote, nil
}

func (c *Controller) waitPoWChallenge(ctx context.Context, client *tcp.Client, cfg *config.TCPClientConfig) (*mitigator.PoWChallenge, error) {
	// Set deadline for reading the challenge
	readCtx, cancelReadChallenge := context.WithTimeout(ctx, cfg.ReadTimeout)
	defer cancelReadChallenge()
	challengeBytes, err := readFromClient(readCtx, client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read PoW challenge")
	}

	var challenge mitigator.PoWChallenge
	if jsonErr := json.Unmarshal(challengeBytes, &challenge); jsonErr != nil {
		// Log the raw bytes received for debugging
		logging.L(ctx).Error("Failed to unmarshal PoW challenge JSON", logging.ErrAttr(jsonErr), logging.StringAttr("raw_data", string(challengeBytes)))
		return nil, errors.Wrap(jsonErr, "failed to unmarshal pow challenge")
	}
	logging.L(ctx).Info("PoW challenge received", logging.IntAttr("difficulty", int(challenge.Difficulty)))

	return &challenge, nil
}

// Helper function to read until newline with context support using the tcp.Client's Read method
func readFromClient(ctx context.Context, client *tcp.Client) ([]byte, error) {
	type result struct {
		data []byte
		err  error
	}
	resChan := make(chan result, 1)

	go func() {
		var buffer []byte
		for {
			chunk, err := client.Read()
			if err != nil {
				resChan <- result{data: buffer, err: err}
				return
			}
			buffer = append(buffer, chunk...)
			if len(chunk) > 0 && chunk[len(chunk)-1] == '\n' {
				resChan <- result{data: buffer, err: nil}
				return
			}
			if len(chunk) == 0 {
				resChan <- result{data: buffer, err: nil} // Consider this as end of stream or no more data for now
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
		// Reading timed out or context was cancelled
		return nil, ctx.Err()
	case res := <-resChan:
		// Got a result (data or error) from reading
		return res.data, res.err
	}
}
