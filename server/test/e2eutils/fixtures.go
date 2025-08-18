package e2eutils

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"server/internal/appbuilder"
	"server/internal/utils"
)

func CreateMsg(t *testing.T, app *appbuilder.App, queue string, payload string, priority int) string {
	t.Helper()

	msgID, err := app.CreateMessage.Do(
		context.Background(),
		queue,
		json.RawMessage(payload),
		priority,
		false,
		nil,
	)
	require.NoError(t, err)

	return msgID
}

func CreateReadyMsg(t *testing.T, app *appbuilder.App, queue string, payload string, priority int) string {
	t.Helper()

	msgID := CreateMsg(t, app, queue, payload, priority)

	err := app.ConfirmMessage.Do(context.Background(), msgID)
	require.NoError(t, err)

	return msgID
}

func CreateProcessingMsg(t *testing.T, app *appbuilder.App, queue string, payload string, priority int) string {
	t.Helper()

	msgID := CreateReadyMsg(t, app, queue, payload, priority)

	result, err := app.TakeWork.Do(context.Background(), []string{queue}, 1, 0)
	require.NoError(t, err)

	require.Len(t, result, 1)
	require.Equal(t, msgID, result[0].ID)

	return msgID
}

func CreateDelayedMsg(t *testing.T, app *appbuilder.App, queue string, payload string) string {
	t.Helper()

	msgID := CreateProcessingMsg(t, app, queue, payload, 100)

	err := app.FinishWork.Do(context.Background(), msgID, nil, utils.P("timeout_error"))
	require.NoError(t, err)

	return msgID
}

func CreateCompletedMsg(t *testing.T, app *appbuilder.App, queue string, payload string, result string) string {
	t.Helper()

	msgID := CreateProcessingMsg(t, app, queue, payload, 100)

	err := app.FinishWork.Do(context.Background(), msgID, json.RawMessage(result), nil)
	require.NoError(t, err)

	return msgID
}
