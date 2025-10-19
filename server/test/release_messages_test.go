package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/pkg/httpmodels"
	"server/test/e2eutils"
)

func TestReleaseMessage(t *testing.T) {
	app, _ := e2eutils.Prepare(t)
	client := e2eutils.PrepareHTTPClient(t, app)

	const (
		msgQueue    = "test"
		msgPayload  = `{"arg": 123}`
		msgPriority = 100
	)

	// Arrange
	msgID := e2eutils.CreateMsg(t, app, msgQueue, msgPayload, msgPriority)

	// Act
	err := client.ReleaseMessages(httpmodels.ReleaseRequest{msgID})

	// Assert response
	require.NoError(t, err)

	// Assert the message in DB
	message, err := app.MsgRepo.GetByID(context.Background(), app.DB, msgID)
	require.NoError(t, err)

	require.Equal(t, msgQueue, message.Queue().String())
	require.Equal(t, msgPayload, message.Payload())
	require.Equal(t, msgPriority, message.Priority())
	require.Equal(t, app.Clock.Now(), message.CreatedAt())
	require.Equal(t, domain.MsgStatusAvailable, message.Status())
}
