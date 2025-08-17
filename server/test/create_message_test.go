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

func TestCreateMessage(t *testing.T) {
	app, _ := e2eutils.Prepare(t)

	const (
		msgKind     = "test"
		msgPayload  = `{"arg": 123}`
		msgPriority = 100
	)

	// Act
	body, err := json.Marshal(map[string]any{
		"kind":    msgKind,
		"payload": json.RawMessage(msgPayload),
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/message/create", bytes.NewBuffer(body))
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

	var respDTO struct {
		ID string `json:"id"`
	}
	err = json.Unmarshal(*respWrapper.Result, &respDTO)
	require.NoError(t, err)

	// Assert the message in DB
	message, err := app.MsgRepo.GetByID(context.Background(), app.DB, respDTO.ID)
	require.NoError(t, err)

	require.Equal(t, msgKind, message.Kind())
	require.JSONEq(t, msgPayload, string(message.Payload()))
	require.Equal(t, app.Clock.Now(), message.CreatedAt())
	require.Equal(t, domain.MsgStatusCreated, message.Status())
	require.Equal(t, msgPriority, message.Priority())
}

func TestCreateMessageWithPriorityAndConfirm(t *testing.T) {
	app, _ := e2eutils.Prepare(t)

	const (
		msgKind     = "test"
		msgPayload  = `{"arg": 123}`
		msgPriority = 5
	)

	// Act
	body, err := json.Marshal(map[string]any{
		"kind":         msgKind,
		"payload":      json.RawMessage(msgPayload),
		"auto_confirm": true,
		"priority":     msgPriority,
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/message/create", bytes.NewBuffer(body))
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

	var respDTO struct {
		ID string `json:"id"`
	}
	err = json.Unmarshal(*respWrapper.Result, &respDTO)
	require.NoError(t, err)

	// Assert the message in DB
	message, err := app.MsgRepo.GetByID(context.Background(), app.DB, respDTO.ID)
	require.NoError(t, err)

	require.Equal(t, msgKind, message.Kind())
	require.JSONEq(t, msgPayload, string(message.Payload()))
	require.Equal(t, app.Clock.Now(), message.CreatedAt())
	require.Equal(t, domain.MsgStatusReady, message.Status())
	require.Equal(t, msgPriority, message.Priority())
}
