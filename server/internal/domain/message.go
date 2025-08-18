package domain

import (
	"encoding/json"
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"

	"server/internal/utils"
	"server/internal/utils/timeutils"
)

type MessageStatus string

const (
	MsgStatusCreated    MessageStatus = "CREATED"
	MsgStatusReady      MessageStatus = "READY"
	MsgStatusProcessing MessageStatus = "PROCESSING"
	MsgStatusDelayed    MessageStatus = "DELAYED"
	MsgStatusCompleted  MessageStatus = "COMPLETED"
	MsgStatusFailed     MessageStatus = "FAILED"
)

type Message struct {
	id              uuid.UUID
	queue           string
	payload         json.RawMessage
	createdAt       time.Time
	finalizedAt     *time.Time
	status          MessageStatus
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

func NewMessage(
	clock timeutils.Clock,
	id uuid.UUID,
	queue string,
	payload json.RawMessage,
	priority int,
	startAt *time.Time,
) (*Message, error) {
	if startAt != nil && startAt.Before(clock.Now()) {
		return nil, errors.New("start time must be in the future")
	}
	return &Message{
		id:              id,
		queue:           queue,
		payload:         payload,
		createdAt:       clock.Now(),
		finalizedAt:     nil,
		status:          MsgStatusCreated,
		statusChangedAt: clock.Now(),
		delayedUntil:    startAt,
		timeoutAt:       nil,
		priority:        priority,
		retries:         0,
		result:          nil,
		version:         0,
		isNew:           true,
		isResultNew:     false,
	}, nil
}

func (m *Message) ID() uuid.UUID            { return m.id }
func (m *Message) Queue() string            { return m.queue }
func (m *Message) Payload() json.RawMessage { return slices.Clone(m.payload) }
func (m *Message) CreatedAt() time.Time     { return m.createdAt }
func (m *Message) Status() MessageStatus    { return m.status }
func (m *Message) Priority() int            { return m.priority }
func (m *Message) Retries() int             { return m.retries }

func (m *Message) FinalizedAt() *time.Time {
	if m.finalizedAt == nil {
		return nil
	}
	return utils.P(*m.finalizedAt)
}

func (m *Message) Result() *json.RawMessage {
	if m.result == nil {
		return nil
	}
	return utils.P(slices.Clone(*m.result))
}

func (m *Message) Confirm(clock timeutils.Clock, ed EventDispatcher) error {
	if m.status != MsgStatusCreated {
		return errors.New("message must be in CREATED status")
	}

	if m.delayedUntil != nil {
		m.setStatus(clock, MsgStatusDelayed)
	} else {
		m.setStatus(clock, MsgStatusReady)
		ed.Dispatch(NewMsgReadyEvent(m.queue))
	}

	return nil
}

func (m *Message) StartProcessing(clock timeutils.Clock) error {
	if m.status != MsgStatusReady {
		return errors.New("message must be in READY status")
	}

	m.setStatus(clock, MsgStatusProcessing)
	m.timeoutAt = utils.P(clock.Now().Add(5 * time.Minute))

	return nil
}

func (m *Message) Delay(clock timeutils.Clock, delayedUntil time.Time) error {
	if m.status != MsgStatusProcessing {
		return errors.New("message must be in PROCESSING status")
	}

	m.timeoutAt = nil // cleanup after PROCESSING status
	m.retries++

	m.setStatus(clock, MsgStatusDelayed)
	m.delayedUntil = utils.P(delayedUntil)

	return nil
}

func (m *Message) Resume(clock timeutils.Clock, ed EventDispatcher) error {
	if m.status != MsgStatusDelayed {
		return errors.New("message must be in DELAYED status")
	}

	if m.delayedUntil == nil {
		return errors.New("message must have specified delayed_until")
	}
	if clock.Now().Before(*m.delayedUntil) {
		return errors.New("message not ready to be resumed yet")
	}

	m.delayedUntil = nil // cleanup after DELAYED status

	m.setStatus(clock, MsgStatusReady)
	ed.Dispatch(NewMsgReadyEvent(m.queue))

	return nil
}

func (m *Message) Complete(clock timeutils.Clock, result json.RawMessage) error {
	if m.status != MsgStatusProcessing {
		return errors.New("message must be in PROCESSING status")
	}

	m.timeoutAt = nil // cleanup after PROCESSING status

	m.setStatus(clock, MsgStatusCompleted)
	m.finalizedAt = utils.P(clock.Now())
	m.result = utils.P(result)
	m.isResultNew = true

	return nil
}

func (m *Message) Fail(clock timeutils.Clock) error {
	if m.status != MsgStatusProcessing {
		return errors.New("message must be in PROCESSING status")
	}

	m.timeoutAt = nil // cleanup after PROCESSING status

	m.setStatus(clock, MsgStatusFailed)
	m.finalizedAt = utils.P(clock.Now())

	return nil
}

func (m *Message) setStatus(clock timeutils.Clock, newStatus MessageStatus) {
	m.status = newStatus
	m.statusChangedAt = clock.Now()
}
