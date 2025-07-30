package domain

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

// ArchivedTaskDTO supposed to be used only for storage, don't change values manually
type ArchivedTaskDTO struct {
	ID          uuid.UUID
	Kind        string
	Payload     json.RawMessage
	CreatedAt   time.Time
	FinalizedAt time.Time
	Status      TaskStatus
	Priority    int
	Retries     int
	Result      *json.RawMessage
}

func ArchivedTaskFromDTO(dto *ArchivedTaskDTO) *ArchivedTask {
	return &ArchivedTask{
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

func (t *ArchivedTask) ToDTO() *ArchivedTaskDTO {
	return &ArchivedTaskDTO{
		ID:          t.id,
		Kind:        t.kind,
		Payload:     t.payload,
		CreatedAt:   t.createdAt,
		FinalizedAt: t.finalizedAt,
		Status:      t.status,
		Priority:    t.priority,
		Retries:     t.retries,
		Result:      t.result,
	}
}
