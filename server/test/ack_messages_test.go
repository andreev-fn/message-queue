package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/test/e2eutils"
)

func TestAckMessages(t *testing.T) {
	app, _ := e2eutils.Prepare(t)

	const (
		msgQueue    = "test"
		msgPayload  = `{"arg": 123}`
		msgPriority = 100
	)

	// Arrange
	msgID := e2eutils.CreateProcessingMsg(t, app, msgQueue, msgPayload, msgPriority)

	// Act
	requestBody := []any{
		map[string]any{
			"id": msgID,
		},
	}
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/messages/ack", bytes.NewBuffer(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	// Assert response
	require.Equal(t, http.StatusOK, resp.Code, resp.Body.String())
	assert.JSONEq(t, e2eutils.OkResponseJSON, resp.Body.String())

	// Assert the message in DB
	message, err := app.MsgRepo.GetByID(context.Background(), app.DB, msgID)
	require.NoError(t, err)
	require.Equal(t, domain.MsgStatusDelivered, message.Status())
}

func TestAckMessagesAtomicRelease(t *testing.T) {
	app, _ := e2eutils.Prepare(t)

	const (
		msgToAckQueue    = "test"
		msgToAckPayload  = `{"a": 2, "b": 2}`
		msgToAckPriority = 100

		msgToReleaseQueue    = "test.result"
		msgToReleasePayload  = `{"sum": 4}`
		msgToReleasePriority = 100
	)

	// Arrange
	msgToAckID := e2eutils.CreateProcessingMsg(t, app, msgToAckQueue, msgToAckPayload, msgToAckPriority)
	msgToReleaseID := e2eutils.CreateMsg(t, app, msgToReleaseQueue, msgToReleasePayload, msgToReleasePriority)

	// Act
	requestBody := []any{
		map[string]any{
			"id":      msgToAckID,
			"release": []string{msgToReleaseID},
		},
	}
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/messages/ack", bytes.NewBuffer(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	// Assert response
	require.Equal(t, http.StatusOK, resp.Code, resp.Body.String())
	assert.JSONEq(t, e2eutils.OkResponseJSON, resp.Body.String())

	// Assert messages in DB
	ackedMessage, err := app.MsgRepo.GetByID(context.Background(), app.DB, msgToAckID)
	require.NoError(t, err)
	require.Equal(t, domain.MsgStatusDelivered, ackedMessage.Status())

	releasedMessage, err := app.MsgRepo.GetByID(context.Background(), app.DB, msgToReleaseID)
	require.NoError(t, err)
	require.Equal(t, domain.MsgStatusAvailable, releasedMessage.Status())
}

func TestAckUnknownMessage(t *testing.T) {
	app, _ := e2eutils.Prepare(t)

	// Act
	requestBody := []any{
		map[string]any{
			"id": "d8d4d0f7-1bbd-48c0-9f80-c66f5fd45fc2",
		},
	}
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/messages/ack", bytes.NewBuffer(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	// Assert
	require.Equal(t, http.StatusInternalServerError, resp.Code, resp.Body.String())

	var respDTO e2eutils.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&respDTO)
	require.NoError(t, err)

	require.Contains(t, respDTO.Error, "message not found")
}
