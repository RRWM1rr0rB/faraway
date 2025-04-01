package tcp

import (
	"io"
	"log"
)

type ConnectionPool struct {
	factory func() (*Client, error)
	pool    chan *Client
	maxSize int
	logger  *log.Logger
}

func NewConnectionPool(factory func() (*Client, error), maxSize int) *ConnectionPool {
	return &ConnectionPool{
		factory: factory,
		pool:    make(chan *Client, maxSize),
		maxSize: maxSize,
		logger:  log.New(io.Discard, "", 0),
	}
}

func (p *ConnectionPool) Get() (*Client, error) {
	select {
	case conn := <-p.pool:
		return conn, nil
	default:
		return p.factory()
	}
}

func (p *ConnectionPool) Put(conn *Client) {
	select {
	case p.pool <- conn:
		p.logger.Printf("Connection returned to pool")
	default:
		if err := conn.Close(); err != nil {
			p.logger.Printf("Error closing connection: %v", err)
		}
		p.logger.Printf("Connection pool full, connection closed")
	}
}
