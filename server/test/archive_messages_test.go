package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/internal/storage"
	"server/test/e2eutils"
)

func TestArchiveMessagesFinalized(t *testing.T) {
	app, clock := e2eutils.Prepare(t)

	const (
		msgQueue   = "test"
		msgPayload = `{"arg": 123}`
		msgResult  = `{"result":"success"}`
	)

	// Arrange
	msgID := e2eutils.CreateCompletedMsg(t, app, msgQueue, msgPayload, msgResult)
	clock.Set(clock.Now().Add(time.Minute))

	// Act
	affected, err := app.ArchiveMessages.Do(context.Background(), 10)
	require.NoError(t, err)

	// Assert
	require.Equal(t, 1, affected)

	_, err = app.MsgRepo.GetByID(context.Background(), app.DB, msgID)
	require.ErrorIs(t, err, storage.ErrMsgNotFound)

	archivedMsg, err := app.ArchivedMsgRepo.GetByID(context.Background(), app.DB, msgID)
	require.NoError(t, err)

	require.Equal(t, domain.MsgStatusCompleted, archivedMsg.Status())
	require.Equal(t, msgQueue, archivedMsg.Queue())
	require.JSONEq(t, msgPayload, string(archivedMsg.Payload()))
	require.NotNil(t, archivedMsg.Result())
	require.JSONEq(t, msgResult, string(*archivedMsg.Result()))
}

func TestArchiveMessagesNotFinal(t *testing.T) {
	app, clock := e2eutils.Prepare(t)

	const (
		msgQueue    = "test"
		msgPayload  = `{"arg": 123}`
		msgPriority = 100
	)

	// Arrange
	msgID := e2eutils.CreateProcessingMsg(t, app, msgQueue, msgPayload, msgPriority)
	clock.Set(clock.Now().Add(time.Minute))

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
