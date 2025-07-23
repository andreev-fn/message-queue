package usecases

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log/slog"
	"server/internal/domain"
	"server/internal/storage"
	"server/internal/utils/timeutils"
	"time"
)

type CreateTask struct {
	logger   *slog.Logger
	clock    timeutils.Clock
	db       *sql.DB
	taskRepo *storage.TaskRepository
}

func NewCreateTask(
	logger *slog.Logger,
	clock timeutils.Clock,
	db *sql.DB,
	taskRepo *storage.TaskRepository,
) *CreateTask {
	return &CreateTask{
		logger:   logger,
		clock:    clock,
		db:       db,
		taskRepo: taskRepo,
	}
}

func (uc *CreateTask) Do(
	ctx context.Context,
	kind string,
	payload json.RawMessage,
	priority int,
	autoConfirm bool,
	startAt *time.Time,
) (string, error) {
	task, err := domain.NewTask(uc.clock, uuid.New(), kind, payload, priority, startAt)
	if err != nil {
		return "", err
	}

	if autoConfirm {
		if err := task.Confirm(uc.clock); err != nil {
			return "", fmt.Errorf("task.Confirm: %w", err)
		}
	}

	if err := uc.taskRepo.SaveInNewTransaction(ctx, uc.db, task); err != nil {
		return "", fmt.Errorf("taskRepo.SaveInNewTransaction: %w", err)
	}

	return task.ID().String(), nil
}
