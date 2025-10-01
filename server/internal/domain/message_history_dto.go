package domain

import (
	"time"

	"github.com/google/uuid"
)

func historyFromDTO(dto []*MessageChapterDTO) *MessageHistory {
	if dto == nil {
		return newMessageHistory(false)
	}

	chapters := make([]*MessageChapter, 0, len(dto))
	for _, chDTO := range dto {
		chapters = append(chapters, chapterFromDTO(chDTO))
	}

	return &MessageHistory{
		isLoaded: true,
		chapters: chapters,
	}
}

func (h *MessageHistory) toDTO() []*MessageChapterDTO {
	dto := make([]*MessageChapterDTO, 0, len(h.chapters))
	for _, ch := range h.chapters {
		dto = append(dto, ch.toDTO())
	}
	return dto
}

type MessageChapterDTO struct {
	MsgID        uuid.UUID
	Generation   int
	Queue        string
	RedirectedAt time.Time
	Priority     int
	Retries      int
	IsNew        bool
}

func chapterFromDTO(dto *MessageChapterDTO) *MessageChapter {
	return &MessageChapter{
		msgID:        dto.MsgID,
		generation:   dto.Generation,
		queue:        UnsafeQueueName(dto.Queue),
		redirectedAt: dto.RedirectedAt,
		priority:     dto.Priority,
		retries:      dto.Retries,
		isNew:        dto.IsNew,
	}
}

func (c *MessageChapter) toDTO() *MessageChapterDTO {
	return &MessageChapterDTO{
		MsgID:        c.msgID,
		Generation:   c.generation,
		Queue:        c.queue.String(),
		RedirectedAt: c.redirectedAt,
		Priority:     c.priority,
		Retries:      c.retries,
		IsNew:        c.isNew,
	}
}
