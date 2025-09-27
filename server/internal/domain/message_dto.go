package domain

import (
	"time"

	"github.com/google/uuid"
)

// MessageDTO supposed to be used only for storage, don't change values manually
type MessageDTO struct {
	ID              uuid.UUID
	Queue           string
	Payload         string
	CreatedAt       time.Time
	FinalizedAt     *time.Time
	Status          MessageStatus
	StatusChangedAt time.Time
	DelayedUntil    *time.Time
	TimeoutAt       *time.Time
	Priority        int
	Retries         int
	Version         int
	IsNew           bool
}

func FromDTO(dto *MessageDTO) *Message {
	return &Message{
		id:              dto.ID,
		queue:           UnsafeQueueName(dto.Queue),
		payload:         dto.Payload,
		createdAt:       dto.CreatedAt,
		finalizedAt:     dto.FinalizedAt,
		status:          dto.Status,
		statusChangedAt: dto.StatusChangedAt,
		delayedUntil:    dto.DelayedUntil,
		timeoutAt:       dto.TimeoutAt,
		priority:        dto.Priority,
		retries:         dto.Retries,
		version:         dto.Version,
		isNew:           dto.IsNew,
	}
}

func (m *Message) ToDTO() *MessageDTO {
	return &MessageDTO{
		ID:              m.id,
		Queue:           m.queue.String(),
		Payload:         m.payload,
		CreatedAt:       m.createdAt,
		FinalizedAt:     m.finalizedAt,
		Status:          m.status,
		StatusChangedAt: m.statusChangedAt,
		DelayedUntil:    m.delayedUntil,
		TimeoutAt:       m.timeoutAt,
		Priority:        m.priority,
		Retries:         m.retries,
		Version:         m.version,
		IsNew:           m.isNew,
	}
}
