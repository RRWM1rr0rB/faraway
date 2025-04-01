package mitigator

import "github.com/RRWM1rr0rB/faraway_lib/backend/golang/errors"

var (
	ErrPoWTimeout      = errors.New("proof-of-work challenge solving timed out")
	ErrPoWInvalidInput = errors.New("invalid input for PoW challenge")
)
