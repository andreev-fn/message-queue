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

func (t *Task) ToDTO() *TaskDTO {
	return &TaskDTO{
		ID:              t.id,
		Kind:            t.kind,
		Payload:         t.payload,
		CreatedAt:       t.createdAt,
		FinalizedAt:     t.finalizedAt,
		Status:          t.status,
		StatusChangedAt: t.statusChangedAt,
		DelayedUntil:    t.delayedUntil,
		TimeoutAt:       t.timeoutAt,
		Priority:        t.priority,
		Retries:         t.retries,
		Result:          t.result,
		Version:         t.version,
		IsNew:           t.isNew,
		IsResultNew:     t.isResultNew,
	}
}
