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

type ReleaseMessages struct {
	logger       *slog.Logger
	clock        timeutils.Clock
	db           *sql.DB
	msgRepo      *storage.MessageRepository
	scopeFactory requestscope.Factory
	maxBatchSize int
}

func NewReleaseMessages(
	logger *slog.Logger,
	clock timeutils.Clock,
	db *sql.DB,
	msgRepo *storage.MessageRepository,
	scopeFactory requestscope.Factory,
	maxBatchSize int,
) *ReleaseMessages {
	return &ReleaseMessages{
		logger:       logger,
		clock:        clock,
		db:           db,
		msgRepo:      msgRepo,
		scopeFactory: scopeFactory,
		maxBatchSize: maxBatchSize,
	}
}

func (uc *ReleaseMessages) Do(ctx context.Context, ids []string) error {
	if len(ids) > uc.maxBatchSize {
		return errors.New("batch size limit exceeded")
	}

	scope := uc.scopeFactory.New()

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

		if err := message.Release(uc.clock, scope.Dispatcher); err != nil {
			return fmt.Errorf("message.Release: %w", err)
		}

		if err := uc.msgRepo.Save(ctx, tx, message); err != nil {
			return fmt.Errorf("msgRepo.Save: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	if err := scope.MsgAvailabilityNotifier.Flush(); err != nil {
		uc.logger.Error("scope.MsgAvailabilityNotifier.Flush", "error", err)
	}

	return nil
}
