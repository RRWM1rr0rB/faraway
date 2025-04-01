package mitigator

import (
	"app-server/app/internal/config"
	"app-server/app/internal/policy"
)

type Policy struct {
	*policy.BasePolicy
	service Service

	cfg *config.Config
}

func NewPolicy(
	basePolicy *policy.BasePolicy,
	service Service,
	cfg *config.Config) *Policy {
	return &Policy{
		BasePolicy: basePolicy,
		service:    service,
		cfg:        cfg,
	}
}
