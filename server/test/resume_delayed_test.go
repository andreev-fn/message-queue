package test

import (
	"context"
	"github.com/stretchr/testify/require"
	"server/internal/domain"
	"testing"
	"time"
)

func TestResumeDelayedAfterTimeout(t *testing.T) {
	app, clock := buildTestApp(t)
	cleanupDatabase(t, app.DB)

	// Create test data directly in the database
	const taskID = "00000000-0000-0000-0000-000000000001"
	testTask := createTask(t, app, taskID, "test", 100)

	err := testTask.Confirm(app.Clock)
	require.NoError(t, err)

	err = testTask.StartProcessing(app.Clock)
	require.NoError(t, err)

	err = testTask.Delay(app.Clock, app.Clock.Now().Add(5*time.Minute))
	require.NoError(t, err)

	err = app.TaskRepo.SaveInNewTransaction(context.Background(), app.DB, testTask)
	require.NoError(t, err)

	clock.Set(clock.Now().Add(6 * time.Minute))

	affected, err := app.ResumeDelayed.Do(context.Background(), 10)
	require.NoError(t, err)

	// Assert
	require.Equal(t, 1, affected)

	updatedTask, err := app.TaskRepo.GetTaskByID(context.Background(), app.DB, taskID)
	require.NoError(t, err)

	require.Equal(t, domain.TaskStatusReady, updatedTask.Status())
	require.Equal(t, 1, updatedTask.Retries())
}

func TestResumeDelayedBeforeTimeout(t *testing.T) {
	app, clock := buildTestApp(t)
	cleanupDatabase(t, app.DB)

	// Create test data directly in the database
	const taskID = "00000000-0000-0000-0000-000000000001"
	testTask := createTask(t, app, taskID, "test", 100)

	err := testTask.Confirm(app.Clock)
	require.NoError(t, err)

	err = testTask.StartProcessing(app.Clock)
	require.NoError(t, err)

	err = testTask.Delay(app.Clock, app.Clock.Now().Add(5*time.Minute))
	require.NoError(t, err)

	err = app.TaskRepo.SaveInNewTransaction(context.Background(), app.DB, testTask)
	require.NoError(t, err)

	clock.Set(clock.Now().Add(4 * time.Minute))

	affected, err := app.ResumeDelayed.Do(context.Background(), 10)
	require.NoError(t, err)

	// Assert
	require.Equal(t, 0, affected)

	updatedTask, err := app.TaskRepo.GetTaskByID(context.Background(), app.DB, taskID)
	require.NoError(t, err)

	require.Equal(t, domain.TaskStatusDelayed, updatedTask.Status())
	require.Equal(t, 1, updatedTask.Retries())
}
