package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/test/e2eutils"
)

func TestResumeDelayedAfterTimeout(t *testing.T) {
	app, clock := e2eutils.Prepare(t)

	const (
		msgQueue   = "test"
		msgPayload = `{"arg": 123}`
	)

	// Arrange
	msgID := e2eutils.CreateDelayedMsg(t, app, msgQueue, msgPayload)
	clock.Set(clock.Now().Add(3 * time.Minute))

	// Act
	affected, err := app.ResumeDelayed.Do(context.Background(), 10)
	require.NoError(t, err)

	// Assert
	require.Equal(t, 1, affected)

	updatedMsg, err := app.MsgRepo.GetByID(context.Background(), app.DB, msgID)
	require.NoError(t, err)

	require.Equal(t, domain.MsgStatusReady, updatedMsg.Status())
	require.Equal(t, 1, updatedMsg.Retries())
}

func TestResumeDelayedBeforeTimeout(t *testing.T) {
	app, clock := e2eutils.Prepare(t)

	const (
		msgQueue   = "test"
		msgPayload = `{"arg": 123}`
	)

	// Arrange
	msgID := e2eutils.CreateDelayedMsg(t, app, msgQueue, msgPayload)
	clock.Set(clock.Now().Add(1 * time.Minute))

	// Act
	affected, err := app.ResumeDelayed.Do(context.Background(), 10)
	require.NoError(t, err)

	// Assert
	require.Equal(t, 0, affected)

	unchangedMsg, err := app.MsgRepo.GetByID(context.Background(), app.DB, msgID)
	require.NoError(t, err)

	require.Equal(t, domain.MsgStatusDelayed, unchangedMsg.Status())
	require.Equal(t, 1, unchangedMsg.Retries())
}
