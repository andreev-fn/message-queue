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
		dto.ID,
		dto.Kind,
		dto.Payload,
		dto.CreatedAt,
		dto.FinalizedAt,
		dto.Status,
		dto.Priority,
		dto.Retries,
		dto.Result,
	}
}

func (t *ArchivedTask) ToDTO() *ArchivedTaskDTO {
	return &ArchivedTaskDTO{
		t.id,
		t.kind,
		t.payload,
		t.createdAt,
		t.finalizedAt,
		t.status,
		t.priority,
		t.retries,
		t.result,
	}
}
