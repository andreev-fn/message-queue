package domain

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"server/internal/utils"
	"server/internal/utils/timeutils"
	"slices"
	"time"
)

type TaskStatus string

const (
	TaskStatusCreated    TaskStatus = "CREATED"
	TaskStatusReady      TaskStatus = "READY"
	TaskStatusProcessing TaskStatus = "PROCESSING"
	TaskStatusDelayed    TaskStatus = "DELAYED"
	TaskStatusCompleted  TaskStatus = "COMPLETED"
	TaskStatusFailed     TaskStatus = "FAILED"
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

type Task struct {
	id              uuid.UUID
	kind            string
	payload         json.RawMessage
	createdAt       time.Time
	finalizedAt     *time.Time
	status          TaskStatus
	statusChangedAt time.Time
	delayedUntil    *time.Time
	timeoutAt       *time.Time
	priority        int
	retries         int
	result          *json.RawMessage

	version     int  // for optimistic locking
	isNew       bool // to distinguish between insert and update
	isResultNew bool // save the result only when necessary
}

func NewTask(
	clock timeutils.Clock,
	id uuid.UUID,
	kind string,
	payload json.RawMessage,
	priority int,
	startAt *time.Time,
) (*Task, error) {
	if startAt != nil && startAt.Before(clock.Now()) {
		return nil, errors.New("start time must be in the future")
	}
	return &Task{
		id,
		kind,
		payload,
		clock.Now(),
		nil,
		TaskStatusCreated,
		clock.Now(),
		startAt,
		nil,
		priority,
		0,
		nil,
		0,
		true,
		false,
	}, nil
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

func (t *Task) ID() uuid.UUID            { return t.id }
func (t *Task) Kind() string             { return t.kind }
func (t *Task) Payload() json.RawMessage { return slices.Clone(t.payload) }
func (t *Task) CreatedAt() time.Time     { return t.createdAt }
func (t *Task) Status() TaskStatus       { return t.status }
func (t *Task) Priority() int            { return t.priority }
func (t *Task) Retries() int             { return t.retries }

func (t *Task) FinalizedAt() *time.Time {
	if t.finalizedAt == nil {
		return nil
	}
	return utils.P(*t.finalizedAt)
}

func (t *Task) Result() *json.RawMessage {
	if t.result == nil {
		return nil
	}
	return utils.P(slices.Clone(*t.result))
}

func (t *Task) Confirm(clock timeutils.Clock) error {
	if t.status != TaskStatusCreated {
		return errors.New("task must be in CREATED status")
	}

	if t.delayedUntil != nil {
		t.setStatus(clock, TaskStatusDelayed)
	} else {
		t.setStatus(clock, TaskStatusReady)
	}

	return nil
}

func (t *Task) StartProcessing(clock timeutils.Clock) error {
	if t.status != TaskStatusReady {
		return errors.New("task must be in READY status")
	}

	t.setStatus(clock, TaskStatusProcessing)
	t.timeoutAt = utils.P(clock.Now().Add(5 * time.Minute))

	return nil
}

func (t *Task) Delay(clock timeutils.Clock, delayedUntil time.Time) error {
	if t.status != TaskStatusProcessing {
		return errors.New("task must be in PROCESSING status")
	}

	t.timeoutAt = nil // cleanup after PROCESSING status
	t.retries++

	t.setStatus(clock, TaskStatusDelayed)
	t.delayedUntil = utils.P(delayedUntil)

	return nil
}

func (t *Task) Resume(clock timeutils.Clock) error {
	if t.status != TaskStatusDelayed {
		return errors.New("task must be in DELAYED status")
	}

	if t.delayedUntil == nil {
		return errors.New("task must have specified delayed_until")
	}
	if clock.Now().Before(*t.delayedUntil) {
		return errors.New("task not ready to be resumed yet")
	}

	t.delayedUntil = nil // cleanup after DELAYED status

	t.setStatus(clock, TaskStatusReady)

	return nil
}

func (t *Task) Complete(clock timeutils.Clock, result json.RawMessage) error {
	if t.status != TaskStatusProcessing {
		return errors.New("task must be in PROCESSING status")
	}

	t.timeoutAt = nil // cleanup after PROCESSING status

	t.setStatus(clock, TaskStatusCompleted)
	t.finalizedAt = utils.P(clock.Now())
	t.result = utils.P(result)
	t.isResultNew = true

	return nil
}

func (t *Task) Fail(clock timeutils.Clock) error {
	if t.status != TaskStatusProcessing {
		return errors.New("task must be in PROCESSING status")
	}

	t.timeoutAt = nil // cleanup after PROCESSING status

	t.setStatus(clock, TaskStatusFailed)
	t.finalizedAt = utils.P(clock.Now())

	return nil
}

func (t *Task) setStatus(clock timeutils.Clock, newStatus TaskStatus) {
	t.status = newStatus
	t.statusChangedAt = clock.Now()
}
