package domain

import (
	"time"

	"server/internal/utils/timeutils"
)

type ConfigProvider interface {
	GetConfig(queue QueueName) (*QueueConfig, error)
}

type NackActionKind int

const (
	NackActionDelay NackActionKind = 1 << iota
	NackActionDrop
	NackActionDLQ
)

type NackAction struct {
	Type          NackActionKind
	DelayDuration time.Duration // only valid for NackActionDelay
}

type NackPolicy struct {
	clock          timeutils.Clock
	configProvider ConfigProvider
}

func NewNackPolicy(
	clock timeutils.Clock,
	configProvider ConfigProvider,
) *NackPolicy {
	return &NackPolicy{
		clock:          clock,
		configProvider: configProvider,
	}
}

func (eh *NackPolicy) Decide(msg *Message, redeliveryRequested bool) (*NackAction, error) {
	conf, err := eh.configProvider.GetConfig(msg.Queue())
	if err != nil {
		return nil, err
	}

	return pureDecide(msg.Retries(), conf, redeliveryRequested), nil
}

func pureDecide(msgRetries int, conf *QueueConfig, redeliveryRequested bool) *NackAction {
	backoffConf, backoffIsSet := conf.Backoff().Value()

	if !redeliveryRequested || !backoffIsSet {
		return handleExhausted(conf)
	}

	if maxAttempt, isSet := backoffConf.MaxAttempts().Value(); isSet && msgRetries >= maxAttempt {
		return handleExhausted(conf)
	}

	duration := getDelayDuration(backoffConf.Shape(), msgRetries)

	return &NackAction{
		Type:          NackActionDelay,
		DelayDuration: duration,
	}
}

func handleExhausted(conf *QueueConfig) *NackAction {
	if conf.IsDeadLetteringOn() {
		return &NackAction{Type: NackActionDLQ}
	}
	return &NackAction{Type: NackActionDrop}
}

func getDelayDuration(shape []time.Duration, retries int) time.Duration {
	for i, dur := range shape {
		if i == retries {
			return dur
		}
	}
	return shape[len(shape)-1]
}
