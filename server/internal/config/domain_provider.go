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

func (b DomainProvider) GetConfig(queueName string) (*domain.QueueConfig, error) {
	qConf, exist := b.conf.GetQueueConfig(queueName)
	if !exist {
		return nil, fmt.Errorf("queue %s not defined", queueName)
	}

	return qConf, nil
}
