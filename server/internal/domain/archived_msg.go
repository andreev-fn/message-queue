package domain

import (
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"
)

type ArchivedMsg struct {
	id          uuid.UUID
	queue       string
	payload     string
	createdAt   time.Time
	finalizedAt time.Time
	status      MessageStatus
	priority    int
	retries     int
}

func NewArchivedMsg(msg *Message) (*ArchivedMsg, error) {
	if !slices.Contains([]MessageStatus{MsgStatusDelivered, MsgStatusUndeliverable}, msg.Status()) {
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
	}, nil
}

func (m *ArchivedMsg) ID() uuid.UUID          { return m.id }
func (m *ArchivedMsg) Queue() string          { return m.queue }
func (m *ArchivedMsg) Payload() string        { return m.payload }
func (m *ArchivedMsg) CreatedAt() time.Time   { return m.createdAt }
func (m *ArchivedMsg) FinalizedAt() time.Time { return m.finalizedAt }
func (m *ArchivedMsg) Status() MessageStatus  { return m.status }
func (m *ArchivedMsg) Priority() int          { return m.priority }
func (m *ArchivedMsg) Retries() int           { return m.retries }
