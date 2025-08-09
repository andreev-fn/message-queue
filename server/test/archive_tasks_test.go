package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/internal/storage"
	"server/test/e2eutils"
)

func TestArchiveTasksFinalized(t *testing.T) {
	app, clock := e2eutils.Prepare(t)

	const (
		taskKind    = "test"
		taskPayload = `{"arg": 123}`
		taskResult  = `{"result":"success"}`
	)

	// Arrange
	taskID := e2eutils.CreateCompletedTask(t, app, taskKind, taskPayload, taskResult)
	clock.Set(clock.Now().Add(time.Minute))

	// Act
	affected, err := app.ArchiveTasks.Do(context.Background(), 10)
	require.NoError(t, err)

	// Assert
	require.Equal(t, 1, affected)

	_, err = app.TaskRepo.GetTaskByID(context.Background(), app.DB, taskID)
	require.ErrorIs(t, err, storage.ErrTaskNotFound)

	archivedTask, err := app.ArchivedTaskRepo.GetTaskByID(context.Background(), app.DB, taskID)
	require.NoError(t, err)

	require.Equal(t, domain.TaskStatusCompleted, archivedTask.Status())
	require.Equal(t, taskKind, archivedTask.Kind())
	require.JSONEq(t, taskPayload, string(archivedTask.Payload()))
	require.NotNil(t, archivedTask.Result())
	require.JSONEq(t, taskResult, string(*archivedTask.Result()))
}

func TestArchiveTasksNotFinal(t *testing.T) {
	app, clock := e2eutils.Prepare(t)

	const (
		taskKind     = "test"
		taskPayload  = `{"arg": 123}`
		taskPriority = 100
	)

	// Arrange
	taskID := e2eutils.CreateProcessingTask(t, app, taskKind, taskPayload, taskPriority)
	clock.Set(clock.Now().Add(time.Minute))

	// Act
	affected, err := app.ArchiveTasks.Do(context.Background(), 10)
	require.NoError(t, err)

	// Assert
	require.Equal(t, 0, affected)

	_, err = app.ArchivedTaskRepo.GetTaskByID(context.Background(), app.DB, taskID)
	require.ErrorIs(t, err, storage.ErrArchivedTaskNotFound)

	unchangedTask, err := app.TaskRepo.GetTaskByID(context.Background(), app.DB, taskID)
	require.NoError(t, err)

	require.Equal(t, domain.TaskStatusProcessing, unchangedTask.Status())
}
