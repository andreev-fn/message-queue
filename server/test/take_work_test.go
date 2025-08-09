package test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"server/internal/domain"
	"server/test/e2eutils"
)

func TestTakeWork(t *testing.T) {
	app, _ := e2eutils.Prepare(t)

	const (
		task1Kind     = "test1"
		task1Payload  = `{"arg": 123}`
		task1Priority = 10

		task2Kind     = "test2"
		task2Payload  = `{"arg": 213}`
		task2Priority = 200

		task3Kind     = "test3"
		task3Payload  = `{"arg": 321}`
		task3Priority = 100
	)

	// Arrange
	task1ID := e2eutils.CreateReadyTask(t, app, task1Kind, task1Payload, task1Priority)
	task2ID := e2eutils.CreateReadyTask(t, app, task2Kind, task2Payload, task2Priority)
	task3ID := e2eutils.CreateReadyTask(t, app, task3Kind, task3Payload, task3Priority)

	// Act
	req, err := http.NewRequest(http.MethodPost, "/work/take?kind=test1,test2,test3&limit=1", nil)
	require.NoError(t, err)

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

	var results []struct {
		ID      string `json:"id"`
		Payload string `json:"payload"`
	}
	err = json.Unmarshal(*respWrapper.Result, &results)
	require.NoError(t, err)

	require.Len(t, results, 1)
	require.Equal(t, task2ID, results[0].ID)
	require.JSONEq(t, task2Payload, results[0].Payload)

	// Assert tasks in DB
	takenTask, err := app.TaskRepo.GetTaskByID(context.Background(), app.DB, task2ID)
	require.NoError(t, err)
	require.Equal(t, domain.TaskStatusProcessing, takenTask.Status())

	task1, err := app.TaskRepo.GetTaskByID(context.Background(), app.DB, task1ID)
	require.NoError(t, err)
	require.Equal(t, domain.TaskStatusReady, task1.Status())

	task3, err := app.TaskRepo.GetTaskByID(context.Background(), app.DB, task3ID)
	require.NoError(t, err)
	require.Equal(t, domain.TaskStatusReady, task3.Status())
}

func TestTakeWorkEmptyQueue(t *testing.T) {
	app, _ := e2eutils.Prepare(t)

	// Act
	req, err := http.NewRequest(http.MethodPost, "/work/take?kind=test&limit=1", nil)
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

	var results []any
	err = json.Unmarshal(*respWrapper.Result, &results)
	require.NoError(t, err)

	require.Empty(t, results, "Expected empty result list for empty queue")
}
