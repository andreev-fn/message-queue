package test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"server/internal/domain"
	"testing"
)

func TestFinishSuccessfulWork(t *testing.T) {
	app, _ := buildTestApp(t)
	cleanupDatabase(t, app.DB)

	// Create test data directly in the database
	const taskID = "00000000-0000-0000-0000-000000000001"
	testTask := createTask(t, app, taskID, "test", 100)

	err := testTask.Confirm(app.Clock)
	require.NoError(t, err)

	err = testTask.StartProcessing(app.Clock)
	require.NoError(t, err)

	err = app.TaskRepo.SaveInNewTransaction(context.Background(), app.DB, testTask)
	require.NoError(t, err)

	// Prepare request
	requestBody := map[string]interface{}{
		"id":     taskID,
		"report": `{"result":"success"}`,
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

	var respWrapper ResponseWrapper
	err = json.NewDecoder(resp.Body).Decode(&respWrapper)
	require.NoError(t, err)

	require.True(t, respWrapper.Success)
	require.Nil(t, respWrapper.Result)
	require.Nil(t, respWrapper.Error)

	// Verify task status in database
	task, err := app.TaskRepo.GetTaskByID(context.Background(), app.DB, taskID)
	require.NoError(t, err)
	require.Equal(t, domain.TaskStatusCompleted, task.Status())
	require.JSONEq(t, `{"result":"success"}`, string(*task.Result()))
}

func TestFinishUnsuccessfulWork(t *testing.T) {
	app, _ := buildTestApp(t)
	cleanupDatabase(t, app.DB)

	const taskID = "00000000-0000-0000-0000-000000000001"
	testTask := createTask(t, app, taskID, "test", 100)

	err := testTask.Confirm(app.Clock)
	require.NoError(t, err)

	err = testTask.StartProcessing(app.Clock)
	require.NoError(t, err)

	err = app.TaskRepo.SaveInNewTransaction(context.Background(), app.DB, testTask)
	require.NoError(t, err)

	// Prepare request with error
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

	require.Equal(t, http.StatusOK, resp.Result().StatusCode)

	var respWrapper ResponseWrapper
	err = json.NewDecoder(resp.Body).Decode(&respWrapper)
	require.NoError(t, err)

	require.True(t, respWrapper.Success)
	require.Nil(t, respWrapper.Result)
	require.Nil(t, respWrapper.Error)

	// Verify task status in database
	task, err := app.TaskRepo.GetTaskByID(context.Background(), app.DB, taskID)
	require.NoError(t, err)
	require.Equal(t, domain.TaskStatusDelayed, task.Status())
}

func TestFinishWorkUnknownTask(t *testing.T) {
	app, _ := buildTestApp(t)
	cleanupDatabase(t, app.DB)

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

	require.Equal(t, http.StatusInternalServerError, resp.Result().StatusCode)

	var respWrapper ResponseWrapper
	err = json.NewDecoder(resp.Body).Decode(&respWrapper)
	require.NoError(t, err)

	require.False(t, respWrapper.Success)
	require.Nil(t, respWrapper.Result)
	require.NotNil(t, respWrapper.Error)
	require.Contains(t, *respWrapper.Error, "task not found")
}
