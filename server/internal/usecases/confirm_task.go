package usecases

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"server/internal/storage"
	"server/internal/utils/timeutils"
)

type ConfirmTask struct {
	logger   *slog.Logger
	clock    timeutils.Clock
	db       *sql.DB
	taskRepo *storage.TaskRepository
}

func NewConfirmTask(
	logger *slog.Logger,
	clock timeutils.Clock,
	db *sql.DB,
	taskRepo *storage.TaskRepository,
) *ConfirmTask {
	return &ConfirmTask{
		logger:   logger,
		clock:    clock,
		db:       db,
		taskRepo: taskRepo,
	}
}

func (uc *ConfirmTask) Do(ctx context.Context, id string) error {
	task, err := uc.taskRepo.GetTaskByID(ctx, uc.db, id)
	if err != nil {
		return fmt.Errorf("taskRepo.GetTaskByID: %w", err)
	}

	if err := task.Confirm(uc.clock); err != nil {
		return fmt.Errorf("task.Confirm: %w", err)
	}

	if err := uc.taskRepo.SaveInNewTransaction(ctx, uc.db, task); err != nil {
		return fmt.Errorf("taskRepo.SaveInNewTransaction: %w", err)
	}

	return nil
}
