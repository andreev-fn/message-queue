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
	clock    timeutils.Clock
	logger   *slog.Logger
	db       *sql.DB
	taskRepo *storage.TaskRepository
}

func NewExpireProcessing(
	clock timeutils.Clock,
	logger *slog.Logger,
	db *sql.DB,
	taskRepo *storage.TaskRepository,
) *ExpireProcessing {
	return &ExpireProcessing{
		clock:    clock,
		logger:   logger,
		db:       db,
		taskRepo: taskRepo,
	}
}

func (uc *ExpireProcessing) Do(ctx context.Context, limit int) (int, error) {
	tasks, err := uc.taskRepo.GetProcessingToExpire(ctx, uc.db, limit)
	if err != nil {
		return 0, fmt.Errorf("taskRepo.GetProcessingToExpire: %w", err)
	}

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer dbutils.RollbackWithLog(tx, uc.logger)

	for _, task := range tasks {
		// todo: replace with factory
		errorHandler := domain.NewExponentialErrorHandler(uc.clock)

		if err := errorHandler.HandleError(task, "timeout"); err != nil {
			return 0, err
		}

		if err := uc.taskRepo.Save(ctx, tx, task); err != nil {
			return 0, fmt.Errorf("taskRepo.Save: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("tx.Commit: %w", err)
	}

	return len(tasks), nil
}
