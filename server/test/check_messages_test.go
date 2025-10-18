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

		msgHistoryQueue = "test.result"
	)

	// Arrange
	msg1ID := e2eutils.CreateArchivedMsg(t, app, msgQueue, msgPayload)
	msg2ID := e2eutils.CreateMsg(t, app, msgQueue, msgPayload, msgPriority)
	msg3ID := e2eutils.CreateAvailableMsgWithHistory(t, app, msgHistoryQueue, msgQueue, msgPayload)

	// Act
	requestBody := []string{msg1ID, msg2ID, msg3ID}
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/messages/check", bytes.NewBuffer(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	// Assert
	require.Equal(t, http.StatusOK, resp.Code, resp.Body.String())

	var respDTO []struct {
		ID          string     `json:"id"`
		Queue       string     `json:"queue"`
		CreatedAt   time.Time  `json:"created_at"`
		FinalizedAt *time.Time `json:"finalized_at"`
		Status      string     `json:"status"`
		Priority    int        `json:"priority"`
		Retries     int        `json:"retries"`
		Generation  int        `json:"generation"`
		History     []struct {
			Generation   int       `json:"generation"`
			Queue        string    `json:"queue"`
			RedirectedAt time.Time `json:"redirected_at"`
			Priority     int       `json:"priority"`
			Retries      int       `json:"retries"`
		} `json:"history"`
		Payload string `json:"payload"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respDTO)
	require.NoError(t, err)

	require.Len(t, respDTO, 3)

	require.Equal(t, msg1ID, respDTO[0].ID)
	require.Equal(t, msgQueue, respDTO[0].Queue)
	require.Equal(t, app.Clock.Now(), respDTO[0].CreatedAt)
	require.NotNil(t, respDTO[0].FinalizedAt)
	require.Equal(t, app.Clock.Now(), *respDTO[0].FinalizedAt)
	require.Equal(t, string(domain.MsgStatusDelivered), respDTO[0].Status)
	require.Equal(t, msgPriority, respDTO[0].Priority)
	require.Equal(t, 0, respDTO[0].Retries)
	require.Equal(t, 0, respDTO[0].Generation)
	require.Len(t, respDTO[0].History, 0)
	require.Equal(t, msgPayload, respDTO[0].Payload)

	require.Equal(t, msg2ID, respDTO[1].ID)
	require.Equal(t, msgQueue, respDTO[1].Queue)
	require.Equal(t, app.Clock.Now(), respDTO[1].CreatedAt)
	require.Nil(t, respDTO[1].FinalizedAt)
	require.Equal(t, string(domain.MsgStatusPrepared), respDTO[1].Status)
	require.Equal(t, msgPriority, respDTO[1].Priority)
	require.Equal(t, 0, respDTO[1].Retries)
	require.Equal(t, 0, respDTO[1].Generation)
	require.Len(t, respDTO[1].History, 0)
	require.Equal(t, msgPayload, respDTO[1].Payload)

	require.Equal(t, msg3ID, respDTO[2].ID)
	require.Equal(t, msgQueue, respDTO[2].Queue)
	require.Equal(t, app.Clock.Now(), respDTO[2].CreatedAt)
	require.Nil(t, respDTO[2].FinalizedAt)
	require.Equal(t, string(domain.MsgStatusAvailable), respDTO[2].Status)
	require.Equal(t, msgPriority, respDTO[2].Priority)
	require.Equal(t, 0, respDTO[2].Retries)
	require.Equal(t, 1, respDTO[2].Generation)
	require.Len(t, respDTO[2].History, 1)
	require.Equal(t, 0, respDTO[2].History[0].Generation)
	require.Equal(t, msgHistoryQueue, respDTO[2].History[0].Queue)
	require.Equal(t, app.Clock.Now(), respDTO[2].History[0].RedirectedAt)
	require.Equal(t, msgPriority, respDTO[2].History[0].Priority)
	require.Equal(t, 0, respDTO[2].History[0].Retries)
	require.Equal(t, msgPayload, respDTO[2].Payload)
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
	require.Equal(t, http.StatusInternalServerError, resp.Code, resp.Body.String())

	var responseDTO e2eutils.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&responseDTO)
	require.NoError(t, err)

	require.Contains(t, responseDTO.Error, "message not found")
}
