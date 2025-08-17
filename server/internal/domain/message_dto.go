package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// MessageDTO supposed to be used only for storage, don't change values manually
type MessageDTO struct {
	ID              uuid.UUID
	Kind            string
	Payload         json.RawMessage
	CreatedAt       time.Time
	FinalizedAt     *time.Time
	Status          MessageStatus
	StatusChangedAt time.Time
	DelayedUntil    *time.Time
	TimeoutAt       *time.Time
	Priority        int
	Retries         int
	Result          *json.RawMessage
	Version         int
	IsNew           bool
	IsResultNew     bool
}

func FromDTO(dto *MessageDTO) *Message {
	return &Message{
		id:              dto.ID,
		kind:            dto.Kind,
		payload:         dto.Payload,
		createdAt:       dto.CreatedAt,
		finalizedAt:     dto.FinalizedAt,
		status:          dto.Status,
		statusChangedAt: dto.StatusChangedAt,
		delayedUntil:    dto.DelayedUntil,
		timeoutAt:       dto.TimeoutAt,
		priority:        dto.Priority,
		retries:         dto.Retries,
		result:          dto.Result,
		version:         dto.Version,
		isNew:           dto.IsNew,
		isResultNew:     dto.IsResultNew,
	}
}

func (m *Message) ToDTO() *MessageDTO {
	return &MessageDTO{
		ID:              m.id,
		Kind:            m.kind,
		Payload:         m.payload,
		CreatedAt:       m.createdAt,
		FinalizedAt:     m.finalizedAt,
		Status:          m.status,
		StatusChangedAt: m.statusChangedAt,
		DelayedUntil:    m.delayedUntil,
		TimeoutAt:       m.timeoutAt,
		Priority:        m.priority,
		Retries:         m.retries,
		Result:          m.result,
		Version:         m.version,
		IsNew:           m.isNew,
		IsResultNew:     m.isResultNew,
	}
}
