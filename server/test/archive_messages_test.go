package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/internal/storage"
	"server/internal/utils/testutils"
	"server/test/fixtures"
	"server/test/testkit"
)

func TestArchiveMessagesFinalized(t *testing.T) {
	testutils.SkipIfNotInTestEnv(t)

	app := testkit.NewApp(testkit.NewAppConfig())
	testkit.CleanupDatabase(app.DB)

	// Arrange
	msgID := fixtures.CreateDeliveredMsg(app)
	testkit.AdvanceClock(app, time.Minute)

	// Act
	err := app.ArchiveMessages.Do(context.Background())
	require.NoError(t, err)

	// Assert
	_, err = app.MsgRepo.GetByID(context.Background(), app.DB, msgID)
	require.ErrorIs(t, err, storage.ErrMsgNotFound)

	archivedMsg, err := app.ArchivedMsgRepo.GetByID(context.Background(), app.DB, msgID)
	require.NoError(t, err)

	require.Equal(t, domain.MsgStatusDelivered, archivedMsg.Status())
	require.Equal(t, fixtures.DefaultMsgQueue, archivedMsg.Queue().String())
	require.Equal(t, fixtures.DefaultMsgPayload, archivedMsg.Payload())
}

func TestArchiveMessagesNotFinal(t *testing.T) {
	testutils.SkipIfNotInTestEnv(t)

	app := testkit.NewApp(testkit.NewAppConfig())
	testkit.CleanupDatabase(app.DB)

	// Arrange
	msgID := fixtures.CreateProcessingMsg(app)
	testkit.AdvanceClock(app, time.Minute)

	// Act
	err := app.ArchiveMessages.Do(context.Background())
	require.NoError(t, err)

	// Assert
	_, err = app.ArchivedMsgRepo.GetByID(context.Background(), app.DB, msgID)
	require.ErrorIs(t, err, storage.ErrArchivedMsgNotFound)

	unchangedMsg, err := app.MsgRepo.GetByID(context.Background(), app.DB, msgID)
	require.NoError(t, err)

	require.Equal(t, domain.MsgStatusProcessing, unchangedMsg.Status())
}
