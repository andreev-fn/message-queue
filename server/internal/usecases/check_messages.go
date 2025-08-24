package usecases

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"server/internal/storage"
	"server/internal/utils"
)

type CheckMsgResult struct {
	ID          string
	Queue       string
	CreatedAt   time.Time
	FinalizedAt *time.Time
	Status      string
	Retries     int
	Payload     string
}

type CheckMessages struct {
	db              *sql.DB
	msgRepo         *storage.MessageRepository
	archivedMsgRepo *storage.ArchivedMsgRepository
	maxBatchSize    int
}

func NewCheckMessages(
	db *sql.DB,
	msgRepo *storage.MessageRepository,
	archivedMsgRepo *storage.ArchivedMsgRepository,
	maxBatchSize int,
) *CheckMessages {
	return &CheckMessages{
		db:              db,
		msgRepo:         msgRepo,
		archivedMsgRepo: archivedMsgRepo,
		maxBatchSize:    maxBatchSize,
	}
}

func (uc *CheckMessages) Do(ctx context.Context, ids []string) ([]CheckMsgResult, error) {
	if len(ids) > uc.maxBatchSize {
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
	message, err := uc.msgRepo.GetByID(ctx, uc.db, id)
	if err != nil {
		if errors.Is(err, storage.ErrMsgNotFound) {
			return uc.checkArchived(ctx, id)
		}
		return CheckMsgResult{}, fmt.Errorf("msgRepo.GetByID: %w", err)
	}

	return CheckMsgResult{
		ID:          message.ID().String(),
		Queue:       message.Queue(),
		Payload:     message.Payload(),
		CreatedAt:   message.CreatedAt(),
		FinalizedAt: message.FinalizedAt(),
		Status:      string(message.Status()),
		Retries:     message.Retries(),
	}, nil
}

func (uc *CheckMessages) checkArchived(ctx context.Context, id string) (CheckMsgResult, error) {
	archivedMsg, err := uc.archivedMsgRepo.GetByID(ctx, uc.db, id)
	if err != nil {
		return CheckMsgResult{}, fmt.Errorf("archivedMsgRepo.GetByID: %w", err)
	}

	return CheckMsgResult{
		ID:          archivedMsg.ID().String(),
		Queue:       archivedMsg.Queue(),
		Payload:     archivedMsg.Payload(),
		CreatedAt:   archivedMsg.CreatedAt(),
		FinalizedAt: utils.P(archivedMsg.FinalizedAt()),
		Status:      string(archivedMsg.Status()),
		Retries:     archivedMsg.Retries(),
	}, nil
}
