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

type ResumeDelayed struct {
	clock    timeutils.Clock
	logger   *slog.Logger
	db       *sql.DB
	taskRepo *storage.TaskRepository
}

func NewResumeDelayed(
	clock timeutils.Clock,
	logger *slog.Logger,
	db *sql.DB,
	taskRepo *storage.TaskRepository,
) *ResumeDelayed {
	return &ResumeDelayed{
		clock:    clock,
		logger:   logger,
		db:       db,
		taskRepo: taskRepo,
	}
}

func (uc *ResumeDelayed) Do(ctx context.Context, limit int) (int, error) {
	tasks, err := uc.taskRepo.GetDelayedReadyToResume(ctx, uc.db, limit)
	if err != nil {
		return 0, fmt.Errorf("taskRepo.GetDelayedReadyToResume: %w", err)
	}

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer dbutils.RollbackWithLog(tx, uc.logger)

	for _, task := range tasks {
		if err := task.Resume(uc.clock); err != nil {
			return 0, fmt.Errorf("task.Resume: %w", err)
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
