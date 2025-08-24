package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/test/e2eutils"
)

func TestCheckExistingMessage(t *testing.T) {
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

	req, err := http.NewRequest(http.MethodPost, "/messages/check", bytes.NewBuffer(body))
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

	var respDTO []struct {
		ID        string          `json:"id"`
		Queue     string          `json:"queue"`
		CreatedAt time.Time       `json:"created_at"`
		Status    string          `json:"status"`
		Retries   int             `json:"retries"`
		Payload   json.RawMessage `json:"payload"`
	}
	err = json.Unmarshal(*respWrapper.Result, &respDTO)
	require.NoError(t, err)

	require.Len(t, respDTO, 1)
	require.Equal(t, msgID, respDTO[0].ID)
	require.Equal(t, msgQueue, respDTO[0].Queue)
	require.Equal(t, app.Clock.Now(), respDTO[0].CreatedAt)
	require.Equal(t, string(domain.MsgStatusCreated), respDTO[0].Status)
	require.Equal(t, 0, respDTO[0].Retries)
	require.JSONEq(t, msgPayload, string(respDTO[0].Payload))
}

func TestCheckNonExistentMessage(t *testing.T) {
	app, _ := e2eutils.Prepare(t)

	const nonExistentID = "00000000-0000-0000-0000-000000000002"

	// Act
	requestBody := []string{nonExistentID}
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/messages/check", bytes.NewBuffer(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	// Assert
	require.Equal(t, http.StatusInternalServerError, resp.Result().StatusCode)

	var respWrapper e2eutils.ResponseWrapper
	err = json.NewDecoder(resp.Body).Decode(&respWrapper)
	require.NoError(t, err)

	require.False(t, respWrapper.Success)
	require.Nil(t, respWrapper.Result)
	require.NotNil(t, respWrapper.Error)
	require.Contains(t, *respWrapper.Error, "message not found")
}
