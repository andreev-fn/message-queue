package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/internal/utils"
	"server/pkg/httpmodels"
	"server/test/e2eutils"
)

func TestNackMessages(t *testing.T) {
	app, _ := e2eutils.Prepare(t)
	client := e2eutils.PrepareHTTPClient(t, app)

	const (
		msgQueue    = "test"
		msgPayload  = `{"arg": 123}`
		msgPriority = 100
	)

	// Arrange
	msgID := e2eutils.CreateProcessingMsg(t, app, msgQueue, msgPayload, msgPriority)

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

	const (
		msgQueue    = "test"
		msgPayload  = `{"arg": 123}`
		msgPriority = 100
	)

	// Arrange
	msgID := e2eutils.CreateProcessingMsg(t, app, msgQueue, msgPayload, msgPriority)

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
