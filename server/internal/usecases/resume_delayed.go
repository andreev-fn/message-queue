package usecases

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"server/internal/appbuilder/requestscope"
	"server/internal/storage"
	"server/internal/utils/dbutils"
	"server/internal/utils/timeutils"
)

type ResumeDelayed struct {
	clock        timeutils.Clock
	logger       *slog.Logger
	db           *sql.DB
	msgRepo      *storage.MessageRepository
	scopeFactory requestscope.Factory
}

func NewResumeDelayed(
	clock timeutils.Clock,
	logger *slog.Logger,
	db *sql.DB,
	msgRepo *storage.MessageRepository,
	scopeFactory requestscope.Factory,
) *ResumeDelayed {
	return &ResumeDelayed{
		clock:        clock,
		logger:       logger,
		db:           db,
		msgRepo:      msgRepo,
		scopeFactory: scopeFactory,
	}
}

func (uc *ResumeDelayed) Do(ctx context.Context, limit int) (int, error) {
	scope := uc.scopeFactory.New()

	messages, err := uc.msgRepo.GetDelayedReadyToResume(ctx, uc.db, limit)
	if err != nil {
		return 0, fmt.Errorf("msgRepo.GetDelayedReadyToResume: %w", err)
	}

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer dbutils.RollbackWithLog(tx, uc.logger)

	for _, message := range messages {
		if err := message.Resume(uc.clock, scope.Dispatcher); err != nil {
			return 0, fmt.Errorf("message.Resume: %w", err)
		}

		if err := uc.msgRepo.Save(ctx, tx, message); err != nil {
			return 0, fmt.Errorf("msgRepo.Save: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("tx.Commit: %w", err)
	}

	if err := scope.MsgAvailabilityNotifier.Flush(); err != nil {
		uc.logger.Error("scope.MsgAvailabilityNotifier.Flush", "error", err)
	}

	return len(messages), nil
}
