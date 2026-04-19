package test

import (
	"context"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"server/internal/config"
	"server/internal/domain"
	"server/internal/utils"
	"server/internal/utils/testutils"
	"server/pkg/httpclient"
	"server/pkg/httpmodels"
	"server/test/fixtures"
	"server/test/testkit"
)

func TestPrepareMessage(t *testing.T) {
	testutils.SkipIfNotInTestEnv(t)

	app := testkit.NewApp(testkit.NewAppConfig())
	client := testkit.NewHTTPClient(t, app)
	testkit.CleanupDatabase(app.DB)

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
	require.Len(t, respDTO.Results, 1)
	require.Nil(t, respDTO.Results[0].Error)

	// Assert the message in DB
	message, err := app.MsgRepo.GetByID(context.Background(), app.DB, respDTO.Results[0].Data.ID)
	require.NoError(t, err)

	require.Equal(t, msgQueue, message.Queue().String())
	require.Equal(t, msgPayload, message.Payload())
	require.Equal(t, app.Clock.Now(), message.CreatedAt())
	require.Equal(t, domain.MsgStatusPrepared, message.Status())
	require.Equal(t, msgPriority, message.Priority())
}

func TestPublishMessage(t *testing.T) {
	testutils.SkipIfNotInTestEnv(t)

	app := testkit.NewApp(testkit.NewAppConfig(testkit.WithDeadLettering()))
	client := testkit.NewHTTPClient(t, app)
	testkit.CleanupDatabase(app.DB)

	const (
		unknownMsgQueue = "undefined_queue"
		invalidMsgQueue = "invalid+queue"
	)

	// Act
	respDTO, err := client.PublishMessages(httpmodels.PublishRequest{
		httpmodels.PublishRequestItem{
			Queue:   fixtures.DefaultMsgQueue,
			Payload: fixtures.DefaultMsgPayload,
		},
		httpmodels.PublishRequestItem{
			Queue:    fixtures.DefaultMsgQueue,
			Payload:  fixtures.DefaultMsgPayload,
			Priority: utils.P(fixtures.DefaultMsgPriority + 1),
		},
		httpmodels.PublishRequestItem{
			Queue:   unknownMsgQueue,
			Payload: fixtures.DefaultMsgPayload,
		},
		httpmodels.PublishRequestItem{
			Queue:   invalidMsgQueue,
			Payload: fixtures.DefaultMsgPayload,
		},
		httpmodels.PublishRequestItem{
			Queue:   testkit.GetDLQ(fixtures.DefaultMsgQueue),
			Payload: fixtures.DefaultMsgPayload,
		},
	})

	// Assert response
	require.NoError(t, err)
	require.Len(t, respDTO.Results, 5)

	t.Run("creates message with default priority", func(t *testing.T) {
		require.Nil(t, respDTO.Results[0].Error)

		message, err := app.MsgRepo.GetByID(context.Background(), app.DB, respDTO.Results[0].Data.ID)
		require.NoError(t, err)

		require.Equal(t, fixtures.DefaultMsgQueue, message.Queue().String())
		require.Equal(t, fixtures.DefaultMsgPayload, message.Payload())
		require.Equal(t, app.Clock.Now(), message.CreatedAt())
		require.Equal(t, domain.MsgStatusAvailable, message.Status())
		require.Equal(t, fixtures.DefaultMsgPriority, message.Priority())
	})

	t.Run("creates message with custom priority", func(t *testing.T) {
		require.Nil(t, respDTO.Results[1].Error)

		message, err := app.MsgRepo.GetByID(context.Background(), app.DB, respDTO.Results[1].Data.ID)
		require.NoError(t, err)

		require.Equal(t, fixtures.DefaultMsgQueue, message.Queue().String())
		require.Equal(t, fixtures.DefaultMsgPayload, message.Payload())
		require.Equal(t, app.Clock.Now(), message.CreatedAt())
		require.Equal(t, domain.MsgStatusAvailable, message.Status())
		require.Equal(t, fixtures.DefaultMsgPriority+1, message.Priority())
	})

	t.Run("fails for unknown queue", func(t *testing.T) {
		require.NotNil(t, respDTO.Results[2].Error)
		require.True(t, httpclient.IsCode(respDTO.Results[2].Error, httpmodels.ErrorCodeQueueNotFound))
	})

	t.Run("fails for invalid queue name", func(t *testing.T) {
		require.NotNil(t, respDTO.Results[3].Error)
		require.True(t, httpclient.IsCode(respDTO.Results[3].Error, httpmodels.ErrorCodeRequestInvalid))
	})

	t.Run("fails for non-writable queue", func(t *testing.T) {
		require.NotNil(t, respDTO.Results[4].Error)
		require.True(t, httpclient.IsCode(respDTO.Results[4].Error, httpmodels.ErrorCodeQueueNotWritable))
	})
}

func TestPublishGeneralFailure(t *testing.T) {
	testutils.SkipIfNotInTestEnv(t)

	app := testkit.NewApp(testkit.NewAppConfig())
	client := testkit.NewHTTPClient(t, app)
	testkit.CleanupDatabase(app.DB)

	// Act
	_, err := client.PublishMessages(slices.Repeat(httpmodels.PublishRequest{
		httpmodels.PublishRequestItem{
			Queue:   fixtures.DefaultMsgQueue,
			Payload: fixtures.DefaultMsgPayload,
		},
	}, config.DefaultBatchSizeLimit+1))

	// Assert response
	require.True(t, httpclient.IsCode(err, httpmodels.ErrorCodeBatchSizeTooBig))
}
