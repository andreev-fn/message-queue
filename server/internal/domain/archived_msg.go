package domain

import (
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"
)

type ArchivedMsg struct {
	id          uuid.UUID
	queue       QueueName
	payload     string
	createdAt   time.Time
	finalizedAt time.Time
	status      MessageStatus
	priority    int
	retries     int
	generation  int
	history     []*ArchivedChapter
}

type ArchivedChapter struct {
	generation   int
	queue        QueueName
	redirectedAt time.Time
	priority     int
	retries      int
}

func NewArchivedMsg(msg *Message) (*ArchivedMsg, error) {
	if !slices.Contains([]MessageStatus{MsgStatusDelivered, MsgStatusUndeliverable}, msg.Status()) {
		return nil, errors.New("message status not final")
	}

	finalizedAt := msg.FinalizedAt()
	if finalizedAt == nil {
		return nil, errors.New("message finalization time unset")
	}

	chapters, loaded := msg.History().Chapters()
	if !loaded {
		return nil, errors.New("message history not loaded")
	}

	archChapters := make([]*ArchivedChapter, 0, len(chapters))
	for _, chapter := range chapters {
		archChapters = append(archChapters, &ArchivedChapter{
			generation:   chapter.Generation(),
			queue:        chapter.Queue(),
			redirectedAt: chapter.RedirectedAt(),
			priority:     chapter.Priority(),
			retries:      chapter.Retries(),
		})
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
		generation:  msg.Generation(),
		history:     archChapters,
	}, nil
}

func (m *ArchivedMsg) ID() uuid.UUID               { return m.id }
func (m *ArchivedMsg) Queue() QueueName            { return m.queue }
func (m *ArchivedMsg) Payload() string             { return m.payload }
func (m *ArchivedMsg) CreatedAt() time.Time        { return m.createdAt }
func (m *ArchivedMsg) FinalizedAt() time.Time      { return m.finalizedAt }
func (m *ArchivedMsg) Status() MessageStatus       { return m.status }
func (m *ArchivedMsg) Priority() int               { return m.priority }
func (m *ArchivedMsg) Retries() int                { return m.retries }
func (m *ArchivedMsg) Generation() int             { return m.generation }
func (m *ArchivedMsg) History() []*ArchivedChapter { return m.history }

func (c *ArchivedChapter) Generation() int         { return c.generation }
func (c *ArchivedChapter) Queue() QueueName        { return c.queue }
func (c *ArchivedChapter) RedirectedAt() time.Time { return c.redirectedAt }
func (c *ArchivedChapter) Priority() int           { return c.priority }
func (c *ArchivedChapter) Retries() int            { return c.retries }
