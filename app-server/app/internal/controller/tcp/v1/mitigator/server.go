package mitigator

import (
	"context"
	"time"

	"app-server/app/internal/policy/mitigator"
)

// policy defines the interface for getting wisdom.
type policy interface {
	GetWisdom(ctx context.Context) (mitigator.WisdomDTO, error)
	GeneratePoWChallenge(difficulty int32) (*mitigator.PoWChallenge, error)
	ValidatePoWSolution(challenge *mitigator.PoWChallenge, solution *mitigator.PoWSolution) bool
}

// Controller processes incoming TCP connections for the mitigator service.
type Controller struct {
	policy         policy
	handlerTimeout time.Duration
}

// NewController creates a new TCP handler.
func NewController(policy policy, timeout time.Duration) *Controller {
	return &Controller{
		policy:         policy,
		handlerTimeout: timeout,
	}
}
