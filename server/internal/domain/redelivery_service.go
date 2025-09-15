package domain

import (
	"time"

	"server/internal/utils/opt"
	"server/internal/utils/timeutils"
)

type BackoffConfig interface {
	Shape() []time.Duration
	MaxAttempts() opt.Val[int]
}

type BackoffConfigProvider interface {
	GetConfig(queueName string) (BackoffConfig, error)
}

type RedeliveryService struct {
	clock          timeutils.Clock
	configProvider BackoffConfigProvider
}

func NewRedeliveryService(
	clock timeutils.Clock,
	configProvider BackoffConfigProvider,
) *RedeliveryService {
	return &RedeliveryService{
		clock:          clock,
		configProvider: configProvider,
	}
}

func (eh *RedeliveryService) HandleNack(msg *Message) error {
	conf, err := eh.configProvider.GetConfig(msg.Queue())
	if err != nil {
		return err
	}

	if conf.MaxAttempts().IsSet() && msg.Retries() >= conf.MaxAttempts().MustValue() {
		return msg.MarkUndeliverable(eh.clock)
	}

	duration := getDelayDuration(conf.Shape(), msg.Retries())

	return msg.Delay(eh.clock, eh.clock.Now().Add(duration))
}

func getDelayDuration(shape []time.Duration, retries int) time.Duration {
	for i, dur := range shape {
		if i == retries {
			return dur
		}
	}
	return shape[len(shape)-1]
}
