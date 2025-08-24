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

func TestReleaseMessage(t *testing.T) {
	app, _ := e2eutils.Prepare(t)

	const (
		msgQueue    = "test"
		msgPayload  = `{"arg": 123}`
		msgPriority = 100
	)

	// Arrange
	msgID := e2eutils.CreateMsg(t, app, msgQueue, msgPayload, msgPriority)

	// Act
	requestBody := []string{msgID}
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/messages/release", bytes.NewBuffer(body))
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

	require.Equal(t, msgQueue, message.Queue())
	require.Equal(t, msgPayload, message.Payload())
	require.Equal(t, msgPriority, message.Priority())
	require.Equal(t, app.Clock.Now(), message.CreatedAt())
	require.Equal(t, domain.MsgStatusReady, message.Status())
}
