package usecases

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"server/internal/storage"
	"server/internal/utils/dbutils"
	"server/internal/utils/timeutils"
)

type AckMessages struct {
	clock   timeutils.Clock
	logger  *slog.Logger
	db      *sql.DB
	msgRepo *storage.MessageRepository
}

func NewAckMessages(
	clock timeutils.Clock,
	logger *slog.Logger,
	db *sql.DB,
	msgRepo *storage.MessageRepository,
) *AckMessages {
	return &AckMessages{
		clock:   clock,
		logger:  logger,
		db:      db,
		msgRepo: msgRepo,
	}
}

func (uc *AckMessages) Do(ctx context.Context, ids []string) error {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer dbutils.RollbackWithLog(tx, uc.logger)

	for _, id := range ids {
		message, err := uc.msgRepo.GetByID(ctx, uc.db, id)
		if err != nil {
			return fmt.Errorf("msgRepo.GetByID: %w", err)
		}

		if err := message.Complete(uc.clock); err != nil {
			return fmt.Errorf("message.Complete: %w", err)
		}

		if err := uc.msgRepo.Save(ctx, tx, message); err != nil {
			return fmt.Errorf("msgRepo.Save: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}
