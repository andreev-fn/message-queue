package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/test/e2eutils"
)

func TestNackMessages(t *testing.T) {
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

	req, err := http.NewRequest(http.MethodPost, "/messages/nack", bytes.NewBuffer(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	// Assert response
	require.Equal(t, http.StatusOK, resp.Result().StatusCode)

	var respWrapper e2eutils.ResponseWrapper
	err = json.NewDecoder(resp.Body).Decode(&respWrapper)
	require.NoError(t, err)

	require.True(t, respWrapper.Success)
	require.Nil(t, respWrapper.Result)
	require.Nil(t, respWrapper.Error)

	// Assert the message in DB
	message, err := app.MsgRepo.GetByID(context.Background(), app.DB, msgID)
	require.NoError(t, err)
	require.Equal(t, domain.MsgStatusDelayed, message.Status())
}

func TestNackMessagesNoRedeliver(t *testing.T) {
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
			"id":        msgID,
			"redeliver": false,
		},
	}
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/messages/nack", bytes.NewBuffer(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	// Assert response
	require.Equal(t, http.StatusOK, resp.Result().StatusCode)

	var respWrapper e2eutils.ResponseWrapper
	err = json.NewDecoder(resp.Body).Decode(&respWrapper)
	require.NoError(t, err)

	require.True(t, respWrapper.Success)
	require.Nil(t, respWrapper.Result)
	require.Nil(t, respWrapper.Error)

	// Assert the message in DB
	message, err := app.MsgRepo.GetByID(context.Background(), app.DB, msgID)
	require.NoError(t, err)
	require.Equal(t, domain.MsgStatusUndeliverable, message.Status())
}
