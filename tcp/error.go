package tcp

import (
	"errors"
	"fmt"
	"net"
)

const (
	TCP   = "tcp"
	Read  = "read"
	Write = "write"
)

var (
	ErrConnectionClosed = errors.New("connection closed")
	ErrTimeout          = errors.New("operation timeout")
)

type ConnectionError struct {
	Op          string
	Err         error
	IsRetryable bool
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("%s error: %v", e.Op, e.Err)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}

func wrapError(op string, err error, retryable bool) error {
	return &ConnectionError{
		Op:          op,
		Err:         err,
		IsRetryable: retryable,
	}
}

func isNetworkErrorRetryable(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	return false
}
