package test

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"server/internal/domain"
	"testing"
)

func TestTakeWork(t *testing.T) {
	app, _ := buildTestApp(t)
	cleanupDatabase(t, app.DB)

	// Create test data directly in the database
	const task1ID = "00000000-0000-0000-0000-000000000001"
	const task2ID = "00000000-0000-0000-0000-000000000002"
	const task3ID = "00000000-0000-0000-0000-000000000003"

	testTasks := []*domain.Task{
		createTask(t, app, task1ID, "test1", 10),
		createTask(t, app, task2ID, "test2", 200),
		createTask(t, app, task3ID, "test3", 100),
	}

	for _, testTask := range testTasks {
		err := testTask.Confirm(app.Clock, NoopEventDispatcher{})
		require.NoError(t, err)

		err = app.TaskRepo.SaveInNewTransaction(context.Background(), app.DB, testTask)
		require.NoError(t, err)
	}

	// Test the get task endpoint
	req, err := http.NewRequest(http.MethodPost, "/work/take?kind=test1,test2,test3&limit=1", nil)
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

	var results []struct {
		ID      string `json:"id"`
		Payload string `json:"payload"`
	}
	err = json.Unmarshal(*respWrapper.Result, &results)
	require.NoError(t, err)

	require.Len(t, results, 1)
	require.Equal(t, task2ID, results[0].ID)
	require.JSONEq(t, `{"arg": "4dc8564dbd8210652f4097f7c50f67cf"}`, results[0].Payload)

	// Verify task status was updated to PROCESSING
	updatedTask, err := app.TaskRepo.GetTaskByID(context.Background(), app.DB, task2ID)
	require.NoError(t, err)
	require.Equal(t, domain.TaskStatusProcessing, updatedTask.Status())
}

func TestTakeWorkEmptyQueue(t *testing.T) {
	app, _ := buildTestApp(t)
	cleanupDatabase(t, app.DB)

	req, err := http.NewRequest(http.MethodPost, "/work/take?kind=test&limit=1", nil)
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

	var results []any
	err = json.Unmarshal(*respWrapper.Result, &results)
	require.NoError(t, err)

	require.Empty(t, results, "Expected empty result list for empty queue")
}
