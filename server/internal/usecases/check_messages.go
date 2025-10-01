package usecases

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"server/internal/config"
	"server/internal/domain"
	"server/internal/storage"
	"server/internal/utils"
)

type CheckMsgResult struct {
	ID          string
	Queue       domain.QueueName
	CreatedAt   time.Time
	FinalizedAt *time.Time
	Status      string
	Priority    int
	Retries     int
	Generation  int
	Payload     string
	History     []CheckMsgChapter
}

type CheckMsgChapter struct {
	Generation   int
	Queue        domain.QueueName
	RedirectedAt time.Time
	Priority     int
	Retries      int
}

type CheckMessages struct {
	db              *sql.DB
	msgRepo         *storage.MessageRepository
	archivedMsgRepo *storage.ArchivedMsgRepository
	conf            *config.Config
}

func NewCheckMessages(
	db *sql.DB,
	msgRepo *storage.MessageRepository,
	archivedMsgRepo *storage.ArchivedMsgRepository,
	conf *config.Config,
) *CheckMessages {
	return &CheckMessages{
		db:              db,
		msgRepo:         msgRepo,
		archivedMsgRepo: archivedMsgRepo,
		conf:            conf,
	}
}

func (uc *CheckMessages) Do(ctx context.Context, ids []string) ([]CheckMsgResult, error) {
	if len(ids) > uc.conf.BatchSizeLimit() {
		return []CheckMsgResult{}, errors.New("batch size limit exceeded")
	}

	allResults := make([]CheckMsgResult, 0, len(ids))

	for _, id := range ids {
		result, err := uc.checkMessage(ctx, id)
		if err != nil {
			return nil, err
		}

		allResults = append(allResults, result)
	}
	return allResults, nil
}

func (uc *CheckMessages) checkMessage(ctx context.Context, id string) (CheckMsgResult, error) {
	message, err := uc.msgRepo.GetByIDWithHistory(ctx, uc.db, id)
	if err != nil {
		if errors.Is(err, storage.ErrMsgNotFound) {
			return uc.checkArchived(ctx, id)
		}
		return CheckMsgResult{}, fmt.Errorf("msgRepo.GetByID: %w", err)
	}

	chapters, loaded := message.History().Chapters()
	if !loaded {
		return CheckMsgResult{}, errors.New("logic error: message history must be loaded")
	}

	mappedChapters := make([]CheckMsgChapter, 0, len(chapters))
	for _, chapter := range chapters {
		mappedChapters = append(mappedChapters, CheckMsgChapter{
			Generation:   chapter.Generation(),
			Queue:        chapter.Queue(),
			RedirectedAt: chapter.RedirectedAt(),
			Priority:     chapter.Priority(),
			Retries:      chapter.Retries(),
		})
	}

	return CheckMsgResult{
		ID:          message.ID().String(),
		Queue:       message.Queue(),
		Payload:     message.Payload(),
		CreatedAt:   message.CreatedAt(),
		FinalizedAt: message.FinalizedAt(),
		Status:      string(message.Status()),
		Priority:    message.Priority(),
		Retries:     message.Retries(),
		Generation:  message.Generation(),
		History:     mappedChapters,
	}, nil
}

func (uc *CheckMessages) checkArchived(ctx context.Context, id string) (CheckMsgResult, error) {
	archivedMsg, err := uc.archivedMsgRepo.GetByID(ctx, uc.db, id)
	if err != nil {
		return CheckMsgResult{}, fmt.Errorf("archivedMsgRepo.GetByID: %w", err)
	}

	chapters := archivedMsg.History()
	mappedChapters := make([]CheckMsgChapter, 0, len(chapters))
	for _, chapter := range chapters {
		mappedChapters = append(mappedChapters, CheckMsgChapter{
			Generation:   chapter.Generation(),
			Queue:        chapter.Queue(),
			RedirectedAt: chapter.RedirectedAt(),
			Priority:     chapter.Priority(),
			Retries:      chapter.Retries(),
		})
	}

	return CheckMsgResult{
		ID:          archivedMsg.ID().String(),
		Queue:       archivedMsg.Queue(),
		Payload:     archivedMsg.Payload(),
		CreatedAt:   archivedMsg.CreatedAt(),
		FinalizedAt: utils.P(archivedMsg.FinalizedAt()),
		Status:      string(archivedMsg.Status()),
		Priority:    archivedMsg.Priority(),
		Retries:     archivedMsg.Retries(),
		Generation:  archivedMsg.Generation(),
		History:     mappedChapters,
	}, nil
}
