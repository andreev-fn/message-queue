package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/test/e2eutils"
)

func TestRedirectMessages(t *testing.T) {
	app, clock := e2eutils.Prepare(t)

	const (
		msgQueue    = "test.result"
		msgPayload  = `{"arg": 123}`
		msgPriority = 100

		destinationQueue = "all_results"
	)

	// Arrange
	msgID := e2eutils.CreateProcessingMsg(t, app, msgQueue, msgPayload, msgPriority)
	clock.Set(clock.Now().Add(time.Minute))

	// Act
	requestBody := []any{
		map[string]any{
			"id":          msgID,
			"destination": destinationQueue,
		},
	}
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/messages/redirect", bytes.NewBuffer(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	// Assert response
	require.Equal(t, http.StatusOK, resp.Code, resp.Body.String())
	assert.JSONEq(t, e2eutils.OkResponseJSON, resp.Body.String())

	// Assert the message in DB
	message, err := app.MsgRepo.GetByIDWithHistory(context.Background(), app.DB, msgID)
	require.NoError(t, err)

	require.Equal(t, domain.MsgStatusAvailable, message.Status())
	require.Equal(t, destinationQueue, message.Queue().String())
	require.Equal(t, 1, message.Generation())

	chapters, loaded := message.History().Chapters()
	require.True(t, loaded)
	require.Len(t, chapters, 1)
	require.Equal(t, msgQueue, chapters[0].Queue().String())
	require.Equal(t, 0, chapters[0].Generation())
}

func TestRedirectToUnknownQueue(t *testing.T) {
	app, _ := e2eutils.Prepare(t)

	const (
		msgQueue    = "test.result"
		msgPayload  = `{"arg": 123}`
		msgPriority = 100
	)

	// Arrange
	msgID := e2eutils.CreateProcessingMsg(t, app, msgQueue, msgPayload, msgPriority)

	// Act
	requestBody := []any{
		map[string]any{
			"id":          msgID,
			"destination": "unknown_queue",
		},
	}
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/messages/redirect", bytes.NewBuffer(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	// Assert response
	require.Equal(t, http.StatusInternalServerError, resp.Code, resp.Body.String())

	var respDTO e2eutils.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&respDTO)
	require.NoError(t, err)

	require.Contains(t, respDTO.Error, "queue not defined")
}

func TestRedirectUnknownMessage(t *testing.T) {
	app, _ := e2eutils.Prepare(t)

	// Act
	requestBody := []any{
		map[string]any{
			"id":          "d8d4d0f7-1bbd-48c0-9f80-c66f5fd45fc2",
			"destination": "all_results",
		},
	}
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/messages/redirect", bytes.NewBuffer(body))
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
