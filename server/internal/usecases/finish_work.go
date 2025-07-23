package usecases

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"server/internal/domain"
	"server/internal/storage"
	"server/internal/utils/timeutils"
)

type FinishWork struct {
	clock    timeutils.Clock
	logger   *slog.Logger
	db       *sql.DB
	taskRepo *storage.TaskRepository
}

func NewFinishWork(
	clock timeutils.Clock,
	logger *slog.Logger,
	db *sql.DB,
	taskRepo *storage.TaskRepository,
) *FinishWork {
	return &FinishWork{
		clock:    clock,
		logger:   logger,
		db:       db,
		taskRepo: taskRepo,
	}
}

func (uc *FinishWork) Do(ctx context.Context, id string, taskResult json.RawMessage, errorCode *string) error {
	if taskResult == nil && errorCode == nil {
		return errors.New("exactly one of result or error expected")
	}

	task, err := uc.taskRepo.GetTaskByID(ctx, uc.db, id)
	if err != nil {
		return fmt.Errorf("taskRepo.GetTaskByID: %w", err)
	}

	if errorCode != nil {
		// todo: replace with factory
		errorHandler := domain.NewExponentialErrorHandler(uc.clock)

		if err := errorHandler.HandleError(task, *errorCode); err != nil {
			return err
		}
	} else {
		if err := task.Complete(uc.clock, taskResult); err != nil {
			return fmt.Errorf("task.Complete: %w", err)
		}
	}

	if err := uc.taskRepo.SaveInNewTransaction(ctx, uc.db, task); err != nil {
		return fmt.Errorf("taskRepo.SaveInNewTransaction: %w", err)
	}

	return nil
}
