package policy

import (
	"time"
)

type Clock interface {
	Now() time.Time
}
type BasePolicy struct {
	Clock
}

func NewBasePolicy(clock Clock) *BasePolicy {
	return &BasePolicy{Clock: clock}
}
