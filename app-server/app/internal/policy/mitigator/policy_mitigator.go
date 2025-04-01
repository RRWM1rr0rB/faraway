package mitigator

import (
	"context"

	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/logging"
)

// GetWisdom returns a random piece of wisdom.
func (p *StaticWisdomProvider) GetWisdom(ctx context.Context) (WisdomDTO, error) {
	if len(p.quotes) == 0 {
		logging.L(ctx).Warn("No quotes configured")
		return WisdomDTO{}, ErrNoWisdomFound
	}

	// Select a random quote
	index := p.r.Intn(len(p.quotes))
	quote := p.quotes[index]

	logging.L(ctx).Info("Providing wisdom", logging.StringAttr("quote", quote))

	return WisdomDTO{Quote: quote}, nil
}
