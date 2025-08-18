package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/test/e2eutils"
)

func TestExpireProcessingAfterTimeout(t *testing.T) {
	app, clock := e2eutils.Prepare(t)

	const (
		msgQueue    = "test"
		msgPayload  = `{"arg": 123}`
		msgPriority = 100
	)

	// Arrange
	msgID := e2eutils.CreateProcessingMsg(t, app, msgQueue, msgPayload, msgPriority)
	clock.Set(clock.Now().Add(6 * time.Minute))

	// Act
	affected, err := app.ExpireProcessing.Do(context.Background(), 10)
	require.NoError(t, err)

	// Assert
	require.Equal(t, 1, affected)

	updatedMsg, err := app.MsgRepo.GetByID(context.Background(), app.DB, msgID)
	require.NoError(t, err)

	require.Equal(t, domain.MsgStatusDelayed, updatedMsg.Status())
	require.Equal(t, 1, updatedMsg.Retries())
}

func TestExpireProcessingBeforeTimeout(t *testing.T) {
	app, clock := e2eutils.Prepare(t)

	const (
		msgQueue    = "test"
		msgPayload  = `{"arg": 123}`
		msgPriority = 100
	)

	// Arrange
	msgID := e2eutils.CreateProcessingMsg(t, app, msgQueue, msgPayload, msgPriority)
	clock.Set(clock.Now().Add(3 * time.Minute))

	// Act
	affected, err := app.ExpireProcessing.Do(context.Background(), 10)
	require.NoError(t, err)

	// Assert
	require.Equal(t, 0, affected)

	unchangedMsg, err := app.MsgRepo.GetByID(context.Background(), app.DB, msgID)
	require.NoError(t, err)

	require.Equal(t, domain.MsgStatusProcessing, unchangedMsg.Status())
	require.Equal(t, 0, unchangedMsg.Retries())
}
