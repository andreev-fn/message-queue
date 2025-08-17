package usecases

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"server/internal/domain"
	"server/internal/storage"
	"server/internal/utils/dbutils"
	"server/internal/utils/timeutils"
)

type ExpireProcessing struct {
	clock   timeutils.Clock
	logger  *slog.Logger
	db      *sql.DB
	msgRepo *storage.MessageRepository
}

func NewExpireProcessing(
	clock timeutils.Clock,
	logger *slog.Logger,
	db *sql.DB,
	msgRepo *storage.MessageRepository,
) *ExpireProcessing {
	return &ExpireProcessing{
		clock:   clock,
		logger:  logger,
		db:      db,
		msgRepo: msgRepo,
	}
}

func (uc *ExpireProcessing) Do(ctx context.Context, limit int) (int, error) {
	messages, err := uc.msgRepo.GetProcessingToExpire(ctx, uc.db, limit)
	if err != nil {
		return 0, fmt.Errorf("msgRepo.GetProcessingToExpire: %w", err)
	}

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer dbutils.RollbackWithLog(tx, uc.logger)

	for _, message := range messages {
		// todo: replace with factory
		errorHandler := domain.NewExponentialErrorHandler(uc.clock)

		if err := errorHandler.HandleError(message, "timeout"); err != nil {
			return 0, err
		}

		if err := uc.msgRepo.Save(ctx, tx, message); err != nil {
			return 0, fmt.Errorf("msgRepo.Save: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("tx.Commit: %w", err)
	}

	return len(messages), nil
}
