package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/internal/utils"
	"server/pkg/apierror"
	"server/pkg/httpmodels"
	"server/test/e2eutils"
	"server/test/fixtures"
)

func TestNackMessages(t *testing.T) {
	app, _ := e2eutils.Prepare(t)
	client := e2eutils.PrepareHTTPClient(t, app)

	// Arrange
	msgID := fixtures.CreateProcessingMsg(app)

	// Act
	err := client.NackMessages(httpmodels.NackRequest{
		httpmodels.NackRequestItem{ID: msgID},
	})

	// Assert response
	require.NoError(t, err)

	// Assert the message in DB
	message, err := app.MsgRepo.GetByID(context.Background(), app.DB, msgID)
	require.NoError(t, err)
	require.Equal(t, domain.MsgStatusDelayed, message.Status())
}

func TestNackMessagesNoRedeliver(t *testing.T) {
	app, _ := e2eutils.Prepare(t)
	client := e2eutils.PrepareHTTPClient(t, app)

	// Arrange
	msgID := fixtures.CreateProcessingMsg(app)

	// Act
	err := client.NackMessages(httpmodels.NackRequest{
		httpmodels.NackRequestItem{ID: msgID, Redeliver: utils.P(false)},
	})

	// Assert response
	require.NoError(t, err)

	// Assert the message in DB
	message, err := app.MsgRepo.GetByID(context.Background(), app.DB, msgID)
	require.NoError(t, err)
	require.Equal(t, domain.MsgStatusDropped, message.Status())
}

func TestNackUnknownMessage(t *testing.T) {
	app, _ := e2eutils.Prepare(t)
	client := e2eutils.PrepareHTTPClient(t, app)

	// Act
	err := client.NackMessages(httpmodels.NackRequest{
		httpmodels.NackRequestItem{ID: "d8d4d0f7-1bbd-48c0-9f80-c66f5fd45fc2", Redeliver: utils.P(false)},
	})

	// Assert
	require.True(t, apierror.IsCode(err, apierror.CodeMessageNotFound))
}
