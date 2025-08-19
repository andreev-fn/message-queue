package usecases

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"server/internal/domain"
	"server/internal/storage"
	"server/internal/utils/timeutils"
)

type FinishWork struct {
	clock   timeutils.Clock
	logger  *slog.Logger
	db      *sql.DB
	msgRepo *storage.MessageRepository
}

func NewFinishWork(
	clock timeutils.Clock,
	logger *slog.Logger,
	db *sql.DB,
	msgRepo *storage.MessageRepository,
) *FinishWork {
	return &FinishWork{
		clock:   clock,
		logger:  logger,
		db:      db,
		msgRepo: msgRepo,
	}
}

func (uc *FinishWork) Do(ctx context.Context, id string, errorCode *string) error {
	message, err := uc.msgRepo.GetByID(ctx, uc.db, id)
	if err != nil {
		return fmt.Errorf("msgRepo.GetByID: %w", err)
	}

	if errorCode != nil {
		// todo: replace with factory
		errorHandler := domain.NewExponentialErrorHandler(uc.clock)

		if err := errorHandler.HandleError(message, *errorCode); err != nil {
			return err
		}
	} else {
		if err := message.Complete(uc.clock); err != nil {
			return fmt.Errorf("message.Complete: %w", err)
		}
	}

	if err := uc.msgRepo.SaveInNewTransaction(ctx, uc.db, message); err != nil {
		return fmt.Errorf("msgRepo.SaveInNewTransaction: %w", err)
	}

	return nil
}
