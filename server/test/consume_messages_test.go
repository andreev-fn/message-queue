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

func TestConsumeMessages(t *testing.T) {
	app, _ := e2eutils.Prepare(t)

	const (
		msgQueue = "test"

		msg1Payload  = `{"arg": 123}`
		msg1Priority = 10

		msg2Payload  = `{"arg": 213}`
		msg2Priority = 200

		msg3Payload  = `{"arg": 321}`
		msg3Priority = 100
	)

	// Arrange
	msg1ID := e2eutils.CreateReadyMsg(t, app, msgQueue, msg1Payload, msg1Priority)
	msg2ID := e2eutils.CreateReadyMsg(t, app, msgQueue, msg2Payload, msg2Priority)
	msg3ID := e2eutils.CreateReadyMsg(t, app, msgQueue, msg3Payload, msg3Priority)

	// Act
	requestBody := map[string]any{
		"queue": msgQueue,
		"limit": 1,
	}
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/messages/consume", bytes.NewBuffer(body))
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
	require.NotNil(t, respWrapper.Result)
	require.Nil(t, respWrapper.Error)

	var results []struct {
		ID      string `json:"id"`
		Payload string `json:"payload"`
	}
	err = json.Unmarshal(*respWrapper.Result, &results)
	require.NoError(t, err)

	require.Len(t, results, 1)
	require.Equal(t, msg2ID, results[0].ID)
	require.JSONEq(t, msg2Payload, results[0].Payload)

	// Assert messages in DB
	takenMsg, err := app.MsgRepo.GetByID(context.Background(), app.DB, msg2ID)
	require.NoError(t, err)
	require.Equal(t, domain.MsgStatusProcessing, takenMsg.Status())

	msg1, err := app.MsgRepo.GetByID(context.Background(), app.DB, msg1ID)
	require.NoError(t, err)
	require.Equal(t, domain.MsgStatusReady, msg1.Status())

	msg3, err := app.MsgRepo.GetByID(context.Background(), app.DB, msg3ID)
	require.NoError(t, err)
	require.Equal(t, domain.MsgStatusReady, msg3.Status())
}

func TestConsumeMessagesEmptyQueue(t *testing.T) {
	app, _ := e2eutils.Prepare(t)

	// Act
	requestBody := map[string]any{
		"queue": "test",
		"limit": 1,
	}
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/messages/consume", bytes.NewBuffer(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	// Assert
	require.Equal(t, http.StatusOK, resp.Result().StatusCode)

	var respWrapper e2eutils.ResponseWrapper
	err = json.NewDecoder(resp.Body).Decode(&respWrapper)
	require.NoError(t, err)

	require.True(t, respWrapper.Success)
	require.NotNil(t, respWrapper.Result)
	require.Nil(t, respWrapper.Error)

	var results []any
	err = json.Unmarshal(*respWrapper.Result, &results)
	require.NoError(t, err)

	require.Empty(t, results, "Expected empty result list for empty queue")
}
