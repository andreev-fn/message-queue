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
	DefaultDeadLettering      = true
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

func DefaultDLQueueConfig(parentProcessingTimeout time.Duration) *domain.QueueConfig {
	backoffConf, err := domain.NewBackoffConfig(
		[]time.Duration{time.Minute},
		opt.None[int](), // infinite retries
	)
	if err != nil {
		panic(err)
	}

	// derive timeout from the parent queue, but not less than a minute
	timeout := time.Minute
	if parentProcessingTimeout > timeout {
		timeout = parentProcessingTimeout
	}

	conf, err := domain.NewQueueConfig(
		opt.Some(backoffConf),
		timeout,
		false,
	)
	if err != nil {
		panic(err)
	}

	return conf
}
