package domain

import (
	"errors"
	"strings"
)

type QueueName struct {
	v string
}

func NewQueueName(name string) (QueueName, error) {
	if len(name) > 255 {
		return QueueName{}, errors.New("queue name too long")
	}

	for _, part := range strings.Split(name, ".") {
		if part == "" {
			return QueueName{}, errors.New("queue name parts can't be empty")
		}

		for _, c := range part {
			if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'z') {
				return QueueName{}, errors.New("queue name can only contain alphanumeric characters")
			}
		}
	}

	return QueueName{v: name}, nil
}

func UnsafeQueueName(name string) QueueName {
	return QueueName{v: name}
}

func (q QueueName) String() string {
	return q.v
}
