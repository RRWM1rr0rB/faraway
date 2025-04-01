package tcp

import (
	"crypto/tls"
	"log"
	"time"
)

type ClientOption func(*Client)
type ServerOption func(*Server)

func WithTimeouts(read, write time.Duration) ClientOption {
	return func(c *Client) {
		c.readTimeout = read
		c.writeTimeout = write
	}
}

func WithBufferSize(size int) ClientOption {
	return func(c *Client) {
		c.bufferSize = size
	}
}

func WithClientLogger(logger *log.Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

func WithTLSClientConfig(config *tls.Config) ClientOption {
	return func(c *Client) {
		c.tlsConfig = config
	}
}

func WithServerTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.idleTimeout = timeout
	}
}

func WithServerLogger(logger *log.Logger) ServerOption {
	return func(s *Server) {
		s.logger = logger
	}
}

func WithServerTLS(config *tls.Config) ServerOption {
	return func(s *Server) {
		s.tlsConfig = config
	}
}
