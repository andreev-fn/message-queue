package e2eutils

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"server/internal/appbuilder"
	"server/internal/utils"
)

func CreateTask(t *testing.T, app *appbuilder.App, kind string, payload string, priority int) string {
	t.Helper()

	taskID, err := app.CreateTask.Do(
		context.Background(),
		kind,
		json.RawMessage(payload),
		priority,
		false,
		nil,
	)
	require.NoError(t, err)

	return taskID
}

func CreateReadyTask(t *testing.T, app *appbuilder.App, kind string, payload string, priority int) string {
	t.Helper()

	taskID := CreateTask(t, app, kind, payload, priority)

	err := app.ConfirmTask.Do(context.Background(), taskID)
	require.NoError(t, err)

	return taskID
}

func CreateProcessingTask(t *testing.T, app *appbuilder.App, kind string, payload string, priority int) string {
	t.Helper()

	taskID := CreateReadyTask(t, app, kind, payload, priority)

	result, err := app.TakeWork.Do(context.Background(), []string{kind}, 1, 0)
	require.NoError(t, err)

	require.Len(t, result, 1)
	require.Equal(t, taskID, result[0].ID)

	return taskID
}

func CreateDelayedTask(t *testing.T, app *appbuilder.App, kind string, payload string) string {
	t.Helper()

	taskID := CreateProcessingTask(t, app, kind, payload, 100)

	err := app.FinishWork.Do(context.Background(), taskID, nil, utils.P("timeout_error"))
	require.NoError(t, err)

	return taskID
}

func CreateCompletedTask(t *testing.T, app *appbuilder.App, kind string, payload string, result string) string {
	t.Helper()

	taskID := CreateProcessingTask(t, app, kind, payload, 100)

	err := app.FinishWork.Do(context.Background(), taskID, json.RawMessage(result), nil)
	require.NoError(t, err)

	return taskID
}
