package test

import (
	"context"
	"encoding/json"
	"server/internal/domain"
	"server/internal/storage"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestArchiveTasksFinalized(t *testing.T) {
	app, clock := buildTestApp(t)
	cleanupDatabase(t, app.DB)

	// Create test data directly in the database
	const taskID = "00000000-0000-0000-0000-000000000001"
	testTask := createTask(t, app, taskID, "test", 100)

	err := testTask.Confirm(app.Clock, NoopEventDispatcher{})
	require.NoError(t, err)

	err = testTask.StartProcessing(app.Clock)
	require.NoError(t, err)

	err = testTask.Complete(app.Clock, json.RawMessage(`{"result":"success"}`))
	require.NoError(t, err)

	err = app.TaskRepo.SaveInNewTransaction(context.Background(), app.DB, testTask)
	require.NoError(t, err)

	clock.Set(clock.Now().Add(time.Minute))

	affected, err := app.ArchiveTasks.Do(context.Background(), 10)
	require.NoError(t, err)

	require.Equal(t, 1, affected)

	_, err = app.TaskRepo.GetTaskByID(context.Background(), app.DB, taskID)
	require.ErrorIs(t, err, storage.ErrTaskNotFound)

	archivedTask, err := app.ArchivedTaskRepo.GetTaskByID(context.Background(), app.DB, taskID)
	require.NoError(t, err)

	require.Equal(t, domain.TaskStatusCompleted, archivedTask.Status())
}

func TestArchiveTasksNotFinal(t *testing.T) {
	app, clock := buildTestApp(t)
	cleanupDatabase(t, app.DB)

	// Create test data directly in the database
	const taskID = "00000000-0000-0000-0000-000000000001"
	testTask := createTask(t, app, taskID, "test", 100)

	err := testTask.Confirm(app.Clock, NoopEventDispatcher{})
	require.NoError(t, err)

	err = testTask.StartProcessing(app.Clock)
	require.NoError(t, err)

	err = app.TaskRepo.SaveInNewTransaction(context.Background(), app.DB, testTask)
	require.NoError(t, err)

	clock.Set(clock.Now().Add(time.Minute))

	affected, err := app.ArchiveTasks.Do(context.Background(), 10)
	require.NoError(t, err)

	require.Equal(t, 0, affected)

	_, err = app.ArchivedTaskRepo.GetTaskByID(context.Background(), app.DB, taskID)
	require.ErrorIs(t, err, storage.ErrArchivedTaskNotFound)

	unchangedTask, err := app.TaskRepo.GetTaskByID(context.Background(), app.DB, taskID)
	require.NoError(t, err)

	require.Equal(t, domain.TaskStatusProcessing, unchangedTask.Status())
}
