package domain

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"server/internal/utils"
	"slices"
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

type ArchivedTask struct {
	id          uuid.UUID
	kind        string
	payload     json.RawMessage
	createdAt   time.Time
	finalizedAt time.Time
	status      TaskStatus
	priority    int
	retries     int
	result      *json.RawMessage
}

func NewArchivedTask(task *Task) (*ArchivedTask, error) {
	if !slices.Contains([]TaskStatus{TaskStatusCompleted, TaskStatusFailed}, task.Status()) {
		return nil, errors.New("task status not final")
	}

	finalizedAt := task.FinalizedAt()
	if finalizedAt == nil {
		return nil, errors.New("task finalization time unset")
	}

	return &ArchivedTask{
		task.ID(),
		task.Kind(),
		task.Payload(),
		task.CreatedAt(),
		*finalizedAt,
		task.Status(),
		task.Priority(),
		task.Retries(),
		task.Result(),
	}, nil
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

func (t *ArchivedTask) ID() uuid.UUID            { return t.id }
func (t *ArchivedTask) Kind() string             { return t.kind }
func (t *ArchivedTask) Payload() json.RawMessage { return slices.Clone(t.payload) }
func (t *ArchivedTask) CreatedAt() time.Time     { return t.createdAt }
func (t *ArchivedTask) FinalizedAt() time.Time   { return t.finalizedAt }
func (t *ArchivedTask) Status() TaskStatus       { return t.status }
func (t *ArchivedTask) Priority() int            { return t.priority }
func (t *ArchivedTask) Retries() int             { return t.retries }

func (t *ArchivedTask) Result() *json.RawMessage {
	if t.result == nil {
		return t.result
	}
	return utils.P(slices.Clone(*t.result))
}
