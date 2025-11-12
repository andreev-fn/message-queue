package test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"server/internal/utils"
	"server/pkg/apierror"
	"server/pkg/httpmodels"
	"server/test/e2eutils"
	"server/test/fixtures"
)

func TestCheckExistingMessage(t *testing.T) {
	app := e2eutils.Prepare(t)
	client := e2eutils.PrepareHTTPClient(t, app)

	const msgHistoryQueue = "test.result"

	// Arrange
	msg1ID := fixtures.CreateArchivedMsg(app)
	msg2ID := fixtures.CreatePreparedMsg(app)
	msg3ID := fixtures.CreateAvailableMsg(app, fixtures.WithHistory(msgHistoryQueue))

	// Act
	respDTO, err := client.CheckMessages(httpmodels.CheckRequest{msg1ID, msg2ID, msg3ID})

	// Assert
	require.NoError(t, err)

	require.Len(t, respDTO, 3)

	require.Equal(t, httpmodels.Message{
		ID:          msg1ID,
		Queue:       fixtures.DefaultMsgQueue,
		CreatedAt:   app.Clock.Now(),
		FinalizedAt: utils.P(app.Clock.Now()),
		Status:      httpmodels.MsgStatusDelivered,
		Priority:    fixtures.DefaultMsgPriority,
		Retries:     0,
		Generation:  0,
		History:     []httpmodels.MessageChapter{},
		Payload:     fixtures.DefaultMsgPayload,
	}, respDTO[0])

	require.Equal(t, httpmodels.Message{
		ID:          msg2ID,
		Queue:       fixtures.DefaultMsgQueue,
		CreatedAt:   app.Clock.Now(),
		FinalizedAt: nil,
		Status:      httpmodels.MsgStatusPrepared,
		Priority:    fixtures.DefaultMsgPriority,
		Retries:     0,
		Generation:  0,
		History:     []httpmodels.MessageChapter{},
		Payload:     fixtures.DefaultMsgPayload,
	}, respDTO[1])

	require.Equal(t, httpmodels.Message{
		ID:          msg3ID,
		Queue:       fixtures.DefaultMsgQueue,
		CreatedAt:   app.Clock.Now(),
		FinalizedAt: nil,
		Status:      httpmodels.MsgStatusAvailable,
		Priority:    fixtures.DefaultMsgPriority,
		Retries:     0,
		Generation:  1,
		History: []httpmodels.MessageChapter{
			{
				Generation:   0,
				Queue:        msgHistoryQueue,
				RedirectedAt: app.Clock.Now(),
				Priority:     fixtures.DefaultMsgPriority,
				Retries:      0,
			},
		},
		Payload: fixtures.DefaultMsgPayload,
	}, respDTO[2])
}

func TestCheckUnknownMessage(t *testing.T) {
	app := e2eutils.Prepare(t)
	client := e2eutils.PrepareHTTPClient(t, app)

	const nonExistentID = "00000000-0000-0000-0000-000000000002"

	// Act
	_, err := client.CheckMessages(httpmodels.CheckRequest{nonExistentID})

	// Assert
	require.True(t, apierror.IsCode(err, apierror.CodeMessageNotFound))
}
