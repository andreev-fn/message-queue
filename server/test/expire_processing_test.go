package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/test/e2eutils"
)

func TestExpireProcessingAfterTimeout(t *testing.T) {
	app, clock := e2eutils.Prepare(t)

	const (
		taskKind     = "test"
		taskPayload  = `{"arg": 123}`
		taskPriority = 100
	)

	// Arrange
	taskID := e2eutils.CreateProcessingTask(t, app, taskKind, taskPayload, taskPriority)
	clock.Set(clock.Now().Add(6 * time.Minute))

	// Act
	affected, err := app.ExpireProcessing.Do(context.Background(), 10)
	require.NoError(t, err)

	// Assert
	require.Equal(t, 1, affected)

	updatedTask, err := app.TaskRepo.GetTaskByID(context.Background(), app.DB, taskID)
	require.NoError(t, err)

	require.Equal(t, domain.TaskStatusDelayed, updatedTask.Status())
	require.Equal(t, 1, updatedTask.Retries())
}

func TestExpireProcessingBeforeTimeout(t *testing.T) {
	app, clock := e2eutils.Prepare(t)

	const (
		taskKind     = "test"
		taskPayload  = `{"arg": 123}`
		taskPriority = 100
	)

	// Arrange
	taskID := e2eutils.CreateProcessingTask(t, app, taskKind, taskPayload, taskPriority)
	clock.Set(clock.Now().Add(3 * time.Minute))

	// Act
	affected, err := app.ExpireProcessing.Do(context.Background(), 10)
	require.NoError(t, err)

	// Assert
	require.Equal(t, 0, affected)

	unchangedTask, err := app.TaskRepo.GetTaskByID(context.Background(), app.DB, taskID)
	require.NoError(t, err)

	require.Equal(t, domain.TaskStatusProcessing, unchangedTask.Status())
	require.Equal(t, 0, unchangedTask.Retries())
}
