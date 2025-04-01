package tcp

import (
	"fmt"
	"time"
)

func (c *Client) WriteWithRetry(data []byte, maxRetries int, backoff time.Duration) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if err := c.Write(data); err == nil {
			return nil
		} else {
			lastErr = err
			time.Sleep(backoff)
			reconnectErr := c.Reconnect()
			if reconnectErr != nil {
				return reconnectErr
			}
		}
	}

	return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

func (c *Client) ReadWithRetry(maxRetries int, backoff time.Duration) ([]byte, error) {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if data, err := c.Read(); err == nil {
			return data, nil
		} else {
			lastErr = err
			time.Sleep(backoff)
		}
	}

	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
