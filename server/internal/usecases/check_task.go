package usecases

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"server/internal/storage"
	"server/internal/utils"
	"time"
)

type CheckTaskResult struct {
	ID          string
	Kind        string
	CreatedAt   time.Time
	FinalizedAt *time.Time
	Status      string
	Retries     int
	Payload     json.RawMessage
	Result      *json.RawMessage
}

type CheckTask struct {
	db               *sql.DB
	taskRepo         *storage.TaskRepository
	archivedTaskRepo *storage.ArchivedTaskRepository
}

func NewCheckTask(
	db *sql.DB,
	taskRepo *storage.TaskRepository,
	archivedTaskRepo *storage.ArchivedTaskRepository,
) *CheckTask {
	return &CheckTask{
		db:               db,
		taskRepo:         taskRepo,
		archivedTaskRepo: archivedTaskRepo,
	}
}

func (uc *CheckTask) Do(ctx context.Context, id string) (*CheckTaskResult, error) {
	task, err := uc.taskRepo.GetTaskByID(ctx, uc.db, id)
	if err != nil {
		if errors.Is(err, storage.ErrTaskNotFound) {
			return uc.checkArchived(ctx, id)
		}
		return nil, fmt.Errorf("taskRepo.GetTaskByID: %w", err)
	}

	return &CheckTaskResult{
		ID:          task.ID().String(),
		Kind:        task.Kind(),
		Payload:     task.Payload(),
		CreatedAt:   task.CreatedAt(),
		FinalizedAt: task.FinalizedAt(),
		Status:      string(task.Status()),
		Retries:     task.Retries(),
		Result:      task.Result(),
	}, nil
}

func (uc *CheckTask) checkArchived(ctx context.Context, id string) (*CheckTaskResult, error) {
	archivedTask, err := uc.archivedTaskRepo.GetTaskByID(ctx, uc.db, id)
	if err != nil {
		return nil, fmt.Errorf("taskRepo.GetTaskByID: %w", err)
	}

	return &CheckTaskResult{
		ID:          archivedTask.ID().String(),
		Kind:        archivedTask.Kind(),
		Payload:     archivedTask.Payload(),
		CreatedAt:   archivedTask.CreatedAt(),
		FinalizedAt: utils.P(archivedTask.FinalizedAt()),
		Status:      string(archivedTask.Status()),
		Retries:     archivedTask.Retries(),
		Result:      archivedTask.Result(),
	}, nil
}
