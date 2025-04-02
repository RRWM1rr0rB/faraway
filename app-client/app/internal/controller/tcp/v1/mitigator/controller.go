package mitigator

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"time"

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
	challenge, challengeErr := c.waitPoWChallenge(ctx, client)
	if challengeErr != nil {
		return "", errors.Wrap(challengeErr, "failed to wait for PoW challenge")
	}

	// 3. Solve PoW Challenge
	solution, solutionErr := c.solvePoWChallenge(ctx, *challenge, cfg)
	if solutionErr != nil {
		return "", errors.Wrap(solutionErr, "failed to solve PoW challenge")
	}

	// 4. Send PoW Solution
	err = c.sendPoWSolution(ctx, client, solution)
	if err != nil {
		return "", errors.Wrap(err, "failed to send PoW solution")
	}

	// 5. Wait for Quote
	logging.L(ctx).Debug("Waiting for quote...")
	quote, quoteErr := c.waitQuoteResponse(ctx, client, cfg)
	if quoteErr != nil {
		return "", errors.Wrap(quoteErr, "failed to wait for quote")
	}

	return quote, nil
}

// waitPoWChallenge waits for a PoW challenge from the server and returns the parsed
// challenge. It returns an error if the challenge cannot be read within the timeout.
func (c *Controller) waitPoWChallenge(ctx context.Context, client *tcp.Client) (*mitigator.PoWChallenge, error) {
	const challengeSize = 44 // 8 (timestamp) + 32 (random) + 4 (difficulty)

	// Буфер для накопления прочитанных данных
	challengeBuf := bytes.NewBuffer(make([]byte, 0, challengeSize))
	bytesRead := 0

	// Установим общий таймаут на чтение задачи, если нужно
	// Например, 5 секунд. Используйте c.readTimeout из клиента или другое значение.
	readDeadline := time.Now().Add(5 * time.Second) // Пример

	for bytesRead < challengeSize {
		// Проверяем дедлайн и контекст перед каждым чтением
		if time.Now().After(readDeadline) {
			return nil, errors.New("timeout waiting for PoW challenge")
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled while waiting for PoW challenge: %w", ctx.Err())
		default:
		}

		// Читаем следующую порцию данных с помощью метода клиента
		// Метод Read() клиента уже должен обрабатывать свои внутренние таймауты и ошибки
		data, err := client.Read() // Read() возвращает прочитанные байты
		if err != nil {
			// Обрабатываем ошибки, возвращаемые client.Read()
			if errors.Is(err, tcp.ErrTimeout) { // Используйте константы ошибок из вашего tcp пакета
				logging.L(ctx).Warn("Read timeout occurred while waiting for challenge, retrying if within deadline...")
				// Не возвращаем ошибку сразу, цикл проверит общий дедлайн
				time.Sleep(100 * time.Millisecond) // Небольшая пауза перед повторной попыткой
				continue
			}
			if errors.Is(err, tcp.ErrConnectionClosed) || errors.Is(err, io.EOF) {
				return nil, errors.New("connection closed while waiting for PoW challenge")
			}
			// Оборачиваем и возвращаем другие неожиданные ошибки
			return nil, errors.Wrap(err, "failed to read PoW challenge data")
		}

		// Добавляем прочитанные данные в буфер
		challengeBuf.Write(data)
		bytesRead = challengeBuf.Len() // Обновляем количество прочитанных байт
		logging.L(ctx).Debug("Read %d bytes for challenge, total %d/%d", len(data), bytesRead, challengeSize)
	}

	// Если вышли из цикла, но прочитали больше, чем нужно (маловероятно с Read)
	if bytesRead > challengeSize {
		// Это странная ситуация, возможно, ошибка в логике сервера или клиента
		return nil, fmt.Errorf("read more bytes (%d) than expected (%d) for PoW challenge", bytesRead, challengeSize)
	}

	// Теперь у нас есть ровно 44 байта в challengeBuf
	finalBuffer := challengeBuf.Bytes()

	// Parsing data
	timestamp := int64(binary.BigEndian.Uint64(finalBuffer[0:8]))
	randomBytes := finalBuffer[8:40]
	difficulty := int32(binary.BigEndian.Uint32(finalBuffer[40:44]))

	challenge := &mitigator.PoWChallenge{
		Timestamp:   timestamp,
		RandomBytes: randomBytes,
		Difficulty:  difficulty,
	}
	logging.L(ctx).Info("Received PoW challenge: Difficulty=%d", challenge.Difficulty)

	return challenge, nil
}

// solvePoWChallenge solves the given Proof-of-Work challenge using the policy.
// It returns the solution or an error if the challenge cannot be solved within the timeout.
func (c *Controller) solvePoWChallenge(
	ctx context.Context,
	challenge mitigator.PoWChallenge,
	cfg *config.TCPClientConfig,
) (*mitigator.PoWSolution, error) {
	// Create a context with a timeout for solving the PoW challenge
	solveCtx, cancelSolve := context.WithTimeout(ctx, cfg.SolutionTimeout)
	defer cancelSolve()

	// Use the policy to solve the PoW challenge
	solution, solutionErr := c.policy.SolvePoWChallenge(solveCtx, challenge)
	if solutionErr != nil {
		return nil, errors.Wrap(solutionErr, "failed to solve PoW challenge")
	}

	// Log the successful solution with the found nonce
	logging.L(ctx).Debug("PoW solution found", logging.Uint64Attr("nonce", solution.Nonce))

	return solution, nil
}

func (c *Controller) sendPoWSolution(ctx context.Context, client *tcp.Client, solution *mitigator.PoWSolution) error {
	// Format 8 byte to send.
	solutionBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(solutionBytes, solution.Nonce)

	logging.L(ctx).Debug("Sending PoW solution", logging.Uint64Attr("nonce", solution.Nonce))
	if err := client.Write(solutionBytes); err != nil {
		return errors.Wrap(err, "failed to write PoW solution")
	}
	return nil
}

func (c *Controller) waitQuoteResponse(ctx context.Context, client *tcp.Client, cfg *config.TCPClientConfig) (string, error) {
	logging.L(ctx).Debug("Waiting for quote...")
	// Set deadline for reading the quote
	readCtxQuote, cancelReadQuote := context.WithTimeout(ctx, cfg.ReadTimeout)
	defer cancelReadQuote()
	quoteRespBytes, quoteRespBytesErr := readFromClient(readCtxQuote, client)
	if quoteRespBytesErr != nil {
		return "", errors.Wrap(quoteRespBytesErr, "failed to read quote response")
	}

	var quoteResp mitigator.QuoteResponse
	if jsonUnmarshalErr := json.Unmarshal(quoteRespBytes, &quoteResp); jsonUnmarshalErr != nil {
		logging.L(ctx).Error(
			"Failed to unmarshal quote response JSON",
			logging.ErrAttr(jsonUnmarshalErr),
			logging.StringAttr("raw_data", string(quoteRespBytes)),
		)
		return "", errors.Wrap(jsonUnmarshalErr, "failed to unmarshal quote response")
	}

	if quoteResp.Error != "" {
		logging.L(ctx).Warn("Received error message from server", logging.StringAttr("error", quoteResp.Error))
		return "", fmt.Errorf("server error: %s", quoteResp.Error)
	}

	if quoteResp.Quote == "" {
		logging.L(ctx).Warn("Received empty quote from server")
		return "", fmt.Errorf("received empty quote from server")
	}

	logging.L(ctx).Info("Quote received successfully")

	return quoteResp.Quote, nil
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
