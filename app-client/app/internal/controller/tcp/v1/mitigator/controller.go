package mitigator

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/core/tcp"

	"app-client/app/internal/config"
	"app-client/app/internal/policy/mitigator"
)

func (c *Client) GetChallenge() (string, error) {
	serverConn, err := net.Dial(config.TCP, c.cfg.TCPClient.URL)
	if err != nil {
		return "", fmt.Errorf("failed to connect with server on %v: %w", c.cfg.TCPClient.URL, err)
	}

	defer func() {
		log.Printf("closing connection to server")
		tcp.Close(serverConn)
	}()

	log.Printf("connected with server")

	challenge, waitChallengeErr := c.waitChallenge(serverConn)
	if waitChallengeErr != nil {
		return "", fmt.Errorf("failed to received challenge: %w", waitChallengeErr)
	}

	solution, solvePoWChallengeErr := c.policy.SolvePoWChallenge(challenge)
	if solvePoWChallengeErr != nil {
		return "", fmt.Errorf("failed to solve challenge: %w", solvePoWChallengeErr)
	}

	err = sendSolution(serverConn, solution.Nonce)
	if err != nil {
		return "", fmt.Errorf("failed to send solution to server: %w", err)
	}

	quote, waitQuoteErr := c.waitQuote(serverConn)
	if waitQuoteErr != nil {
		return "", fmt.Errorf("failed to receive quote: %w", waitQuoteErr)
	}

	return quote, nil
}

func (c *Client) waitChallenge(serverConn net.Conn) (*mitigator.PoWChallenge, error) {
	challenge, err := tcp.ReadPoWChallenge(serverConn)
	if err != nil {
		return nil, fmt.Errorf("failed to read PoW challenge: %w", err)
	}

	log.Printf("challenge received: %+v", challenge)

	powChallenge := &mitigator.PoWChallenge{
		Timestamp:   challenge.Timestamp,
		RandomBytes: challenge.RandomBytes,
		Difficulty:  challenge.Difficulty,
	}

	return powChallenge, nil
}

func sendSolution(serverConn net.Conn, solution uint64) error {
	message := mitigator.PoWSolution{
		Nonce: solution,
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal api.Solution: %w", err)
	}

	log.Printf("send solution %s to server", data)

	_, err = serverConn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send solution to server: %w", err)
	}

	return nil
}

func (c *Client) waitQuote(serverConn net.Conn) (string, error) {
	data, err := tcp.ReadWithDeadline(serverConn, c.connReadDeadline)
	if err != nil {
		return "", fmt.Errorf("failed to read quote: %w", err)
	}

	message := mitigator.Quote{}
	err = json.Unmarshal(data, &message)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal quote: %w", err)
	}

	log.Printf("quote received: %+v", message)

	return message.Quote, nil
}
