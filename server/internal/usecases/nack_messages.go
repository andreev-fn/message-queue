package usecases

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"server/internal/domain"
	"server/internal/storage"
	"server/internal/utils/dbutils"
	"server/internal/utils/timeutils"
)

type NackParams struct {
	ID        string
	Redeliver bool
}

type NackMessages struct {
	clock        timeutils.Clock
	logger       *slog.Logger
	db           *sql.DB
	msgRepo      *storage.MessageRepository
	redeliverSvc *domain.RedeliveryService
	maxBatchSize int
}

func NewNackMessages(
	clock timeutils.Clock,
	logger *slog.Logger,
	db *sql.DB,
	msgRepo *storage.MessageRepository,
	redeliverSvc *domain.RedeliveryService,
	maxBatchSize int,
) *NackMessages {
	return &NackMessages{
		clock:        clock,
		logger:       logger,
		db:           db,
		msgRepo:      msgRepo,
		redeliverSvc: redeliverSvc,
		maxBatchSize: maxBatchSize,
	}
}

func (uc *NackMessages) Do(ctx context.Context, nacks []NackParams) error {
	if len(nacks) > uc.maxBatchSize {
		return errors.New("batch size limit exceeded")
	}

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer dbutils.RollbackWithLog(tx, uc.logger)

	for _, nack := range nacks {
		message, err := uc.msgRepo.GetByID(ctx, uc.db, nack.ID)
		if err != nil {
			return fmt.Errorf("msgRepo.GetByID: %w", err)
		}

		if nack.Redeliver {
			if err := uc.redeliverSvc.HandleNack(message); err != nil {
				return err
			}
		} else {
			if err := message.MarkUndeliverable(uc.clock); err != nil {
				return fmt.Errorf("message.MarkUndeliverable: %w", err)
			}
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
