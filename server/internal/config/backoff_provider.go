package config

import (
	"fmt"

	"server/internal/domain"
)

type BackoffConfigProvider struct {
	conf *Config
}

var _ domain.BackoffConfigProvider = BackoffConfigProvider{}

func NewBackoffConfigProvider(conf *Config) BackoffConfigProvider {
	return BackoffConfigProvider{conf}
}

func (b BackoffConfigProvider) GetConfig(queueName string) (domain.BackoffConfig, error) {
	if !b.conf.IsQueueDefined(queueName) {
		return nil, fmt.Errorf("queue %s not defined", queueName)
	}

	result, ok := b.conf.QueueConfig(queueName).Backoff().Value()
	if !ok {
		return nil, fmt.Errorf("backoff disabled for queue %s", queueName)
	}

	return result, nil
}
