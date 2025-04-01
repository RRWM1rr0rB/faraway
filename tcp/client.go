package tcp

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

const (
	defaultBufferSize   = 1024
	defaultReadTimeout  = 3 * time.Second
	defaultWriteTimeout = 3 * time.Second
)

// ConnectionStats represents statistics about the connection
type ConnectionStats struct {
	BytesRead    uint64
	BytesWritten uint64
	LastActivity time.Time
	RetryCount   int
}

// Client represents a TCP client with connection management and statistics
type Client struct {
	address      string
	conn         net.Conn
	readTimeout  time.Duration
	writeTimeout time.Duration
	bufferSize   int
	logger       *log.Logger
	tlsConfig    *tls.Config
	stats        ConnectionStats
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewClient creates a new TCP client with the given configuration
func NewClient(
	address string,
	tlsConfig *tls.Config,
	opts ...ClientOption,
) (*Client, error) {
	if address == "" {
		return nil, errors.New("address cannot be empty")
	}

	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		address:      address,
		readTimeout:  defaultReadTimeout,
		writeTimeout: defaultWriteTimeout,
		bufferSize:   defaultBufferSize,
		tlsConfig:    tlsConfig,
		logger:       log.Default(),
		ctx:          ctx,
		cancel:       cancel,
		stats: ConnectionStats{
			LastActivity: time.Now(),
		},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

// Connect establishes a connection to the server
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return &ConnectionError{Op: "connect", Err: errors.New("already connected")}
	}

	var conn net.Conn
	var err error

	if c.tlsConfig != nil {
		conn, err = tls.Dial(TCP, c.address, c.tlsConfig)
	} else {
		conn, err = net.Dial(TCP, c.address)
	}

	if err != nil {
		return wrapError("connect", err, true)
	}

	c.conn = conn
	c.stats.LastActivity = time.Now()
	c.logger.Printf("Connected to %s", c.address)
	return nil
}

// Read reads data from the connection
func (c *Client) Read() ([]byte, error) {
	c.mu.RLock()
	if c.conn == nil {
		c.mu.RUnlock()
		return nil, &ConnectionError{Op: Read, Err: ErrConnectionClosed}
	}
	c.mu.RUnlock()

	select {
	case <-c.ctx.Done():
		return nil, &ConnectionError{Op: Read, Err: c.ctx.Err()}
	default:
		if err := c.conn.SetReadDeadline(time.Now().Add(c.readTimeout)); err != nil {
			return nil, wrapError("set read deadline", err, false)
		}
		defer func(conn net.Conn, t time.Time) {
			_ = conn.SetReadDeadline(t)
		}(c.conn, time.Time{})

		buf := make([]byte, c.bufferSize)
		n, err := c.conn.Read(buf)
		if err != nil {
			return nil, wrapError(Read, err, isNetworkErrorRetryable(err))
		}

		c.mu.Lock()
		c.stats.BytesRead += uint64(n)
		c.stats.LastActivity = time.Now()
		c.mu.Unlock()
		return buf[:n], nil
	}
}

// Write writes data to the connection
func (c *Client) Write(data []byte) error {
	c.mu.RLock()
	if c.conn == nil {
		c.mu.RUnlock()
		return &ConnectionError{Op: Write, Err: ErrConnectionClosed}
	}
	c.mu.RUnlock()

	select {
	case <-c.ctx.Done():
		return &ConnectionError{Op: Write, Err: c.ctx.Err()}
	default:
		if err := c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout)); err != nil {
			return wrapError("set write deadline", err, false)
		}
		defer func(conn net.Conn, t time.Time) {
			_ = conn.SetWriteDeadline(t)
		}(c.conn, time.Time{})

		n, err := c.conn.Write(data)
		if err != nil {
			return wrapError(Write, err, isNetworkErrorRetryable(err))
		}

		c.mu.Lock()
		c.stats.BytesWritten += uint64(n)
		c.stats.LastActivity = time.Now()
		c.mu.Unlock()
		return nil
	}
}

// Close closes the connection
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil
	}

	c.cancel()
	if err := c.conn.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		c.logger.Printf("Close error: %v", err)
		return wrapError("close", err, false)
	}

	c.logger.Printf("Connection closed")
	c.conn = nil
	return nil
}

// Reconnect closes the current connection and establishes a new one
func (c *Client) Reconnect() error {
	if err := c.Close(); err != nil {
		return err
	}
	return c.Connect()
}

// RemoteAddr returns the remote network address
func (c *Client) RemoteAddr() net.Addr {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.conn != nil {
		return c.conn.RemoteAddr()
	}
	return nil
}

// Stats returns the current connection statistics
func (c *Client) Stats() ConnectionStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats
}
