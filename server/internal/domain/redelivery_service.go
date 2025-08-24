package domain

import (
	"time"

	"server/internal/utils/timeutils"
)

const maxRetries = 10

type RedeliveryService struct {
	clock timeutils.Clock
}

func NewRedeliveryService(clock timeutils.Clock) *RedeliveryService {
	return &RedeliveryService{clock}
}

func (eh *RedeliveryService) HandleNack(msg *Message) error {
	if msg.Retries() >= maxRetries {
		return msg.MarkUndeliverable(eh.clock)
	}

	return msg.Delay(eh.clock, eh.getDelayTime(msg.Retries()))
}

func (eh *RedeliveryService) getDelayTime(retries int) time.Time {
	if retries < 5 {
		return eh.clock.Now().Add(2 * time.Minute)
	}
	if retries < 10 {
		return eh.clock.Now().Add(10 * time.Minute)
	}
	if retries < 50 {
		return eh.clock.Now().Add(30 * time.Minute)
	}
	return eh.clock.Now().Add(3 * time.Hour)
}
