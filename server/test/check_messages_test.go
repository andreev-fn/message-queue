package test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"server/internal/utils"
	"server/pkg/httpmodels"
	"server/test/e2eutils"
)

func TestCheckExistingMessage(t *testing.T) {
	app, _ := e2eutils.Prepare(t)
	client := e2eutils.PrepareHTTPClient(t, app)

	const (
		msgQueue    = "test"
		msgPayload  = `{"arg": 123}`
		msgPriority = 100

		msgHistoryQueue = "test.result"
	)

	// Arrange
	msg1ID := e2eutils.CreateArchivedMsg(t, app, msgQueue, msgPayload)
	msg2ID := e2eutils.CreateMsg(t, app, msgQueue, msgPayload, msgPriority)
	msg3ID := e2eutils.CreateAvailableMsgWithHistory(t, app, msgHistoryQueue, msgQueue, msgPayload)

	// Act
	respDTO, err := client.CheckMessages(httpmodels.CheckRequest{msg1ID, msg2ID, msg3ID})

	// Assert
	require.NoError(t, err)

	require.Len(t, respDTO, 3)

	require.Equal(t, httpmodels.Message{
		ID:          msg1ID,
		Queue:       msgQueue,
		CreatedAt:   app.Clock.Now(),
		FinalizedAt: utils.P(app.Clock.Now()),
		Status:      httpmodels.MsgStatusDelivered,
		Priority:    msgPriority,
		Retries:     0,
		Generation:  0,
		History:     []httpmodels.MessageChapter{},
		Payload:     msgPayload,
	}, respDTO[0])

	require.Equal(t, httpmodels.Message{
		ID:          msg2ID,
		Queue:       msgQueue,
		CreatedAt:   app.Clock.Now(),
		FinalizedAt: nil,
		Status:      httpmodels.MsgStatusPrepared,
		Priority:    msgPriority,
		Retries:     0,
		Generation:  0,
		History:     []httpmodels.MessageChapter{},
		Payload:     msgPayload,
	}, respDTO[1])

	require.Equal(t, httpmodels.Message{
		ID:          msg3ID,
		Queue:       msgQueue,
		CreatedAt:   app.Clock.Now(),
		FinalizedAt: nil,
		Status:      httpmodels.MsgStatusAvailable,
		Priority:    msgPriority,
		Retries:     0,
		Generation:  1,
		History: []httpmodels.MessageChapter{
			{
				Generation:   0,
				Queue:        msgHistoryQueue,
				RedirectedAt: app.Clock.Now(),
				Priority:     msgPriority,
				Retries:      0,
			},
		},
		Payload: msgPayload,
	}, respDTO[2])
}

func TestCheckNonExistentMessage(t *testing.T) {
	app, _ := e2eutils.Prepare(t)
	client := e2eutils.PrepareHTTPClient(t, app)

	const nonExistentID = "00000000-0000-0000-0000-000000000002"

	// Act
	_, err := client.CheckMessages(httpmodels.CheckRequest{nonExistentID})

	// Assert
	require.ErrorContains(t, err, "message not found")
}
