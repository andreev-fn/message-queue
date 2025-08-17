package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ArchivedMsgDTO supposed to be used only for storage, don't change values manually
type ArchivedMsgDTO struct {
	ID          uuid.UUID
	Kind        string
	Payload     json.RawMessage
	CreatedAt   time.Time
	FinalizedAt time.Time
	Status      MessageStatus
	Priority    int
	Retries     int
	Result      *json.RawMessage
}

func ArchivedMsgFromDTO(dto *ArchivedMsgDTO) *ArchivedMsg {
	return &ArchivedMsg{
		id:          dto.ID,
		kind:        dto.Kind,
		payload:     dto.Payload,
		createdAt:   dto.CreatedAt,
		finalizedAt: dto.FinalizedAt,
		status:      dto.Status,
		priority:    dto.Priority,
		retries:     dto.Retries,
		result:      dto.Result,
	}
}

func (m *ArchivedMsg) ToDTO() *ArchivedMsgDTO {
	return &ArchivedMsgDTO{
		ID:          m.id,
		Kind:        m.kind,
		Payload:     m.payload,
		CreatedAt:   m.createdAt,
		FinalizedAt: m.finalizedAt,
		Status:      m.status,
		Priority:    m.priority,
		Retries:     m.retries,
		Result:      m.result,
	}
}
