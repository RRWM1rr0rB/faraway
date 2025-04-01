package mitigator

import (
	"app-client/app/internal/config"
	"app-client/app/internal/policy/mitigator"
)

type policy interface {
	SolveChallenge(mitigator.PoWCalc) (string, error)
}

type Client struct {
	policy policy
	cfg    *config.Config
}

func New(policy policy, cfg *config.Config) *Client {
	return &Client{
		policy: policy,
		cfg:    cfg,
	}
}
