package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/internal/utils/testutils"
	"server/pkg/apierror"
	"server/pkg/httpmodels"
	"server/test/e2eutils"
	"server/test/fixtures"
)

func TestReleaseMessage(t *testing.T) {
	testutils.SkipIfNotIntegration(t)

	app := e2eutils.Prepare()
	client := e2eutils.PrepareHTTPClient(t, app)

	// Arrange
	msgID := fixtures.CreatePreparedMsg(app)

	// Act
	err := client.ReleaseMessages(httpmodels.ReleaseRequest{msgID})

	// Assert response
	require.NoError(t, err)

	// Assert the message in DB
	message, err := app.MsgRepo.GetByID(context.Background(), app.DB, msgID)
	require.NoError(t, err)

	require.Equal(t, fixtures.DefaultMsgQueue, message.Queue().String())
	require.Equal(t, fixtures.DefaultMsgPayload, message.Payload())
	require.Equal(t, fixtures.DefaultMsgPriority, message.Priority())
	require.Equal(t, app.Clock.Now(), message.CreatedAt())
	require.Equal(t, domain.MsgStatusAvailable, message.Status())
}

func TestReleaseUnknownMessage(t *testing.T) {
	testutils.SkipIfNotIntegration(t)

	app := e2eutils.Prepare()
	client := e2eutils.PrepareHTTPClient(t, app)

	// Act
	err := client.ReleaseMessages(httpmodels.ReleaseRequest{"d8d4d0f7-1bbd-48c0-9f80-c66f5fd45fc2"})

	// Assert
	require.True(t, apierror.IsCode(err, apierror.CodeMessageNotFound))
}
