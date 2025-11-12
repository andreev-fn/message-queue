package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/internal/storage"
	"server/test/e2eutils"
	"server/test/fixtures"
)

func TestArchiveMessagesFinalized(t *testing.T) {
	app := e2eutils.Prepare(t)

	// Arrange
	msgID := fixtures.CreateDeliveredMsg(app)
	e2eutils.AdvanceClock(app, time.Minute)

	// Act
	affected, err := app.ArchiveMessages.Do(context.Background(), 10)
	require.NoError(t, err)

	// Assert
	require.Equal(t, 1, affected)

	_, err = app.MsgRepo.GetByID(context.Background(), app.DB, msgID)
	require.ErrorIs(t, err, storage.ErrMsgNotFound)

	archivedMsg, err := app.ArchivedMsgRepo.GetByID(context.Background(), app.DB, msgID)
	require.NoError(t, err)

	require.Equal(t, domain.MsgStatusDelivered, archivedMsg.Status())
	require.Equal(t, fixtures.DefaultMsgQueue, archivedMsg.Queue().String())
	require.Equal(t, fixtures.DefaultMsgPayload, archivedMsg.Payload())
}

func TestArchiveMessagesNotFinal(t *testing.T) {
	app := e2eutils.Prepare(t)

	// Arrange
	msgID := fixtures.CreateProcessingMsg(app)
	e2eutils.AdvanceClock(app, time.Minute)

	// Act
	affected, err := app.ArchiveMessages.Do(context.Background(), 10)
	require.NoError(t, err)

	// Assert
	require.Equal(t, 0, affected)

	_, err = app.ArchivedMsgRepo.GetByID(context.Background(), app.DB, msgID)
	require.ErrorIs(t, err, storage.ErrArchivedMsgNotFound)

	unchangedMsg, err := app.MsgRepo.GetByID(context.Background(), app.DB, msgID)
	require.NoError(t, err)

	require.Equal(t, domain.MsgStatusProcessing, unchangedMsg.Status())
}
