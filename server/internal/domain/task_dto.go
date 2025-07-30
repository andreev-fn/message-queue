package domain

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

// TaskDTO supposed to be used only for storage, don't change values manually
type TaskDTO struct {
	ID              uuid.UUID
	Kind            string
	Payload         json.RawMessage
	CreatedAt       time.Time
	FinalizedAt     *time.Time
	Status          TaskStatus
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

func FromDTO(dto *TaskDTO) *Task {
	return &Task{
		dto.ID,
		dto.Kind,
		dto.Payload,
		dto.CreatedAt,
		dto.FinalizedAt,
		dto.Status,
		dto.StatusChangedAt,
		dto.DelayedUntil,
		dto.TimeoutAt,
		dto.Priority,
		dto.Retries,
		dto.Result,
		dto.Version,
		dto.IsNew,
		dto.IsResultNew,
	}
}

func (t *Task) ToDTO() *TaskDTO {
	return &TaskDTO{
		t.id,
		t.kind,
		t.payload,
		t.createdAt,
		t.finalizedAt,
		t.status,
		t.statusChangedAt,
		t.delayedUntil,
		t.timeoutAt,
		t.priority,
		t.retries,
		t.result,
		t.version,
		t.isNew,
		t.isResultNew,
	}
}
