package test

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"server/internal/domain"
	"testing"
	"time"
)

func TestCheckExistingTask(t *testing.T) {
	app, _ := buildTestApp(t)
	cleanupDatabase(t, app.DB)

	// Create test data directly in the database
	const taskID = "00000000-0000-0000-0000-000000000001"
	testTask := createTask(t, app, taskID, "test", 100)

	err := app.TaskRepo.SaveInNewTransaction(context.Background(), app.DB, testTask)
	require.NoError(t, err)

	// Test the check task endpoint
	req, err := http.NewRequest(http.MethodGet, "/task/check?id="+taskID, nil)
	require.NoError(t, err)

	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	// Assert
	require.Equal(t, http.StatusOK, resp.Result().StatusCode)

	var respWrapper ResponseWrapper
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

	// Verify response fields
	require.Equal(t, taskID, respDTO.ID)
	require.Equal(t, "test", respDTO.Kind)
	require.Equal(t, app.Clock.Now(), respDTO.CreatedAt)
	require.Equal(t, string(domain.TaskStatusCreated), respDTO.Status)
	require.Equal(t, 0, respDTO.Retries)
	require.JSONEq(t, `{"arg": "cf404dc806178c245b5b4fe2531e6d8c"}`, string(respDTO.Payload))
	require.Nil(t, respDTO.Result)
}

func TestCheckNonExistentTask(t *testing.T) {
	app, _ := buildTestApp(t)
	cleanupDatabase(t, app.DB)

	const nonExistentID = "00000000-0000-0000-0000-000000000002"
	req, err := http.NewRequest(http.MethodGet, "/task/check?id="+nonExistentID, nil)
	require.NoError(t, err)

	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	// Assert
	require.Equal(t, http.StatusInternalServerError, resp.Result().StatusCode)

	var respWrapper ResponseWrapper
	err = json.NewDecoder(resp.Body).Decode(&respWrapper)
	require.NoError(t, err)

	require.False(t, respWrapper.Success)
	require.Nil(t, respWrapper.Result)
	require.NotNil(t, respWrapper.Error)
	require.Contains(t, *respWrapper.Error, "task not found")
}
