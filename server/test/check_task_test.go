package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/test/e2eutils"
)

func TestCheckExistingTask(t *testing.T) {
	app, _ := e2eutils.Prepare(t)

	const (
		taskKind     = "test"
		taskPayload  = `{"arg": 123}`
		taskPriority = 100
	)

	// Arrange
	taskID := e2eutils.CreateTask(t, app, taskKind, taskPayload, taskPriority)

	// Act
	req, err := http.NewRequest(http.MethodGet, "/task/check?id="+taskID, nil)
	require.NoError(t, err)

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

	var respDTO struct {
		ID        string           `json:"id"`
		Kind      string           `json:"kind"`
		CreatedAt time.Time        `json:"created_at"`
		Status    string           `json:"status"`
		Retries   int              `json:"retries"`
		Payload   json.RawMessage  `json:"payload"`
		Result    *json.RawMessage `json:"result"`
	}
	err = json.Unmarshal(*respWrapper.Result, &respDTO)
	require.NoError(t, err)

	require.Equal(t, taskID, respDTO.ID)
	require.Equal(t, taskKind, respDTO.Kind)
	require.Equal(t, app.Clock.Now(), respDTO.CreatedAt)
	require.Equal(t, string(domain.TaskStatusCreated), respDTO.Status)
	require.Equal(t, 0, respDTO.Retries)
	require.JSONEq(t, taskPayload, string(respDTO.Payload))
	require.Nil(t, respDTO.Result)
}

func TestCheckNonExistentTask(t *testing.T) {
	app, _ := e2eutils.Prepare(t)

	const nonExistentID = "00000000-0000-0000-0000-000000000002"

	// Act
	req, err := http.NewRequest(http.MethodGet, "/task/check?id="+nonExistentID, nil)
	require.NoError(t, err)

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
	require.Contains(t, *respWrapper.Error, "task not found")
}
