package domain

import (
	"errors"
	"slices"
	"time"

	"server/internal/utils/opt"
)

type QueueConfig struct {
	backoff           opt.Val[*BackoffConfig]
	processingTimeout time.Duration
	deadLetteringOn   bool
}

func NewQueueConfig(
	backoff opt.Val[*BackoffConfig],
	processingTimeout time.Duration,
	deadLetteringOn bool,
) (*QueueConfig, error) {
	if processingTimeout < time.Second {
		return nil, errors.New("processing timeout must be at least 1 second")
	}

	return &QueueConfig{
		backoff:           backoff,
		processingTimeout: processingTimeout,
		deadLetteringOn:   deadLetteringOn,
	}, nil
}

func (c *QueueConfig) Backoff() opt.Val[*BackoffConfig] { return c.backoff }
func (c *QueueConfig) ProcessingTimeout() time.Duration { return c.processingTimeout }
func (c *QueueConfig) IsDeadLetteringOn() bool          { return c.deadLetteringOn }

type BackoffConfig struct {
	shape       []time.Duration
	maxAttempts opt.Val[int]
}

func NewBackoffConfig(
	shape []time.Duration,
	maxAttempts opt.Val[int],
) (*BackoffConfig, error) {
	if len(shape) < 1 {
		return nil, errors.New("shape must have at least one element")
	}

	if value, isSet := maxAttempts.Value(); isSet && value <= 0 {
		return nil, errors.New("max attempts must be greater than zero if provided")
	}

	return &BackoffConfig{
		shape:       shape,
		maxAttempts: maxAttempts,
	}, nil
}

func (c *BackoffConfig) Shape() []time.Duration    { return slices.Clone(c.shape) }
func (c *BackoffConfig) MaxAttempts() opt.Val[int] { return c.maxAttempts }
