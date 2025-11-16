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

func TestPrepareMessage(t *testing.T) {
	testutils.SkipIfNotIntegration(t)

	app := e2eutils.Prepare()
	client := e2eutils.PrepareHTTPClient(t, app)

	const (
		msgQueue    = "test"
		msgPayload  = `{"arg": 123}`
		msgPriority = 100
	)

	// Act
	respDTO, err := client.PrepareMessages(httpmodels.PublishRequest{
		httpmodels.PublishRequestItem{
			Queue:   msgQueue,
			Payload: msgPayload,
		},
	})

	// Assert response
	require.NoError(t, err)
	require.Len(t, respDTO, 1)

	// Assert the message in DB
	message, err := app.MsgRepo.GetByID(context.Background(), app.DB, respDTO[0])
	require.NoError(t, err)

	require.Equal(t, msgQueue, message.Queue().String())
	require.Equal(t, msgPayload, message.Payload())
	require.Equal(t, app.Clock.Now(), message.CreatedAt())
	require.Equal(t, domain.MsgStatusPrepared, message.Status())
	require.Equal(t, msgPriority, message.Priority())
}

func TestPublishMessageWithPriority(t *testing.T) {
	testutils.SkipIfNotIntegration(t)

	app := e2eutils.Prepare()
	client := e2eutils.PrepareHTTPClient(t, app)

	const (
		msgQueue    = "test"
		msgPayload  = `{"arg": 123}`
		msgPriority = 5
	)

	// Act
	respDTO, err := client.PublishMessages(httpmodels.PublishRequest{
		httpmodels.PublishRequestItem{
			Queue:    msgQueue,
			Payload:  msgPayload,
			Priority: utils.P(msgPriority),
		},
	})

	// Assert response
	require.NoError(t, err)
	require.Len(t, respDTO, 1)

	// Assert the message in DB
	message, err := app.MsgRepo.GetByID(context.Background(), app.DB, respDTO[0])
	require.NoError(t, err)

	require.Equal(t, msgQueue, message.Queue().String())
	require.Equal(t, msgPayload, message.Payload())
	require.Equal(t, app.Clock.Now(), message.CreatedAt())
	require.Equal(t, domain.MsgStatusAvailable, message.Status())
	require.Equal(t, msgPriority, message.Priority())
}

func TestPublishToUnknownQueue(t *testing.T) {
	testutils.SkipIfNotIntegration(t)

	app := e2eutils.Prepare()
	client := e2eutils.PrepareHTTPClient(t, app)

	const (
		msgQueue    = "undefined_queue"
		msgPayload  = `{"arg": 123}`
		msgPriority = 5
	)

	// Act
	_, err := client.PublishMessages(httpmodels.PublishRequest{
		httpmodels.PublishRequestItem{
			Queue:    msgQueue,
			Payload:  msgPayload,
			Priority: utils.P(msgPriority),
		},
	})

	// Assert
	require.True(t, apierror.IsCode(err, apierror.CodeQueueNotFound))
}

func TestPublishToDLQNotAllowed(t *testing.T) {
	testutils.SkipIfNotIntegration(t)

	app := e2eutils.Prepare(e2eutils.WithDeadLettering())
	client := e2eutils.PrepareHTTPClient(t, app)

	// Act
	_, err := client.PublishMessages(httpmodels.PublishRequest{
		httpmodels.PublishRequestItem{
			Queue:    e2eutils.GetDLQ(fixtures.DefaultMsgQueue),
			Payload:  fixtures.DefaultMsgPayload,
			Priority: utils.P(fixtures.DefaultMsgPriority),
		},
	})

	// Assert
	require.ErrorContains(t, err, "writing directly to DLQ is not allowed")
}
