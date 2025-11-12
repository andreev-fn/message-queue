package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/pkg/apierror"
	"server/pkg/httpmodels"
	"server/test/e2eutils"
	"server/test/fixtures"
)

func TestRedirectMessages(t *testing.T) {
	app, clock := e2eutils.Prepare(t)
	client := e2eutils.PrepareHTTPClient(t, app)

	const destinationQueue = "all_results"

	// Arrange
	msgID := fixtures.CreateProcessingMsg(app)
	clock.Set(clock.Now().Add(time.Minute))

	// Act
	err := client.RedirectMessages(httpmodels.RedirectRequest{
		httpmodels.RedirectRequestItem{
			ID:          msgID,
			Destination: destinationQueue,
		},
	})

	// Assert response
	require.NoError(t, err)

	// Assert the message in DB
	message, err := app.MsgRepo.GetByIDWithHistory(context.Background(), app.DB, msgID)
	require.NoError(t, err)

	require.Equal(t, domain.MsgStatusAvailable, message.Status())
	require.Equal(t, destinationQueue, message.Queue().String())
	require.Equal(t, 1, message.Generation())

	chapters, loaded := message.History().Chapters()
	require.True(t, loaded)
	require.Len(t, chapters, 1)
	require.Equal(t, fixtures.DefaultMsgQueue, chapters[0].Queue().String())
	require.Equal(t, 0, chapters[0].Generation())
}

func TestRedirectToUnknownQueue(t *testing.T) {
	app, _ := e2eutils.Prepare(t)
	client := e2eutils.PrepareHTTPClient(t, app)

	// Arrange
	msgID := fixtures.CreateProcessingMsg(app)

	// Act
	err := client.RedirectMessages(httpmodels.RedirectRequest{
		httpmodels.RedirectRequestItem{
			ID:          msgID,
			Destination: "unknown_queue",
		},
	})

	// Assert
	require.True(t, apierror.IsCode(err, apierror.CodeQueueNotFound))
}

func TestRedirectUnknownMessage(t *testing.T) {
	app, _ := e2eutils.Prepare(t)
	client := e2eutils.PrepareHTTPClient(t, app)

	// Act
	err := client.RedirectMessages(httpmodels.RedirectRequest{
		httpmodels.RedirectRequestItem{
			ID:          "d8d4d0f7-1bbd-48c0-9f80-c66f5fd45fc2",
			Destination: "all_results",
		},
	})

	// Assert
	require.True(t, apierror.IsCode(err, apierror.CodeMessageNotFound))
}
