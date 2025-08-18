package domain

import (
	"encoding/json"
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"

	"server/internal/utils"
)

type ArchivedMsg struct {
	id          uuid.UUID
	queue       string
	payload     json.RawMessage
	createdAt   time.Time
	finalizedAt time.Time
	status      MessageStatus
	priority    int
	retries     int
	result      *json.RawMessage
}

func NewArchivedMsg(msg *Message) (*ArchivedMsg, error) {
	if !slices.Contains([]MessageStatus{MsgStatusCompleted, MsgStatusFailed}, msg.Status()) {
		return nil, errors.New("message status not final")
	}

	finalizedAt := msg.FinalizedAt()
	if finalizedAt == nil {
		return nil, errors.New("message finalization time unset")
	}

	return &ArchivedMsg{
		id:          msg.ID(),
		queue:       msg.Queue(),
		payload:     msg.Payload(),
		createdAt:   msg.CreatedAt(),
		finalizedAt: *finalizedAt,
		status:      msg.Status(),
		priority:    msg.Priority(),
		retries:     msg.Retries(),
		result:      msg.Result(),
	}, nil
}

func (m *ArchivedMsg) ID() uuid.UUID            { return m.id }
func (m *ArchivedMsg) Queue() string            { return m.queue }
func (m *ArchivedMsg) Payload() json.RawMessage { return slices.Clone(m.payload) }
func (m *ArchivedMsg) CreatedAt() time.Time     { return m.createdAt }
func (m *ArchivedMsg) FinalizedAt() time.Time   { return m.finalizedAt }
func (m *ArchivedMsg) Status() MessageStatus    { return m.status }
func (m *ArchivedMsg) Priority() int            { return m.priority }
func (m *ArchivedMsg) Retries() int             { return m.retries }

func (m *ArchivedMsg) Result() *json.RawMessage {
	if m.result == nil {
		return m.result
	}
	return utils.P(slices.Clone(*m.result))
}
