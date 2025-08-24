package e2eutils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"server/internal/appbuilder"
	"server/internal/usecases"
)

func CreateMsg(t *testing.T, app *appbuilder.App, queue string, payload string, priority int) string {
	t.Helper()

	msgIDs, err := app.PublishMessages.Do(
		context.Background(),
		[]usecases.NewMessageParams{{
			Queue:    queue,
			Payload:  payload,
			Priority: priority,
			StartAt:  nil,
		}},
		false,
	)
	require.NoError(t, err)
	require.Len(t, msgIDs, 1)

	return msgIDs[0]
}

func CreateAvailableMsg(t *testing.T, app *appbuilder.App, queue string, payload string, priority int) string {
	t.Helper()

	msgID := CreateMsg(t, app, queue, payload, priority)

	err := app.ReleaseMessages.Do(context.Background(), []string{msgID})
	require.NoError(t, err)

	return msgID
}

func CreateProcessingMsg(t *testing.T, app *appbuilder.App, queue string, payload string, priority int) string {
	t.Helper()

	msgID := CreateAvailableMsg(t, app, queue, payload, priority)

	result, err := app.ConsumeMessages.Do(context.Background(), queue, 1, 0)
	require.NoError(t, err)

	require.Len(t, result, 1)
	require.Equal(t, msgID, result[0].ID)

	return msgID
}

func CreateDelayedMsg(t *testing.T, app *appbuilder.App, queue string, payload string) string {
	t.Helper()

	msgID := CreateProcessingMsg(t, app, queue, payload, 100)

	err := app.NackMessages.Do(context.Background(), []usecases.NackParams{{ID: msgID, Redeliver: true}})
	require.NoError(t, err)

	return msgID
}

func CreateDeliveredMsg(t *testing.T, app *appbuilder.App, queue string, payload string) string {
	t.Helper()

	msgID := CreateProcessingMsg(t, app, queue, payload, 100)

	err := app.AckMessages.Do(context.Background(), []usecases.AckParams{{ID: msgID}})
	require.NoError(t, err)

	return msgID
}
