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
)

func TestConsumeMessages(t *testing.T) {
	app, _ := e2eutils.Prepare(t)
	client := e2eutils.PrepareHTTPClient(t, app)

	const (
		msgQueue = "test"

		msg1Payload  = `{"arg": 123}`
		msg1Priority = 10

		msg2Payload  = `{"arg": 213}`
		msg2Priority = 200

		msg3Payload  = `{"arg": 321}`
		msg3Priority = 100
	)

	// Arrange
	msg1ID := e2eutils.CreateAvailableMsg(app, msgQueue, msg1Payload, msg1Priority)
	msg2ID := e2eutils.CreateAvailableMsg(app, msgQueue, msg2Payload, msg2Priority)
	msg3ID := e2eutils.CreateAvailableMsg(app, msgQueue, msg3Payload, msg3Priority)

	// Act
	respDTO, err := client.ConsumeMessages(httpmodels.ConsumeRequest{
		Queue: msgQueue,
		Limit: utils.P(1),
	})

	// Assert response
	require.NoError(t, err)

	require.Len(t, respDTO, 1)
	require.Equal(t, httpmodels.ConsumeResponseItem{
		ID:      msg2ID,
		Payload: msg2Payload,
	}, respDTO[0])

	// Assert messages in DB
	takenMsg, err := app.MsgRepo.GetByID(context.Background(), app.DB, msg2ID)
	require.NoError(t, err)
	require.Equal(t, domain.MsgStatusProcessing, takenMsg.Status())

	msg1, err := app.MsgRepo.GetByID(context.Background(), app.DB, msg1ID)
	require.NoError(t, err)
	require.Equal(t, domain.MsgStatusAvailable, msg1.Status())

	msg3, err := app.MsgRepo.GetByID(context.Background(), app.DB, msg3ID)
	require.NoError(t, err)
	require.Equal(t, domain.MsgStatusAvailable, msg3.Status())
}

func TestConsumeMessagesEmptyQueue(t *testing.T) {
	app, _ := e2eutils.Prepare(t)
	client := e2eutils.PrepareHTTPClient(t, app)

	// Act
	respDTO, err := client.ConsumeMessages(httpmodels.ConsumeRequest{
		Queue: "test",
		Limit: utils.P(1),
	})

	// Assert
	require.NoError(t, err)

	require.Empty(t, respDTO)
}

func TestConsumeFromUnknownQueue(t *testing.T) {
	app, _ := e2eutils.Prepare(t)
	client := e2eutils.PrepareHTTPClient(t, app)

	// Act
	_, err := client.ConsumeMessages(httpmodels.ConsumeRequest{
		Queue: "undefined_queue",
		Limit: utils.P(1),
	})

	// Assert
	require.True(t, apierror.IsCode(err, apierror.CodeQueueNotFound))
}
