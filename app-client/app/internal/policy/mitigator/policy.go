package mitigator

import (
	"context"
	"time"
)

// Policy defines the interface for solving Proof-of-Work challenges.
type Policy interface {
	// SolvePoWChallenge attempts to find a valid nonce for the given challenge.
	SolvePoWChallenge(ctx context.Context, challenge PoWChallenge) (*PoWSolution, error)
}

// PoWSolver implements the Policy interface for solving challenges.
type PoWSolver struct {
	maxSolutionTime time.Duration // Max time allowed to find a solution
}

// NewPoWSolver creates a new Proof-of-Work solver.
func NewPoWSolver(maxSolutionTime time.Duration) *PoWSolver {
	return &PoWSolver{
		maxSolutionTime: maxSolutionTime,
	}
}
