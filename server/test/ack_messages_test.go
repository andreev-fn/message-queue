package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/internal/utils/testutils"
	"server/pkg/httpclient"
	"server/pkg/httpmodels"
	"server/test/fixtures"
	"server/test/testkit"
)

func TestAckMessages(t *testing.T) {
	testutils.SkipIfNotInTestEnv(t)

	app := testkit.NewApp(testkit.NewAppConfig())
	client := testkit.NewHTTPClient(t, app)
	testkit.CleanupDatabase(app.DB)

	// Arrange
	msgID := fixtures.CreateProcessingMsg(app)

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
	testutils.SkipIfNotInTestEnv(t)

	app := testkit.NewApp(testkit.NewAppConfig())
	client := testkit.NewHTTPClient(t, app)
	testkit.CleanupDatabase(app.DB)

	const msgToReleaseQueue = "test.result"

	// Arrange
	msgToAckID := fixtures.CreateProcessingMsg(app)
	msgToReleaseID := fixtures.CreatePreparedMsg(app, fixtures.WithQueue(msgToReleaseQueue))

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
	testutils.SkipIfNotInTestEnv(t)

	app := testkit.NewApp(testkit.NewAppConfig())
	client := testkit.NewHTTPClient(t, app)
	testkit.CleanupDatabase(app.DB)

	// Act
	err := client.AckMessages(httpmodels.AckRequest{
		httpmodels.AckRequestItem{
			ID: "d8d4d0f7-1bbd-48c0-9f80-c66f5fd45fc2",
		},
	})

	// Assert
	require.True(t, httpclient.IsCode(err, httpmodels.ErrorCodeMessageNotFound))
}
