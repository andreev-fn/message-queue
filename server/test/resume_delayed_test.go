package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/test/e2eutils"
	"server/test/fixtures"
)

func TestResumeDelayedAfterTimeout(t *testing.T) {
	app, clock := e2eutils.Prepare(t)

	const (
		msgQueue   = "test"
		msgPayload = `{"arg": 123}`
	)

	// Arrange
	msgID := fixtures.CreateDelayedMsg(app, msgQueue, msgPayload)
	clock.Set(clock.Now().Add(40 * time.Second))

	// Act
	affected, err := app.ResumeDelayed.Do(context.Background(), 10)
	require.NoError(t, err)

	// Assert
	require.Equal(t, 1, affected)

	updatedMsg, err := app.MsgRepo.GetByID(context.Background(), app.DB, msgID)
	require.NoError(t, err)

	require.Equal(t, domain.MsgStatusAvailable, updatedMsg.Status())
	require.Equal(t, 1, updatedMsg.Retries())
}

func TestResumeDelayedBeforeTimeout(t *testing.T) {
	app, clock := e2eutils.Prepare(t)

	const (
		msgQueue   = "test"
		msgPayload = `{"arg": 123}`
	)

	// Arrange
	msgID := fixtures.CreateDelayedMsg(app, msgQueue, msgPayload)
	clock.Set(clock.Now().Add(20 * time.Second))

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
