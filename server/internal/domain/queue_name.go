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

	parts := strings.Split(name, ":")
	if len(parts) > 2 {
		return QueueName{}, errors.New("queue name can have at most one colon (':') separator")
	}
	if len(parts) == 2 && parts[1] != "dl" {
		return QueueName{}, errors.New("the only possible special queue kind is ':dl'")
	}
	namePart := parts[0]

	// reserve 15 characters for suffixes
	if len(namePart) > 240 {
		return QueueName{}, errors.New("queue name too long")
	}

	for _, part := range strings.Split(namePart, ".") {
		if part == "" {
			return QueueName{}, errors.New("queue name parts can't be empty")
		}

		for _, c := range part {
			if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c == '_') {
				return QueueName{}, errors.New("queue name can only contain alphanumeric characters and '_'")
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

func (q QueueName) IsDLQ() bool {
	return strings.HasSuffix(q.v, ":dl")
}

func (q QueueName) DLQName() (QueueName, error) {
	if q.IsDLQ() {
		return QueueName{}, errors.New("DLQ cannot have DLQ")
	}
	return QueueName{v: q.v + ":dl"}, nil
}
