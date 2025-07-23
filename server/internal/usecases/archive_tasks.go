package usecases

import (
	"context"
	"database/sql"
	"fmt"
	"server/internal/domain"
	"server/internal/storage"
	"server/internal/utils/timeutils"
)

type ArchiveTasks struct {
	clock            timeutils.Clock
	db               *sql.DB
	taskRepo         *storage.TaskRepository
	archivedTaskRepo *storage.ArchivedTaskRepository
}

func NewArchiveTasks(
	clock timeutils.Clock,
	db *sql.DB,
	taskRepo *storage.TaskRepository,
	archivedTaskRepo *storage.ArchivedTaskRepository,
) *ArchiveTasks {
	return &ArchiveTasks{
		clock:            clock,
		db:               db,
		taskRepo:         taskRepo,
		archivedTaskRepo: archivedTaskRepo,
	}
}

func (uc *ArchiveTasks) Do(ctx context.Context, limit int) (int, error) {
	tasks, err := uc.taskRepo.GetTasksToArchive(ctx, uc.db, limit)
	if err != nil {
		return 0, fmt.Errorf("taskRepo.GetTasksToArchive: %w", err)
	}

	affected := 0

	for _, task := range tasks {
		archivedTask, err := domain.NewArchivedTask(task)
		if err != nil {
			return affected, fmt.Errorf("domain.NewArchivedTask: %w", err)
		}

		if err := uc.archivedTaskRepo.Upsert(ctx, uc.db, archivedTask); err != nil {
			return affected, fmt.Errorf("archivedTaskRepo.Upsert: %w", err)
		}

		if err := uc.taskRepo.DeleteInNewTransaction(ctx, uc.db, task); err != nil {
			return affected, fmt.Errorf("taskRepo.DeleteInNewTransaction: %w", err)
		}

		affected++
	}

	return affected, nil
}
