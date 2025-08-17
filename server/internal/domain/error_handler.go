package domain

import (
	"time"

	"server/internal/utils/timeutils"
)

const maxRetries = 10

type ErrorHandlers interface {
	HandleError(msg *Message, errorCode string) error
}

type ExponentialErrorHandler struct {
	clock timeutils.Clock
}

func NewExponentialErrorHandler(clock timeutils.Clock) *ExponentialErrorHandler {
	return &ExponentialErrorHandler{clock}
}

func (eh *ExponentialErrorHandler) HandleError(msg *Message, errorCode string) error {
	if msg.Retries() >= maxRetries {
		return msg.Fail(eh.clock)
	}

	return msg.Delay(eh.clock, eh.getDelayTime(msg.Retries()))
}

func (eh *ExponentialErrorHandler) getDelayTime(retries int) time.Time {
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
