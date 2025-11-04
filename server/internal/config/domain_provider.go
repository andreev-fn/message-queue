package config

import (
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
	qConf, err := b.conf.GetQueueConfig(queue)
	if err != nil {
		return nil, err
	}

	return qConf, nil
}
