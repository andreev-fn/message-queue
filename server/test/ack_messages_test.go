package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/pkg/apierror"
	"server/pkg/httpmodels"
	"server/test/e2eutils"
	"server/test/fixtures"
)

func TestAckMessages(t *testing.T) {
	app, _ := e2eutils.Prepare(t)
	client := e2eutils.PrepareHTTPClient(t, app)

	const (
		msgQueue    = "test"
		msgPayload  = `{"arg": 123}`
		msgPriority = 100
	)

	// Arrange
	msgID := fixtures.CreateProcessingMsg(app, msgQueue, msgPayload, msgPriority)

	// Act
	err := client.AckMessages(httpmodels.AckRequest{
		httpmodels.AckRequestItem{
			ID: msgID,
		},
	})

	// Assert response
	require.NoError(t, err)

	// Assert the message in DB
	message, err := app.MsgRepo.GetByID(context.Background(), app.DB, msgID)
	require.NoError(t, err)
	require.Equal(t, domain.MsgStatusDelivered, message.Status())
}

func TestAckMessagesAtomicRelease(t *testing.T) {
	app, _ := e2eutils.Prepare(t)
	client := e2eutils.PrepareHTTPClient(t, app)

	const (
		msgToAckQueue    = "test"
		msgToAckPayload  = `{"a": 2, "b": 2}`
		msgToAckPriority = 100

		msgToReleaseQueue    = "test.result"
		msgToReleasePayload  = `{"sum": 4}`
		msgToReleasePriority = 100
	)

	// Arrange
	msgToAckID := fixtures.CreateProcessingMsg(app, msgToAckQueue, msgToAckPayload, msgToAckPriority)
	msgToReleaseID := fixtures.CreateMsg(app, msgToReleaseQueue, msgToReleasePayload, msgToReleasePriority)

	// Act
	err := client.AckMessages(httpmodels.AckRequest{
		httpmodels.AckRequestItem{
			ID:      msgToAckID,
			Release: []httpmodels.MessageID{msgToReleaseID},
		},
	})

	// Assert response
	require.NoError(t, err)

	// Assert messages in DB
	ackedMessage, err := app.MsgRepo.GetByID(context.Background(), app.DB, msgToAckID)
	require.NoError(t, err)
	require.Equal(t, domain.MsgStatusDelivered, ackedMessage.Status())

	releasedMessage, err := app.MsgRepo.GetByID(context.Background(), app.DB, msgToReleaseID)
	require.NoError(t, err)
	require.Equal(t, domain.MsgStatusAvailable, releasedMessage.Status())
}

func TestAckUnknownMessage(t *testing.T) {
	app, _ := e2eutils.Prepare(t)
	client := e2eutils.PrepareHTTPClient(t, app)

	// Act
	err := client.AckMessages(httpmodels.AckRequest{
		httpmodels.AckRequestItem{
			ID: "d8d4d0f7-1bbd-48c0-9f80-c66f5fd45fc2",
		},
	})

	// Assert
	require.True(t, apierror.IsCode(err, apierror.CodeMessageNotFound))
}
