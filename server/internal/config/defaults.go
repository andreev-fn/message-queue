package config

import (
	"time"

	"server/internal/domain"
	"server/internal/utils/opt"
)

const (
	DefaultBatchSizeLimit     = 200
	DefaultBackoffEnabled     = true
	DefaultBackoffMaxAttempts = 5
)

func DefaultBackoffShape() []time.Duration {
	return []time.Duration{
		1 * time.Second,
		5 * time.Second,
		30 * time.Second,
		1 * time.Minute,
		5 * time.Minute,
	}
}

func DefaultBackoffConfig() *domain.BackoffConfig {
	cfg, err := domain.NewBackoffConfig(
		DefaultBackoffShape(),
		opt.Some(DefaultBackoffMaxAttempts),
	)
	if err != nil {
		panic(err)
	}
	return cfg
}
