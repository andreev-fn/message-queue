package domain

import (
	"server/internal/utils/timeutils"
	"time"
)

const maxRetries = 10

type ErrorHandlers interface {
	HandleError(task *Task, errorCode string) error
}

type ExponentialErrorHandler struct {
	clock timeutils.Clock
}

func NewExponentialErrorHandler(clock timeutils.Clock) *ExponentialErrorHandler {
	return &ExponentialErrorHandler{clock}
}

func (eh *ExponentialErrorHandler) HandleError(task *Task, errorCode string) error {
	if task.Retries() >= maxRetries {
		return task.Fail(eh.clock)
	}

	return task.Delay(eh.clock, eh.getDelayTime(task.Retries()))
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
