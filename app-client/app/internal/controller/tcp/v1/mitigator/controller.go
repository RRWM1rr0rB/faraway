package mitigator

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"app-client/app/internal/config"
	"app-client/app/internal/policy/mitigator"
)

func (c *Client) GetQuote() (string, error) {
	serverConn, err := net.Dial(config.TCP, c.cfg.TCPClient.URL)
	if err != nil {
		return "", fmt.Errorf("failed to connect with server on %v: %w", c.cfg.TCPClient.URL, err)
	}

	defer func() {
		log.Printf("closing connection to server")
		tcp.CloseConnection(serverConn)
	}()

	log.Printf("connected with server")

	challenge, err := c.waitChallenge(serverConn)
	if err != nil {
		return "", fmt.Errorf("failed to received challenge: %w", err)
	}

	solution, err := c.policy.SolveChallenge(*challenge)
	if err != nil {
		return "", fmt.Errorf("failed to solve challenge: %w", err)
	}

	err = sendSolution(serverConn, solution)
	if err != nil {
		return "", fmt.Errorf("failed to send solution to server: %w", err)
	}

	quote, err := c.waitQuote(serverConn)
	if err != nil {
		return "", fmt.Errorf("failed to receive quote: %w", err)
	}

	return quote, nil
}

func (c *Client) waitChallenge(serverConn net.Conn) (*mitigator.PoWCalc, error) {
	data, err := tcp.ReadWithDeadline(serverConn, c.connReadDeadline)
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	message := api.Challenge{}
	err = json.Unmarshal(data, &message)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal challenge %s: %w", data, err)
	}

	challenge := challenger.ChallengeInfo{
		RandomString:          message.RandomString,
		NumberLeadingZeros:    message.NumberLeadingZeros,
		SolutionNumberSymbols: message.SolutionNumberSymbols,
	}

	log.Printf("challenge received: %+v", challenge)

	return &challenge, nil
}

func sendSolution(serverConn net.Conn, solution string) error {
	message := api.Solution{
		Solution: solution,
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
		return "", fmt.Errorf("failed to read message: %w", err)
	}

	message := api.Quote{}
	err = json.Unmarshal(data, &message)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal quote %s: %w", data, err)
	}

	log.Printf("quote received: %+v", message)

	return message.Quote, nil
}
