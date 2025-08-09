package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/test/e2eutils"
)

func TestResumeDelayedAfterTimeout(t *testing.T) {
	app, clock := e2eutils.Prepare(t)

	const (
		taskKind    = "test"
		taskPayload = `{"arg": 123}`
	)

	// Arrange
	taskID := e2eutils.CreateDelayedTask(t, app, taskKind, taskPayload)
	clock.Set(clock.Now().Add(3 * time.Minute))

	// Act
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
	app, clock := e2eutils.Prepare(t)

	const (
		taskKind    = "test"
		taskPayload = `{"arg": 123}`
	)

	// Arrange
	taskID := e2eutils.CreateDelayedTask(t, app, taskKind, taskPayload)
	clock.Set(clock.Now().Add(1 * time.Minute))

	// Act
	affected, err := app.ResumeDelayed.Do(context.Background(), 10)
	require.NoError(t, err)

	// Assert
	require.Equal(t, 0, affected)

	unchangedTask, err := app.TaskRepo.GetTaskByID(context.Background(), app.DB, taskID)
	require.NoError(t, err)

	require.Equal(t, domain.TaskStatusDelayed, unchangedTask.Status())
	require.Equal(t, 1, unchangedTask.Retries())
}
