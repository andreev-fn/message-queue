package domain

import (
	"time"

	"github.com/google/uuid"
)

// ArchivedMsgDTO supposed to be used only for storage, don't change values manually
type ArchivedMsgDTO struct {
	ID          uuid.UUID
	Queue       string
	Payload     string
	CreatedAt   time.Time
	FinalizedAt time.Time
	Status      MessageStatus
	Priority    int
	Retries     int
	Generation  int
	History     []ArchivedChapterDTO
}

// Warning! It's not safe to rename fields of ArchivedChapterDTO,
// because it's encoded to JSON as-is without additional mapping.

type ArchivedChapterDTO struct {
	Generation   int
	Queue        string
	RedirectedAt time.Time
	Priority     int
	Retries      int
}

func ArchivedMsgFromDTO(dto *ArchivedMsgDTO) *ArchivedMsg {
	chapters := make([]*ArchivedChapter, 0, len(dto.History))
	for _, chapterDTO := range dto.History {
		chapters = append(chapters, &ArchivedChapter{
			generation:   chapterDTO.Generation,
			queue:        UnsafeQueueName(chapterDTO.Queue),
			redirectedAt: chapterDTO.RedirectedAt,
			priority:     chapterDTO.Priority,
			retries:      chapterDTO.Retries,
		})
	}
	return &ArchivedMsg{
		id:          dto.ID,
		queue:       UnsafeQueueName(dto.Queue),
		payload:     dto.Payload,
		createdAt:   dto.CreatedAt,
		finalizedAt: dto.FinalizedAt,
		status:      dto.Status,
		priority:    dto.Priority,
		retries:     dto.Retries,
		generation:  dto.Generation,
		history:     chapters,
	}
}

func (m *ArchivedMsg) ToDTO() *ArchivedMsgDTO {
	chapterDTOs := make([]ArchivedChapterDTO, 0, len(m.history))
	for _, chapter := range m.history {
		chapterDTOs = append(chapterDTOs, ArchivedChapterDTO{
			Generation:   chapter.generation,
			Queue:        chapter.queue.String(),
			RedirectedAt: chapter.redirectedAt,
			Priority:     chapter.priority,
			Retries:      chapter.retries,
		})
	}
	return &ArchivedMsgDTO{
		ID:          m.id,
		Queue:       m.queue.String(),
		Payload:     m.payload,
		CreatedAt:   m.createdAt,
		FinalizedAt: m.finalizedAt,
		Status:      m.status,
		Priority:    m.priority,
		Retries:     m.retries,
		Generation:  m.generation,
		History:     chapterDTOs,
	}
}
