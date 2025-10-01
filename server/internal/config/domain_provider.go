package config

import (
	"fmt"

	"server/internal/domain"
)

type DomainProvider struct {
	conf *Config
}

var _ domain.ConfigProvider = DomainProvider{}

func NewDomainProvider(conf *Config) DomainProvider {
	return DomainProvider{conf}
}

func (b DomainProvider) GetConfig(queue domain.QueueName) (*domain.QueueConfig, error) {
	qConf, exist := b.conf.GetQueueConfig(queue)
	if !exist {
		return nil, fmt.Errorf("queue not defined: %s", queue.String())
	}

	return qConf, nil
}
