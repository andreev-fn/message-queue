package domain

import (
	"time"

	"github.com/google/uuid"

	"server/internal/utils/timeutils"
)

type MessageHistory struct {
	isLoaded bool
	chapters []*MessageChapter
}

func newMessageHistory(loaded bool) *MessageHistory {
	return &MessageHistory{
		isLoaded: loaded,
		chapters: make([]*MessageChapter, 0),
	}
}

func (h *MessageHistory) Chapters() ([]*MessageChapter, bool) {
	if !h.isLoaded {
		return nil, false
	}
	return h.chapters, true
}

func (h *MessageHistory) addChapter(ch *MessageChapter) {
	h.chapters = append(h.chapters, ch)
}

type MessageChapter struct {
	msgID        uuid.UUID
	generation   int
	queue        QueueName
	redirectedAt time.Time
	priority     int
	retries      int

	isNew bool // to save only new chapters
}

func newChapterFromMessage(clock timeutils.Clock, msg *Message) *MessageChapter {
	return &MessageChapter{
		msgID:        msg.id,
		generation:   msg.generation,
		queue:        msg.queue,
		redirectedAt: clock.Now(),
		priority:     msg.priority,
		retries:      msg.retries,
		isNew:        true,
	}
}

func (c *MessageChapter) Generation() int         { return c.generation }
func (c *MessageChapter) Queue() QueueName        { return c.queue }
func (c *MessageChapter) RedirectedAt() time.Time { return c.redirectedAt }
func (c *MessageChapter) Priority() int           { return c.priority }
func (c *MessageChapter) Retries() int            { return c.retries }
