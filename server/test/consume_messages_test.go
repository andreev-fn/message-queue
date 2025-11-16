package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/internal/utils"
	"server/internal/utils/testutils"
	"server/pkg/apierror"
	"server/pkg/httpmodels"
	"server/test/e2eutils"
	"server/test/fixtures"
)

func TestConsumeMessages(t *testing.T) {
	testutils.SkipIfNotIntegration(t)

	app := e2eutils.Prepare()
	client := e2eutils.PrepareHTTPClient(t, app)

	const msg2Payload = `{"arg": 213}`

	// Arrange
	msg1ID := fixtures.CreateAvailableMsg(app, fixtures.WithPriority(10))
	msg2ID := fixtures.CreateAvailableMsg(app, fixtures.WithPriority(200), fixtures.WithPayload(msg2Payload))
	msg3ID := fixtures.CreateAvailableMsg(app, fixtures.WithPriority(100))

	// Act
	respDTO, err := client.ConsumeMessages(httpmodels.ConsumeRequest{
		Queue: fixtures.DefaultMsgQueue,
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
	testutils.SkipIfNotIntegration(t)

	app := e2eutils.Prepare()
	client := e2eutils.PrepareHTTPClient(t, app)

	// Act
	respDTO, err := client.ConsumeMessages(httpmodels.ConsumeRequest{
		Queue: fixtures.DefaultMsgQueue,
		Limit: utils.P(1),
	})

	// Assert
	require.NoError(t, err)

	require.Empty(t, respDTO)
}

func TestConsumeFromUnknownQueue(t *testing.T) {
	testutils.SkipIfNotIntegration(t)

	app := e2eutils.Prepare()
	client := e2eutils.PrepareHTTPClient(t, app)

	// Act
	_, err := client.ConsumeMessages(httpmodels.ConsumeRequest{
		Queue: "undefined_queue",
		Limit: utils.P(1),
	})

	// Assert
	require.True(t, apierror.IsCode(err, apierror.CodeQueueNotFound))
}

func TestConsumeFromDLQAllowed(t *testing.T) {
	testutils.SkipIfNotIntegration(t)

	app := e2eutils.Prepare(e2eutils.WithDeadLettering())
	client := e2eutils.PrepareHTTPClient(t, app)

	// Arrange
	msgID := fixtures.CreateAvailableMsg(
		app,
		fixtures.WithQueue(e2eutils.GetDLQ(fixtures.DefaultMsgQueue)),
		fixtures.WithHistory(fixtures.DefaultMsgQueue),
	)

	// Act
	respDTO, err := client.ConsumeMessages(httpmodels.ConsumeRequest{
		Queue: e2eutils.GetDLQ(fixtures.DefaultMsgQueue),
		Limit: utils.P(1),
	})

	// Assert response
	require.NoError(t, err)

	require.Len(t, respDTO, 1)
	require.Equal(t, httpmodels.ConsumeResponseItem{
		ID:      msgID,
		Payload: fixtures.DefaultMsgPayload,
	}, respDTO[0])

	// Assert messages in DB
	takenMsg, err := app.MsgRepo.GetByID(context.Background(), app.DB, msgID)
	require.NoError(t, err)
	require.Equal(t, domain.MsgStatusProcessing, takenMsg.Status())
}
