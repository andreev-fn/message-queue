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

func TestFinishSuccessfulWork(t *testing.T) {
	app, _ := e2eutils.Prepare(t)

	const (
		taskKind     = "test"
		taskPayload  = `{"arg": 123}`
		taskPriority = 100
		taskResult   = `{"result":"success"}`
	)

	// Arrange
	taskID := e2eutils.CreateProcessingTask(t, app, taskKind, taskPayload, taskPriority)

	// Act
	requestBody := map[string]interface{}{
		"id":     taskID,
		"report": taskResult,
		"error":  nil,
	}
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/work/finish", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
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

	// Assert task in DB
	task, err := app.TaskRepo.GetTaskByID(context.Background(), app.DB, taskID)
	require.NoError(t, err)
	require.Equal(t, domain.TaskStatusCompleted, task.Status())
	require.NotNil(t, task.Result())
	require.JSONEq(t, taskResult, string(*task.Result()))
}

func TestFinishUnsuccessfulWork(t *testing.T) {
	app, _ := e2eutils.Prepare(t)

	const (
		taskKind     = "test"
		taskPayload  = `{"arg": 123}`
		taskPriority = 100
	)

	// Arrange
	taskID := e2eutils.CreateProcessingTask(t, app, taskKind, taskPayload, taskPriority)

	// Act
	requestBody := map[string]interface{}{
		"id":     taskID,
		"report": nil,
		"error": map[string]interface{}{
			"code":            "timeout_error",
			"message":         "operation timed out",
			"additional_info": nil,
		},
	}
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/work/finish", bytes.NewBuffer(body))
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

	// Assert task in DB
	task, err := app.TaskRepo.GetTaskByID(context.Background(), app.DB, taskID)
	require.NoError(t, err)
	require.Equal(t, domain.TaskStatusDelayed, task.Status())
}

func TestFinishWorkUnknownTask(t *testing.T) {
	app, _ := e2eutils.Prepare(t)

	// Act
	requestBody := map[string]interface{}{
		"id":     "d8d4d0f7-1bbd-48c0-9f80-c66f5fd45fc2",
		"report": `{"result":"success"}`,
		"error":  nil,
	}
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/work/finish", bytes.NewBuffer(body))
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
	require.Contains(t, *respWrapper.Error, "task not found")
}
