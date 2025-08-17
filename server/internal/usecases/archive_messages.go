package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"server/internal/domain"
	"server/internal/storage"
	"server/internal/utils/timeutils"
)

type ArchiveMessages struct {
	clock           timeutils.Clock
	db              *sql.DB
	msgRepo         *storage.MessageRepository
	archivedMsgRepo *storage.ArchivedMsgRepository
}

func NewArchiveMessages(
	clock timeutils.Clock,
	db *sql.DB,
	msgRepo *storage.MessageRepository,
	archivedMsgRepo *storage.ArchivedMsgRepository,
) *ArchiveMessages {
	return &ArchiveMessages{
		clock:           clock,
		db:              db,
		msgRepo:         msgRepo,
		archivedMsgRepo: archivedMsgRepo,
	}
}

func (uc *ArchiveMessages) Do(ctx context.Context, limit int) (int, error) {
	messages, err := uc.msgRepo.GetFinalizedToArchive(ctx, uc.db, limit)
	if err != nil {
		return 0, fmt.Errorf("msgRepo.GetFinalizedToArchive: %w", err)
	}

	affected := 0

	for _, msg := range messages {
		archivedMsg, err := domain.NewArchivedMsg(msg)
		if err != nil {
			return affected, fmt.Errorf("domain.NewArchivedMsg: %w", err)
		}

		if err := uc.archivedMsgRepo.Upsert(ctx, uc.db, archivedMsg); err != nil {
			return affected, fmt.Errorf("archivedMsgRepo.Upsert: %w", err)
		}

		if err := uc.msgRepo.DeleteInNewTransaction(ctx, uc.db, msg); err != nil {
			return affected, fmt.Errorf("msgRepo.DeleteInNewTransaction: %w", err)
		}

		affected++
	}

	return affected, nil
}
