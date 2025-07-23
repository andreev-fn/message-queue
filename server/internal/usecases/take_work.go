package usecases

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"server/internal/storage"
	"server/internal/utils/dbutils"
	"server/internal/utils/timeutils"
)

type TaskToWork struct {
	ID      string
	Payload json.RawMessage
}

type TakeWork struct {
	logger   *slog.Logger
	clock    timeutils.Clock
	db       *sql.DB
	taskRepo *storage.TaskRepository
}

func NewTakeWork(
	logger *slog.Logger,
	clock timeutils.Clock,
	db *sql.DB,
	taskRepo *storage.TaskRepository,
) *TakeWork {
	return &TakeWork{
		logger:   logger,
		clock:    clock,
		db:       db,
		taskRepo: taskRepo,
	}
}

func (uc *TakeWork) Do(ctx context.Context, kinds []string, limit int) ([]TaskToWork, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer dbutils.RollbackWithLog(tx, uc.logger)

	tasks, err := uc.taskRepo.GetReadyWithLock(ctx, tx, kinds, limit)
	if err != nil {
		return nil, fmt.Errorf("taskRepo.GetReadyWithLock: %w", err)
	}

	for _, task := range tasks {
		if err := task.StartProcessing(uc.clock); err != nil {
			return nil, fmt.Errorf("task.StartProcessing: %w", err)
		}

		if err := uc.taskRepo.Save(ctx, tx, task); err != nil {
			return nil, fmt.Errorf("taskRepo.Save: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("tx.Commit: %w", err)
	}

	var result []TaskToWork

	for _, task := range tasks {
		result = append(result, TaskToWork{
			ID:      task.ID().String(),
			Payload: task.Payload(),
		})
	}

	return result, nil
}
