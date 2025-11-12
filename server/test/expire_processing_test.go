package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/internal/utils/testutils"
	"server/test/e2eutils"
	"server/test/fixtures"
)

func TestExpireProcessingAfterTimeout(t *testing.T) {
	testutils.SkipIfNotIntegration(t)

	app := e2eutils.Prepare()

	// Arrange
	msgID := fixtures.CreateProcessingMsg(app)
	e2eutils.AdvanceClock(app, 6*time.Minute)

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
	testutils.SkipIfNotIntegration(t)

	app := e2eutils.Prepare()

	// Arrange
	msgID := fixtures.CreateProcessingMsg(app)
	e2eutils.AdvanceClock(app, 3*time.Minute)

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
