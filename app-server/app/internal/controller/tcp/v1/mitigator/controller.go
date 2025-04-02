package mitigator

import (
	"app-server/app/internal/config"
	"context"
)

const (
	RequestWisdomCommand = "request_wisdom"
	Delimiter            = '\n'
)

// HandleConnection implements the tcp.HandlerFunc interface.
func (h *Controller) HandleConnection(ctx context.Context, cfg *config.TCPConfig) error {

	return nil
}
