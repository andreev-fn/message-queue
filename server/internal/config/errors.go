package config

import (
	"fmt"

	"server/internal/domain"
)

type QueueNotFoundError struct {
	queue domain.QueueName
}

func newQueueNotFoundError(queue domain.QueueName) QueueNotFoundError {
	return QueueNotFoundError{queue: queue}
}

func (e QueueNotFoundError) Error() string {
	return fmt.Sprintf("queue not found: %s", e.queue)
}
