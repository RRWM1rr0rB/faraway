package mitigator

import (
	"app-client/app/internal/config"
)

type Policy struct {
	cfg *config.Config
}

func NewPolicy(cfg *config.Config) *Policy {
	return &Policy{
		cfg: cfg,
	}
}
