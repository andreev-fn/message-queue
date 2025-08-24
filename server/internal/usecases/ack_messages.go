package usecases

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"server/internal/appbuilder/requestscope"
	"server/internal/storage"
	"server/internal/utils/dbutils"
	"server/internal/utils/timeutils"
)

type AckParams struct {
	ID      string
	Release []string
}

type AckMessages struct {
	clock        timeutils.Clock
	logger       *slog.Logger
	db           *sql.DB
	msgRepo      *storage.MessageRepository
	scopeFactory requestscope.Factory
	maxBatchSize int
}

func NewAckMessages(
	clock timeutils.Clock,
	logger *slog.Logger,
	db *sql.DB,
	msgRepo *storage.MessageRepository,
	scopeFactory requestscope.Factory,
	maxBatchSize int,
) *AckMessages {
	return &AckMessages{
		clock:        clock,
		logger:       logger,
		db:           db,
		msgRepo:      msgRepo,
		scopeFactory: scopeFactory,
		maxBatchSize: maxBatchSize,
	}
}

func (uc *AckMessages) Do(ctx context.Context, acks []AckParams) error {
	batchSize := len(acks)
	for _, ack := range acks {
		batchSize += len(ack.Release)
	}
	if batchSize > uc.maxBatchSize {
		return errors.New("batch size limit exceeded")
	}

	scope := uc.scopeFactory.New()

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer dbutils.RollbackWithLog(tx, uc.logger)

	for _, ack := range acks {
		message, err := uc.msgRepo.GetByID(ctx, uc.db, ack.ID)
		if err != nil {
			return fmt.Errorf("msgRepo.GetByID: %w", err)
		}

		if err := message.Complete(uc.clock); err != nil {
			return fmt.Errorf("message.Complete: %w", err)
		}

		if err := uc.msgRepo.Save(ctx, tx, message); err != nil {
			return fmt.Errorf("msgRepo.Save: %w", err)
		}

		for _, releaseID := range ack.Release {
			message, err := uc.msgRepo.GetByID(ctx, uc.db, releaseID)
			if err != nil {
				return fmt.Errorf("msgRepo.GetByID: %w", err)
			}

			if err := message.Release(uc.clock, scope.Dispatcher); err != nil {
				return fmt.Errorf("message.Release: %w", err)
			}

			if err := uc.msgRepo.Save(ctx, tx, message); err != nil {
				return fmt.Errorf("msgRepo.Save: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	if err := scope.MsgReadyNotifier.Flush(); err != nil {
		uc.logger.Error("scope.MsgReadyNotifier.Flush", "error", err)
	}

	return nil
}
